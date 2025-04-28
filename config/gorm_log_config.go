package config

// GormLogConfig 定义 GORM 日志记录器的通用配置项
type GormLogConfig struct {
	Level                     string `mapstructure:"level" yaml:"level"`                                         // GORM 日志级别字符串 (e.g., "info", "warn", "error", "silent")
	SlowThresholdMs           int    `mapstructure:"slowThresholdMs" yaml:"slowThresholdMs"`                     // 慢查询阈值 (毫秒)
	SkipCallerLookup          bool   `mapstructure:"skipCallerLookup" yaml:"skipCallerLookup"`                   // 是否跳过 GORM 的调用者信息查找 (提升性能)
	IgnoreRecordNotFoundError bool   `mapstructure:"ignoreRecordNotFoundError" yaml:"ignoreRecordNotFoundError"` // 是否忽略 'record not found' 错误 (通常为 true)
}
