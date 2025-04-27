# go-common 公共库

[![Go Report Card](https://goreportcard.com/badge/github.com/Xushengqwer/go-common)](https://goreportcard.com/report/github.com/Xushengqwer/go-common)
[![Go Reference](https://pkg.go.dev/badge/github.com/Xushengqwer/go-common.svg)](https://pkg.go.dev/github.com/Xushengqwer/go-common)
一个共享的 Go 模块，为微服务提供通用的工具、中间件、配置和模型。

## 概述

该库旨在为项目内的各个微服务提供标准化的构建模块，以提高一致性并减少样板代码。主要功能包括：

* **配置加载:** 基于 Viper 实现，支持从文件和环境变量加载配置。
* **结构化日志:** 集成了标准配置的 Zap 日志包装器，并支持 GORM 日志记录。
* **分布式追踪:** 提供 OpenTelemetry 初始化辅助函数。
* **Gin 中间件:** 包含错误处理、请求日志（集成追踪 ID）、请求超时等通用中间件。
* **API 响应:** 标准化的 JSON 响应结构体及辅助函数。
* **数据模型:** 共享数据结构（实体、DTO）的中心位置。

## 安装

在其他 Go 模块（例如某个微服务）中使用此库：

```bash
go get [github.com/Xushengqwer/go-common@latest](https://github.com/Xushengqwer/go-common@latest) # 或者使用特定的版本标签，如 @v0.1.0
````

## 功能梳理

以下是使用该库提供的关键组件的示例。

### 1\. 配置加载

`core.LoadConfig` 函数提供了一种将配置加载到你的服务特定配置结构体中的通用方法。

**服务中所需的配置结构体:** 你的服务需要定义自己的配置结构体（例如 `ServiceConfig`）。



### 2\. 结构化日志 (Zap)

提供一个预先配置好的 Zap 日志记录器实例。

**所需配置:** 你的服务配置文件需要包含一个与 `config.ZapConfig` 结构匹配的 `logger` 部分。


### 3\. 分布式追踪 (OpenTelemetry)

提供初始化 OpenTelemetry SDK 的辅助函数。**注意:** 对具体库（如 Gin, GORM, HTTP Client, Kafka 等）的 OTel 埋点（Instrumentation）**必须在每个服务内部单独应用**。

**所需配置:** 你的服务配置文件需要包含一个与 `config.TracerConfig` 结构匹配的 `tracing` 部分。


### 4\. Gin 中间件

提供用于 Gin 应用的通用中间件。使用 `router.Use()` 来应用它们。**重要提示:** 如果你使用了 OTel 追踪，请确保 OTel 的 Gin 中间件（例如 `otelgin.Middleware()`）在依赖追踪上下文的中间件（如改造后的 `RequestLoggerMiddleware`）**之前**运行。


### 5\. API 响应辅助函数

使用 `response` 包来标准化 API 的 JSON 响应。


### 6\. 数据模型 (Models)

该库在 `models` 包下定义共享的数据结构，通常组织如下：

* `models/entities`: 存放核心的领域实体或数据库实体（例如，GORM 模型，如 `Post`）。
* `models/dto`: 存放数据传输对象（Data Transfer Objects），用于 API 请求/响应、Kafka 消息体等（例如，`CreatePostRequest`, `PostKafkaEvent`）。

请参考 `models` 目录下的具体文件以了解可用的数据结构。

## 配置项摘要

使用 `go-common` 的服务通常需要在其配置文件（或环境变量）中定义以下配置段，其结构应与 `go-common/config` 中的结构体匹配：

* `logger`: 对应 `core.ZapConfig` 结构体。
* `tracing`: 对应 `core.TracerConfig` 结构体。

有关所需字段的详细信息，请参阅 `config/zap_config.go` 和 `config/tracing_config.go` 文件。


