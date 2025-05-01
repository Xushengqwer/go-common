package tracing

import (
	"context"
	"errors"
	"fmt"
	"github.com/Xushengqwer/go-common/config"
	"time"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	// 可能需要导入 jaeger exporter 等
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0" // 使用最新的语义约定
)

// InitTracerProvider 初始化并注册全局的 OpenTelemetry TracerProvider
// -- serviceName: 当前服务的名称 (重要!)
// -- serviceVersion: 当前服务的版本 (可选)
// -- cfg: 从服务配置中加载的 TracerConfig
// 返回值: shutdown 函数用于优雅关闭，以及可能出现的错误
func InitTracerProvider(serviceName, serviceVersion string, cfg config.TracerConfig) (func(context.Context) error, error) {
	if !cfg.Enabled {
		fmt.Println("分布式追踪已禁用.")
		// 返回一个无操作的 shutdown 函数和 nil 错误
		return func(context.Context) error { return nil }, nil
	}

	// 创建一个背景上下文，用于初始化过程
	ctx := context.Background()

	// 1. 创建 Exporter (根据配置选择)
	var exporter sdktrace.SpanExporter
	var err error
	switch cfg.ExporterType {
	case "otlp_grpc":
		// 注意: 生产环境通常需要配置 TLS, headers (API Key) 等
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint),
			//  todo 注意：生产环境通常需要加密传输 (TLS)！这里 WithInsecure 是为了方便测试不加密
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithTimeout(5 * time.Second),
		}
		exporter, err = otlptracegrpc.New(ctx, opts...)
	case "otlp_http":
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.ExporterEndpoint),
			otlptracehttp.WithInsecure(), // 生产环境应移除或配置 TLS
			// otlptracehttp.WithURLPath("/v1/traces"), // 有些 OTLP 接收端需要特定的 URL 路径
		}
		exporter, err = otlptracehttp.New(ctx, opts...)
	case "stdout":
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	// case "jaeger":
	// ... Jaeger exporter setup ...
	default:
		err = fmt.Errorf("不支持的 exporter 类型: %s", cfg.ExporterType)
	}
	if err != nil {
		return nil, fmt.Errorf("创建 %s exporter 失败: %w", cfg.ExporterType, err)
	}

	// 2. 创建 Sampler (根据配置选择)
	var sampler sdktrace.Sampler
	switch cfg.SamplerType {
	case "always_on":
		sampler = sdktrace.AlwaysSample()
	case "always_off":
		sampler = sdktrace.NeverSample()
	case "traceid_ratio":
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplerParam) // cfg.SamplerParam 是采样率，如 0.1
	case "parent_based_traceid_ratio":
		// 如果父 Span 被采样，则子 Span 也采样；否则根据比例采样（推荐用于微服务）
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SamplerParam))
	default:
		fmt.Printf("不支持的 sampler 类型: %s，将使用 AlwaysSample\n", cfg.SamplerType)
		sampler = sdktrace.AlwaysSample()
	}

	// 3. 创建 Resource (包含服务名、版本等通用属性)
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			// 可以添加环境 (prod/dev), k8s pod name 等属性
			// attribute.String("environment", environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("创建 resource 失败: %w", err)
	}

	// 4. 创建 TracerProvider
	// 使用 BatchSpanProcessor 提高性能，异步批量导出 Span
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// 5. 注册为全局 Provider 和 Propagator
	otel.SetTracerProvider(tp)
	// 使用 W3C Trace Context (标准) 和 Baggage 进行上下文传播
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // 必须
		propagation.Baggage{},      // 可选，用于传递业务自定义数据
	))

	fmt.Printf("分布式追踪初始化完成: ServiceName=%s, Exporter=%s, Sampler=%s(%.2f)\n",
		serviceName, cfg.ExporterType, cfg.SamplerType, cfg.SamplerParam)

	// 返回 shutdown 函数，用于程序退出时优雅地 flush 数据
	shutdown := func(ctx context.Context) error {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second) // 设置超时
		defer cancel()
		err := tp.Shutdown(shutdownCtx)
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("关闭 TracerProvider 失败: %w", err)
		}
		return nil
	}

	return shutdown, nil
}
