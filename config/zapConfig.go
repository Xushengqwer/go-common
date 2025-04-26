package config

// ZapConfig 定义通用的 Zap 日志配置选项
// 这些字段是共享库期望从调用方接收的配置项
// 因为部署的主要环境是K8S，我们强制使用 stdout/stderr作为标准输出和错误输出
// 在K8s 环境下，无需设计文件日志流转，依赖 K8s 的日志收集机制（通过 stdout/stderr 输出，由 Node Agent 收集并转发到集中式系统即可）。
type ZapConfig struct {
	Level    string `mapstructure:"level" yaml:"level"`       // 日志级别 (e.g., "debug", "info", "warn", "error")
	Encoding string `mapstructure:"encoding" yaml:"encoding"` // 编码格式 ("json" or "console")
}
