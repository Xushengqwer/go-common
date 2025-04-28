package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Xushengqwer/go-common/config"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

var _ logger.Interface = (*GormLogger)(nil)

// GormLogger 内部状态
type GormLogger struct {
	zapLogger                 *ZapLogger
	logLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	SkipCallerLookup          bool
	ignoreRecordNotFoundError bool // 将配置存储在 logger 实例中
}

// NewGormLogger (修改后) - 直接接收共享的 config.GormLogConfig
// zapLogger: 共享库提供的 ZapLogger 实例
// cfg: 从服务配置文件加载的共享 GormLogConfig 实例
func NewGormLogger(zapLogger *ZapLogger, cfg config.GormLogConfig) *GormLogger {
	// --- 在这里进行配置转换 ---
	gormLogLevel := logger.Info // 默认级别
	switch cfg.Level {
	case "warn":
		gormLogLevel = logger.Warn
	case "error":
		gormLogLevel = logger.Error
	case "silent":
		gormLogLevel = logger.Silent
	}

	slowThreshold := time.Duration(cfg.SlowThresholdMs) * time.Millisecond
	if slowThreshold <= 0 { // 如果配置为 0 或负数，也给个默认值
		slowThreshold = 200 * time.Millisecond // 默认 200ms
	}
	// --- 转换结束 ---

	return &GormLogger{
		zapLogger:                 zapLogger,
		logLevel:                  gormLogLevel,  // 使用转换后的级别
		SlowThreshold:             slowThreshold, // 使用转换后的阈值
		SkipCallerLookup:          cfg.SkipCallerLookup,
		ignoreRecordNotFoundError: cfg.IgnoreRecordNotFoundError, // 存储配置值
	}
}

// LogMode 设置日志级别 (保持不变)
func (g *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *g
	newLogger.logLevel = level
	return &newLogger
}

// Info (保持不变)
func (g *GormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if g.logLevel >= logger.Info {
		logFields := g.extractFields(ctx)
		g.zapLogger.Info(fmt.Sprintf(msg, args...), logFields...)
	}
}

// Warn (保持不变)
func (g *GormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if g.logLevel >= logger.Warn {
		logFields := g.extractFields(ctx)
		g.zapLogger.Warn(fmt.Sprintf(msg, args...), logFields...)
	}
}

// Error (保持不变)
func (g *GormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if g.logLevel >= logger.Error {
		logFields := g.extractFields(ctx)
		g.zapLogger.Error(fmt.Sprintf(msg, args...), logFields...)
	}
}

// Trace (修改 - 使用实例中的 ignoreRecordNotFoundError)
func (g *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	logFields := g.extractFields(ctx)
	logFields = append(logFields, zap.Duration("耗时", elapsed))
	logFields = append(logFields, zap.Int64("行数", rows))
	logFields = append(logFields, zap.String("SQL", sql))
	if !g.SkipCallerLookup {
		logFields = append(logFields, zap.String("调用者", utils.FileWithLineNum()))
	}

	switch {
	// 使用存储在 GormLogger 实例中的配置来判断是否忽略错误
	case err != nil && g.logLevel >= logger.Error && !(g.ignoreRecordNotFoundError && errors.Is(err, gorm.ErrRecordNotFound)):
		g.zapLogger.Error("SQL 执行出错", append(logFields, zap.Error(err))...)
	case elapsed > g.SlowThreshold && g.SlowThreshold != 0 && g.logLevel >= logger.Warn:
		slowLog := fmt.Sprintf("慢查询 SQL >= %v", g.SlowThreshold)
		g.zapLogger.Warn(slowLog, logFields...)
	case g.logLevel >= logger.Info:
		g.zapLogger.Info("SQL 执行", logFields...)
	}
}

// extractFields (保持不变)
func (g *GormLogger) extractFields(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0, 2)
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		fields = append(fields, zap.String("trace_id", span.SpanContext().TraceID().String()))
		fields = append(fields, zap.String("span_id", span.SpanContext().SpanID().String()))
	}
	return fields
}
