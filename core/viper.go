// package core a
// in github.com/Xushengqwer/go-common/core

package core

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// LoadConfig 是一个健壮的、支持多源的配置加载函数。
// 它的设计目标是为所有微服务提供一个统一、灵活且可预测的配置解决方案。
//
// 工作流程:
//  1. 设置环境变量的读取规则，这是最高优先级的配置源。
//  2. 确定配置文件的路径，优先级为：环境变量 APP_CONFIG_PATH > 命令行 -config 标志 > 硬编码默认值。
//  3. 如果配置文件存在，则加载它作为配置的“基础”或“默认值”。
//  4. 将所有配置源（文件 + 环境变量）合并，并最终解析（Unmarshal）到传入的结构体指针中。
//  5. （可选）启动对配置文件的热加载监控。
//
// 参数:
//   - configPathFromFlag: 从命令行 -config 标志接收到的配置文件路径。
//   - cfgPtr: 需要被填充配置的目标结构体的指针 (e.g., &config.AppConfig)。
//
// 返回:
//   - error: 如果在加载或解析过程中发生不可恢复的错误，则返回错误。
func LoadConfig(configPathFromFlag string, cfgPtr interface{}) error {
	// 初始化一个新的 Viper 实例，避免使用全局单例，以保证配置的隔离性。
	v := viper.New()

	// --- 步骤 1: 设定环境变量的读取规则 ---
	// 这是实现灵活部署和安全性的关键。

	// `AutomaticEnv` 会让 Viper 自动查找与配置键匹配的环境变量。
	v.AutomaticEnv()

	// `SetEnvKeyReplacer` 定义了从配置键名到环境变量名的映射规则。
	// 例如，对于结构体中的 `AliyunAudit.AccessKeyID` 字段，
	// Viper 会将其转换为 `aliyun_audit.access_key_id`，
	// 然后通过此替换器，最终去寻找名为 `ALIYUNAUDIT_ACCESS_KEY_ID` 的环境变量。
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// --- 步骤 2: 确定最终要使用的配置文件路径 ---
	// 优先级: 环境变量 > 命令行标志。
	// 这使得在 CI/CD 或容器环境中可以通过环境变量轻松覆盖默认路径。
	configFilePath := os.Getenv("APP_CONFIG_PATH")
	if configFilePath == "" {
		configFilePath = configPathFromFlag // 如果环境变量不存在，则使用从命令行标志传入的路径。
	}

	// --- 步骤 3: 加载配置文件（如果路径有效）---
	// 配置文件作为一套“默认”或“基础”配置。
	if configFilePath != "" {
		v.SetConfigFile(configFilePath) // 告诉 Viper 配置文件的完整路径。
		v.SetConfigType("yaml")         // 明确指定配置文件类型。

		// `ReadInConfig` 尝试读取并解析该文件。
		if err := v.ReadInConfig(); err != nil {
			// 使用 `errors.As` 来精确判断错误类型。
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if errors.As(err, &configFileNotFoundError) {
				// 如果错误是“文件未找到”，这是一个可接受的场景。
				// 这意味着服务可以在完全没有配置文件、仅靠环境变量的情况下运行。
				log.Printf("信息: 配置文件 '%s' 未找到，将仅从环境变量加载配置。", configFilePath)
			} else {
				// 如果是其他类型的错误，比如 YAML 格式错误，这是一个严重问题，必须立即失败。
				return fmt.Errorf("无法读取或解析配置文件 '%s': %w", configFilePath, err)
			}
		} else {
			log.Printf("成功从文件加载配置: %s。环境变量(如果存在)将覆盖文件中的值。", configFilePath)
		}
	} else {
		// 如果自始至终都没有提供任何配置文件路径。
		log.Println("信息: 未提供配置文件路径，将仅从环境变量加载配置。")
	}

	// --- 步骤 4: 将所有配置源合并并解析到结构体中 ---
	// 这是最关键的一步。`Unmarshal` 会智能地合并所有来源的配置，
	// 环境变量的优先级高于配置文件，所以在这里会完成最终的覆盖。
	if err := v.Unmarshal(cfgPtr); err != nil {
		return fmt.Errorf("无法将最终配置解析到结构体: %w", err)
	}

	log.Println("配置加载和解析成功。")

	// --- 步骤 5 (可选): 启用配置文件热加载 ---
	// `WatchConfig` 会监控配置文件的变化。
	// 注意：在无状态的、不可变的容器化部署中，这个功能通常不被使用。
	// 但对于传统的、长期运行的虚拟机部署，它非常有用。
	if v.ConfigFileUsed() != "" { // 仅当确实加载了文件时才监控
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			log.Printf("配置文件发生变化: %s。尝试热重载...", e.Name)
			// 当文件变化时，再次执行 Unmarshal，将新的配置加载到原始的结构体指针中。
			if err := v.Unmarshal(cfgPtr); err != nil {
				log.Printf("热重载配置文件失败: %v", err)
			} else {
				log.Printf("配置已通过热重载更新。")
			}
		})
	}

	return nil
}
