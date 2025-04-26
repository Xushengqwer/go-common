package middleware

import (
	"fmt"
	"github.com/Xushengqwer/go-common/core"
	"github.com/Xushengqwer/go-common/response"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandlingMiddleware 定义 Gin 的全局错误处理中间件，用于捕获和处理 Panic
// - 使用自定义的 response.RespondError 进行标准化错误响应
// - 输入: logger ZapLogger 实例，用于记录错误日志
// - 输出: gin.HandlerFunc 中间件函数
func ErrorHandlingMiddleware(logger *core.ZapLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 设置 Panic 捕获逻辑
		defer func() {
			if err := recover(); err != nil {
				// 2. 记录 Panic 错误日志
				logger.Error("Panic recovered",
					zap.Any("error", err),                           // 错误内容
					zap.String("errorType", fmt.Sprintf("%T", err)), // 错误类型
					zap.String("stack", string(debug.Stack())),      // 堆栈跟踪
					zap.String("path", c.Request.URL.Path),          // 请求路径
					zap.String("method", c.Request.Method),          // 请求方法
					zap.String("clientIP", c.ClientIP()),            // 客户端 IP
				)

				// 3. 返回标准化的错误响应
				// - 使用自定义的 RespondError 函数
				response.RespondError(
					c,                              // Gin 上下文
					http.StatusInternalServerError, // HTTP 状态码 500
					response.ErrCodeServerInternal, // 自定义内部服务器错误码
					"服务器内部错误，请稍后再试。", // 对用户友好的错误消息
				)
				// c.Abort() 确保在此中间件之后不再调用其他处理程序
				// RespondError 已经写入了响应，但 Abort 明确停止链式调用
				c.Abort()
			}
		}()

		// 4. 继续处理后续请求
		c.Next()
	}
}
