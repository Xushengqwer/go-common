# go-common 公共库

[![Go Report Card](https://goreportcard.com/badge/github.com/Xushengqwer/go-common)](https://goreportcard.com/report/github.com/Xushengqwer/go-common)
[![Go Reference](https://pkg.go.dev/badge/github.com/Xushengqwer/go-common.svg)](https://pkg.go.dev/github.com/Xushengqwer/go-common)
一个共享的 Go 模块，为微服务项目提供通用的工具、中间件、配置和模型。

## 概述

该库旨在为项目内的各个微服务提供标准化的构建模块，以提高一致性并减少样板代码。主要功能包括：

* **配置加载:** 基于 Viper 实现 (`core.LoadConfig`)，支持文件和环境变量。
* **结构化日志:** 集成 Zap 日志 (`core.NewZapLogger`) 及 GORM 支持 (`core.NewGormLogger`)。
* **分布式追踪:** 提供 OpenTelemetry 初始化 (`tracing.InitTracerProvider`)。
* **Gin 中间件:** 通用错误处理、请求日志 (集成 TraceID)、请求超时等。
* **API 响应:** 标准化的 JSON 响应结构及辅助函数 (`response` 包)。
* **数据模型:** 共享数据结构 (`models` 包)，包含 `entities` 和 `dto`。

## 安装

在其他 Go 模块中使用此库：

```bash
go get [github.com/Xushengqwer/go-common@latest](https://github.com/Xushengqwer/go-common@latest) # 或特定版本 @vX.Y.Z
````

## 主要功能

以下是该库提供的关键组件概览：

### 1\. 配置加载 (`core` 包)

提供 `core.LoadConfig` 函数，用于将配置加载到服务特定的配置结构体中。

* *服务需要定义自己的配置结构体，其中应包含共享库定义的 `core.ZapConfig` 和 `core.TracerConfig` (或类似结构)。*

### 2\. 结构化日志 (Zap) (`core` 包)

提供预配置的 `core.ZapLogger` 实例及 `core.NewZapLogger` 初始化函数。同时包含 `core.NewGormLogger` 用于 GORM 集成。

* *服务配置文件需包含一个与 `core.ZapConfig` 结构匹配的 `logger` 配置段 (假设 `ZapConfig` 定义在 `core` 包内)。*

### 3\. 分布式追踪 (OpenTelemetry) (`tracing` 包)

提供 `tracing.InitTracerProvider` 函数来初始化 OTel SDK。

* **重要提示:** 对具体库（如 Gin, GORM, HTTP Client, Kafka 等）的 OTel 埋点（Instrumentation）**必须在每个服务内部单独应用**。
* *服务配置文件需包含一个与 `tracing.TracerConfig` (或 `core.TracerConfig`，取决于实际位置) 结构匹配的 `tracing` 配置段。*

### 4\. Gin 中间件 (`middleware` 包)

提供用于 Gin 应用的通用中间件 (`ErrorHandlingMiddleware`, `RequestLoggerMiddleware`, `RequestTimeoutMiddleware` 等)。

* 使用 `router.Use()` 应用。
* **顺序建议:** OTel Gin 中间件 (如果使用) -\> `ErrorHandlingMiddleware` -\> `RequestLoggerMiddleware` -\> `RequestTimeoutMiddleware` -\> 其他。
* `RequestLoggerMiddleware` 已适配为包含 OTel TraceID 和 SpanID。

### 5\. API 响应辅助函数 (`response` 包)

使用 `response` 包 (`RespondSuccess`, `RespondError`, `APIResponse`, 预定义错误码) 来标准化 API 的 JSON 响应。

### 6\. 数据模型 (`models` 包)

在 `models` 包下定义共享数据结构，推荐组织方式：

* `models/entities`: 核心领域/数据库实体 (如 GORM 模型)。
* `models/dto`: 数据传输对象 (用于 API, Kafka 消息等)。

*具体模型定义请参考包内文件。*

## 配置项摘要

使用 `go-common` 的服务通常需要在其配置文件中定义以下配置段，其结构应与本共享库中定义的相应结构体匹配：

* `logger`: 对应 `core.ZapConfig` 结构体 (假设定义在 `core` 包内)。
* `tracing`: 对应 `tracing.TracerConfig` 或 `core.TracerConfig` 结构体 (根据其在你项目中的实际位置而定)。

*有关所需字段的详细信息，请参阅 `go-common` 库内定义这些结构体的具体 `.go` 文件 (例如 `core/zap_config.go` 或 `tracing/config.go` 等)。*

## 许可证

本项目采用 **[你的许可证名称, 例如: MIT]** 许可证授权 - 详情请参阅 [https://www.google.com/search?q=LICENSE](https://www.google.com/search?q=LICENSE) 文件。