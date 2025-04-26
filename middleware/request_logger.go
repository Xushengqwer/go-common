package middleware

import (
	"github.com/Xushengqwer/go-common/constants"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace" // 导入 OTel trace
	"go.uber.org/zap"
)

// RequestLoggerMiddleware (已改造) - 记录请求摘要日志，包含 OTel TraceID 和 SpanID
// logger 参数仍然需要，用于实际记录日志
// 移除了 isGateway 参数和逻辑，使其更通用
func RequestLoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 记录请求开始时间
		startTime := time.Now()

		// 2. 处理后续请求
		c.Next() // 执行后续中间件和 Handler

		// 3. 计算总处理时长
		endTime := time.Now()
		totalLatency := endTime.Sub(startTime)

		// 4. 从上下文中获取请求信息
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 5. 从 Gin Context 中获取 TraceID 和 SpanID
		// 假设 RequestIDMiddleware (或类似机制) 已经将它们放入 Context
		traceIDVal, tidExists := c.Get(constants.TraceIDKey)
		spanIDVal, sidExists := c.Get(constants.SpanIDKey)

		traceID, _ := traceIDVal.(string)
		spanID, _ := spanIDVal.(string)

		// 如果 Context 中没有 (例如 RequestIDMiddleware 未运行)，尝试直接从 OTel Context 获取一次
		if !tidExists || !sidExists || traceID == "" || spanID == "" {
			span := trace.SpanFromContext(c.Request.Context())
			sCtx := span.SpanContext()
			if sCtx.IsValid() {
				traceID = sCtx.TraceID().String()
				spanID = sCtx.SpanID().String()
			} else {
				// 确保有默认值，避免日志字段缺失
				if traceID == "" {
					traceID = "unknown-trace-id"
				}
				if spanID == "" {
					spanID = "unknown-span-id"
				}
			}
		}

		// 6. 构建日志字段 (包含 trace_id 和 span_id)
		logFields := []zap.Field{
			zap.String("trace_id", traceID),              // OTel Trace ID
			zap.String("span_id", spanID),                // OTel Span ID
			zap.String("http.method", method),            // 请求方法 (遵循 OTel 语义约定更好)
			zap.String("url.path", path),                 // 请求路径 (遵循 OTel 语义约定更好)
			zap.Int("http.status_code", statusCode),      // 响应状态码 (遵循 OTel 语义约定更好)
			zap.Duration("duration", totalLatency),       // 总请求处理时长 (ms 或 s)
			zap.String("client.address", clientIP),       // 客户端 IP (遵循 OTel 语义约定更好)
			zap.String("user_agent.original", userAgent), // 用户代理
			// 可以添加其他有用的字段，如 request body size, response body size 等
		}

		// 7. 记录请求日志 (通常是 INFO 级别)
		logger.Info("HTTP request processed", logFields...)
	}
}
