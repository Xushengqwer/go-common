# go-common 公共库

[![Go Report Card](https://goreportcard.com/badge/github.com/Xushengqwer/go-common)](https://goreportcard.com/report/github.com/Xushengqwer/go-common)
[![Go Reference](https://pkg.go.dev/badge/github.com/Xushengqwer/go-common.svg)](https://pkg.go.dev/github.com/Xushengqwer/go-common)

一个共享的 Go 模块，为微服务项目提供通用的工具、中间件、配置、常量和模型。

## 概述

该库旨在为项目内的各个微服务提供标准化的构建模块，以提高一致性并减少样板代码。它提供了配置加载、结构化日志、分布式追踪、常用的 Gin 中间件、标准化的 API 响应结构以及共享的数据模型和常量。

## 安装

在其他 Go 模块中使用此库：

```bash
# 将 [[github.com/Xushengqwer/go-common@latest](https://github.com/Xushengqwer/go-common@latest)] 替换为你的实际仓库路径
go get [[github.com/Xushengqwer/go-common@latest](https://github.com/Xushengqwer/go-common@latest)] # 或特定版本 @vX.Y.Z
```

## 主要功能

以下是该库提供的关键组件概览：

### 1. 配置加载 (`core` 和 `config` 包)

* 使用 `core.LoadConfig` 函数加载配置。
* 基于 Viper 实现，支持 YAML 文件和环境变量。
* 加载优先级：`APP_CONFIG_PATH` 环境变量指定的文件 > 命令行 `-config` 参数指定的文件 > 仅环境变量 (当设置 `CONFIG_SOURCE=env` 或未找到配置文件时)。
* 支持配置文件热加载。
* 提供标准配置结构体模板 (`config` 包)，如 `ZapConfig`, `TracerConfig`, `GormLogConfig`, `ServerConfig`。服务应在其配置结构体中嵌入这些共享配置。

### 2. 结构化日志 (Zap) (`core` 和 `config` 包)

* 提供 `core.NewZapLogger` 初始化函数，返回封装好的 `*core.ZapLogger` 实例。
* 基于 Zap 实现高性能结构化日志记录。
* **K8s 友好:** 默认将 `Info`, `Warn`, `Debug` 级别日志输出到 `stdout`，将 `Error` 及以上级别日志输出到 `stderr`，方便容器日志收集。
* 提供 `core.NewGormLogger` 用于 GORM 集成，自动适配 Zap 日志。
    * 将 GORM 事件记录为结构化日志。
    * 自动在日志中添加 `trace_id` 和 `span_id` (如果存在于上下文中)。
    * 支持配置慢查询阈值 (`SlowThresholdMs`)。
    * 支持配置是否忽略 `gorm.ErrRecordNotFound` 错误 (`IgnoreRecordNotFoundError`)。

### 3. 分布式追踪 (OpenTelemetry) (`core/tracing` 和 `config` 包)

* 提供 `tracing.InitTracerProvider` 函数来初始化和注册全局 OpenTelemetry TracerProvider。
* 支持多种 Exporter (OTLP gRPC/HTTP, stdout) 和 Sampler (AlwaysOn, AlwaysOff, RatioBased)。
* **重要提示:**
    * 对具体库（如 Gin, GORM, HTTP Client, Kafka 等）的 OTel **埋点 (Instrumentation) 必须在每个服务内部单独应用**。本库只提供基础 Provider 初始化。
    * OTLP Exporter 默认使用 `WithInsecure()` 以方便测试，**生产环境必须配置 TLS 加密传输**。

### 4. Gin 中间件 (`middleware` 包)

提供用于 Gin Web 框架的通用中间件：

* `ErrorHandlingMiddleware`: 捕获 panic，记录详细错误日志（包括堆栈），并返回标准的 500 错误响应。
* `RequestLoggerMiddleware`: 记录每个请求的处理信息（方法、路径、状态码、耗时、客户端 IP、UserAgent），并包含 `trace_id` 和 `span_id`。
* `RequestTimeoutMiddleware`: 为每个请求设置超时，超时则返回 504 错误响应。
* `TraceInfoMiddleware`:  从 OTel 上下文提取 `trace_id` 和 `span_id`，并将其设置到 Gin 的上下文中，供后续中间件或处理器使用。同时可选地在响应头中添加 `X-Trace-Id`。

* **建议使用顺序:** （如果使用 otelgin）`otelgin` -> `ErrorHandlingMiddleware` -> `RequestLoggerMiddleware` -> `RequestTimeoutMiddleware` -> `TraceInfoMiddleware` -> 其他业务中间件。

### 5. API 响应 (`response` 包)

* 提供泛型 `APIResponse[T]` 结构体，用于标准化 JSON API 响应。
* 提供 `RespondSuccess` 和 `RespondError` 辅助函数，方便生成标准响应。
* 定义了一套通用的业务错误码 (`response/codes.go`)，如 `Success`, `ErrCodeClientInvalidInput`, `ErrCodeServerInternal` 等。

### 6. 数据模型与常量 (`models`, `constants`, `commonerrors` 包)

* `models/enums`: 提供共享的枚举类型，如 `UserRole`, `UserStatus`, `Platform`，并包含验证函数。
* `constants`: 定义共享常量，如上下文键名 (`RoleContextKey`) 和追踪键名 (`TraceIDKey`)。
* `commonerrors`: 定义常用的全局错误变量，如 `ErrRepoNotFound`, `ErrServiceBusy`。

## 配置项摘要

使用 `go-common` 的服务通常需要在其配置文件 (或环境变量) 中定义与以下结构体匹配的配置段：

* `logger`: 对应 `config.ZapConfig` 结构体。
* `tracing`: 对应 `config.TracerConfig` 结构体。
* `gorm_log`: (如果使用 GORM) 对应 `config.GormLogConfig` 结构体。
* `server`: (如果需要统一服务配置) 对应 `config.ServerConfig` 结构体。

*有关所需字段的详细信息，请参阅 `go-common` 库内定义这些结构体的具体 `.go` 文件。*

## 许可证

本项目采用 **[你的许可证名称, 例如: MIT]** 许可证授权 - 详情请参阅 LICENSE 文件。