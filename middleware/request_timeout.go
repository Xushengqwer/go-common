package middleware

import (
	"context"
	"errors"
	"github.com/Xushengqwer/go-common/core"
	"github.com/Xushengqwer/go-common/response"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestTimeoutMiddleware 定义请求超时中间件，为每个请求设置超时时间并在超时后中断处理
// - 使用自定义的 response.RespondError 进行标准化超时错误响应
// - 输入: logger ZapLogger 实例用于日志记录, timeout 请求超时时间
// - 输出: gin.HandlerFunc 中间件函数
func RequestTimeoutMiddleware(logger *core.ZapLogger, timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 创建带超时的上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel() // 确保在函数退出时调用 cancel 释放资源

		// 2. 替换请求上下文
		// 将 gin.Context 的原始请求上下文替换为带超时的上下文
		// 确保后续控制器或服务调用 c.Request.Context() 时能感知到超时
		c.Request = c.Request.WithContext(ctx)

		// 3. 创建完成信号通道
		// 使用大小为 1 的缓冲通道，以便非阻塞发送
		finished := make(chan struct{}, 1)

		// 4. 在协程中处理请求
		go func() {
			defer func() {
				// 在协程退出前（无论是正常完成还是 panic），
				// 尝试发送完成信号。使用非阻塞发送，以防主 select 已超时不再接收。
				select {
				case finished <- struct{}{}:
				default:
				}
				// 如果此协程发生 panic，它应该由全局的 ErrorHandlingMiddleware 捕获。
				// 此处发送信号是为了通知主 select 语句处理已结束（或被中断）。
			}()
			// 调用 c.Next() 执行链中的下一个中间件或处理函数
			c.Next()
		}()

		// 5. 等待请求完成或超时
		select {
		case <-ctx.Done(): // 监听上下文的完成事件（超时或取消）
			// 6. 处理超时或取消情况
			err := ctx.Err() // 获取导致上下文结束的错误
			// 仅在确实是超时错误时返回 504 和 ErrCodeServerTimeout
			if errors.Is(err, context.DeadlineExceeded) {
				// 记录警告日志，包含超时时长和请求信息
				logger.Warn("请求处理超时",
					zap.Duration("timeout", timeout),       // 配置的超时时长
					zap.Error(err),                         // 记录 context deadline exceeded 超时错误
					zap.String("path", c.Request.URL.Path), // 请求路径
					zap.String("method", c.Request.Method), // 请求方法
					zap.String("clientIP", c.ClientIP()),   // 客户端 IP
				)

				// - 返回标准化的超时错误响应
				response.RespondError(
					c,
					http.StatusGatewayTimeout,     // HTTP 状态码 504 (网关超时)
					response.ErrCodeServerTimeout, // 自定义服务器超时错误码 (50002)
					"请求超时，请稍后重试。",                 // 对用户友好的错误消息
				)
				// 中断请求链。a) 阻止后续 Gin 处理程序执行, b) 因为我们已写入响应，所以是必要的。
				c.Abort()
			} else {
				// 上下文因其他原因被取消（例如，客户端断开连接，或者父上下文被取消）
				logger.Info("请求上下文被取消（非超时）",
					zap.Error(err), // 记录具体的取消错误
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
				// （可选）发送不同的响应或仅中止
				// 目前，仅中止而不发送特定的响应体，因为可能是客户端主动断开
				c.Abort()
			}
			return // 从中间件函数返回

		case <-finished:
			// 7. 请求正常完成
			// 如果收到 finished 信号，表示请求在超时前完成，无需额外处理
			// 可以添加 Debug 日志确认
			logger.Debug("请求在超时前完成", zap.String("path", c.Request.URL.Path))
		}
	}
}
