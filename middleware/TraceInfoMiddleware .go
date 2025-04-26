package middleware

import (
	"github.com/Xushengqwer/go-common/constants"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace" // 导入 OTel trace 包
)

// RequestIDMiddleware (已改造) - 主要职责是提取 OTel Trace/Span ID 并放入 Context
// 不再生成 ID，也不处理 X-Request-Id Header 的传入（依赖 otelgin）
// logger 参数不再需要，因为不在此处记录日志
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 OTel Context 中获取当前的 Span Context
		// otelgin 中间件应该已经处理了传入请求的 traceparent Header
		// 并将有效的 Span 放入了 c.Request.Context()
		span := trace.SpanFromContext(c.Request.Context())
		sCtx := span.SpanContext()

		var traceID, spanID string

		// 2. 检查 Span Context 是否有效
		if sCtx.IsValid() {
			traceID = sCtx.TraceID().String()
			spanID = sCtx.SpanID().String()
		} else {
			// 如果没有有效的 Span Context (可能 otelgin 未运行或未提取到)，
			// 可以选择设置一个默认值或空字符串
			traceID = "no-trace-id"
			spanID = "no-span-id"
		}

		// 3. 将 TraceID 和 SpanID 存入 gin.Context，方便后续使用
		// 使用共享库中定义的常量作为 Key
		c.Set(constants.TraceIDKey, traceID)
		c.Set(constants.SpanIDKey, spanID)

		// 4. [可选] 将 TraceID 设置到响应头 X-Trace-Id
		// OTel 标准不要求，但有时方便客户端或 Nginx 等关联日志
		// 注意 Header 名称可以自定义或标准化
		if traceID != "no-trace-id" {
			c.Header("X-Trace-Id", traceID)
		}

		// 5. 继续处理请求
		c.Next()
	}
}
