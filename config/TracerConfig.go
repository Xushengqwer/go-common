package config

// TracerConfig 定义分布式追踪的配置选项
type TracerConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled"` // 是否启用追踪
	// ServiceName 会由各个服务自己定义，不放在这里
	ExporterType     string `mapstructure:"exporter_type" yaml:"exporter_type"`         // Exporter 类型: "otlp_grpc", "otlp_http", "stdout", "jaeger" 等
	ExporterEndpoint string `mapstructure:"exporter_endpoint" yaml:"exporter_endpoint"` // Exporter 地址 (e.g., "otel-collector:4317" for grpc, "otel-collector:4318" for http)
	// ExporterTimeout // 可以添加导出超时等配置
	SamplerType  string  `mapstructure:"sampler_type" yaml:"sampler_type"`   // 采样器类型: "always_on", "always_off", "traceid_ratio", "parent_based_traceid_ratio"
	SamplerParam float64 `mapstructure:"sampler_param" yaml:"sampler_param"` // 采样器参数 (e.g., for traceid_ratio, 0.1 means 10%)
}
