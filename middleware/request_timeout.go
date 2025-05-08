package middleware

import (
	"context"
	"errors"
	"fmt"           // 引入 fmt 用于 panic 恢复时的错误格式化
	"runtime/debug" // 引入 debug 用于打印 panic 堆栈

	"github.com/Xushengqwer/go-common/core"
	"github.com/Xushengqwer/go-common/response" // 确保您的 response 包路径正确

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// skipTimeoutKey 定义用于在 Gin Context 中传递“跳过超时”标志的键名
// 您可以在自己的项目中定义这个常量，或者如果 go-common 被多个项目共享，也可以放在 go-common 的常量包中
const skipTimeoutKey = "skipTimeout"

// RequestTimeoutMiddleware (改进版)
// 功能:
// 1. 为每个请求设置超时时间。
// 2. 如果请求在超时时间内完成，则正常继续。
// 3. 如果请求超时，则安全地尝试返回 504 Gateway Timeout 响应，并中断请求处理，确保服务器不 panic。
// 4. 支持通过在 Gin Context 中设置 skipTimeoutKey=true 来跳过特定请求的超时处理。
// 5. 捕获处理请求的 goroutine 中可能发生的 panic，并将其重新抛出给上层错误处理中间件。
func RequestTimeoutMiddleware(logger *core.ZapLogger, timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- 检查是否需要跳过超时处理 ---
		// 允许上游中间件 (例如 SkipTimeoutForPaths) 通过设置 context key 来禁用此中间件
		if skip, _ := c.Get(skipTimeoutKey); skip == true {
			logger.Debug("跳过请求超时处理", zap.String("path", c.Request.URL.Path))
			c.Next() // 直接执行后续处理
			return
		}

		// --- 设置带超时的 Context ---
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		// defer cancel() 在这里是必须的，确保无论如何（正常完成、超时、panic）都能释放与 context 关联的资源
		defer cancel()

		// 将带有超时的 context 替换到请求中，以便下游处理程序可以感知到超时
		c.Request = c.Request.WithContext(ctx)

		// --- 使用通道进行同步和 Panic 处理 ---
		// finished 通道：用于通知主 goroutine 请求处理已（尝试）完成
		finished := make(chan struct{}, 1)
		// panicChan 通道：用于从请求处理 goroutine 向主 goroutine 传递 panic 信息
		panicChan := make(chan interface{}, 1)

		// --- 启动 Goroutine 处理请求 ---
		go func() {
			// 使用 defer 确保此 goroutine 退出前执行清理和信号发送
			defer func() {
				// 捕获此 goroutine 中发生的 panic
				if p := recover(); p != nil {
					// 记录详细的 panic 信息和堆栈跟踪
					err := fmt.Errorf("请求处理协程 panic: %v\n%s", p, string(debug.Stack()))
					logger.Error("请求处理协程捕获到 Panic", zap.Error(err), zap.String("path", c.Request.URL.Path))
					// 尝试将 panic 值发送到 panicChan (非阻塞)
					select {
					case panicChan <- p:
					default:
						// 如果 panicChan 因为主 goroutine 已退出而无法发送，也没关系，日志已记录
					}
				}
				// 无论是否发生 panic，都尝试通知主 goroutine 处理已结束 (非阻塞)
				select {
				case finished <- struct{}{}:
				default:
					// 如果 finished 因为主 goroutine 已退出而无法发送，也没关系
				}
			}()
			// 执行 Gin 的下一个中间件或最终的 Handler
			c.Next()
		}()

		// --- 等待结果：超时、完成 或 Panic ---
		select {
		case <-ctx.Done(): // 情况1：Context 超时或被取消
			err := ctx.Err() // 获取取消原因
			if errors.Is(err, context.DeadlineExceeded) {
				// === 处理请求超时 ===
				logger.Warn("请求处理超时",
					zap.Duration("configured_timeout", timeout), // 使用明确的字段名记录配置的超时值
					zap.Error(err), // 记录 context deadline exceeded
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("clientIP", c.ClientIP()),
				)

				// !!! 关键：在写入 504 响应前，检查响应头是否已被下游写入 !!!
				if !c.Writer.Written() {
					// 如果响应头未被写入，安全地发送 504 错误响应
					response.RespondError(
						c,
						http.StatusGatewayTimeout,
						response.ErrCodeServerTimeout, // 假设这是您定义的超时错误码
						"请求超时，请稍后重试。",
					)
					// Abort 中断后续的 Gin 处理程序（虽然可能已经没有了），
					// 并且因为我们已经写入了响应，所以必须调用它。
					c.Abort()
				} else {
					// 如果响应头已被写入（例如下游处理程序动作快），
					// 就不能再写入 504 了，否则会 panic 或产生 "superfluous" 错误。
					// 此时只记录日志，表明发生了超时但无法覆盖已发送的响应头。
					logger.Warn("请求超时，但响应头已写入，无法发送 504",
						zap.String("path", c.Request.URL.Path),
						zap.Int("actual_status", c.Writer.Status()), // 记录下游实际写入的状态码
					)
					// 仍然调用 Abort 来标记请求处理结束
					c.Abort()
				}
			} else {
				// === 处理其他 Context 取消原因 (例如客户端断开连接) ===
				logger.Info("请求上下文被取消（非超时）",
					zap.Error(err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
				// 同样检查是否已写入头，虽然可能性较小
				if !c.Writer.Written() {
					// 可以选择什么都不做，或者返回一个特定代码如 499 Client Closed Request
					// Gin 没有内置 499，所以这里仅 Abort
				}
				c.Abort() // 确保中断
			}
			// 注意：这里不需要显式 return，select 结束后函数自然结束

		case p := <-panicChan: // 情况2：处理请求的 Goroutine 发生了 Panic
			// 记录从通道接收到的 panic 值
			logger.Error("主 select 捕获到请求处理协程的 Panic", zap.Any("panic_value", p), zap.String("path", c.Request.URL.Path))
			// 调用 Abort 确保 Gin 知道请求已异常结束
			c.Abort()
			// !!! 关键：将 Panic 重新抛出 !!!
			// 这样做是为了让注册在更外层（通常是 Gin 引擎上的第一个中间件）的
			// 全局 Panic 恢复中间件 (例如您使用的 ErrorHandlingMiddleware)
			// 能够捕获这个 Panic，并按统一的策略处理（例如记录详细信息、返回 500 错误响应）。
			// 如果不重新抛出，panic 就会被这个中间件“吃掉”，全局处理程序将无法感知。
			panic(p)

		case <-finished: // 情况3：请求处理在超时前正常完成
			// 记录 Debug 日志表示请求按预期完成
			logger.Debug("请求在超时前完成", zap.String("path", c.Request.URL.Path))
			// 不需要做任何事情，主 goroutine 从 select 退出，中间件函数返回，
			// Gin 会继续执行后续的步骤（例如响应写入的收尾工作）。
		}
	}
}
