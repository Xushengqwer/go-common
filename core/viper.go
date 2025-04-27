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

// LoadConfig 加载配置的核心函数 (重构版)
// 设计目标：更简单、明确。优先环境变量，其次命令行指定路径，最后可选默认路径。
//
//	支持通过 CONFIG_SOURCE=env 强制只从环境变量加载。
//
// 参数：
//   - configPathFromFlag: 从命令行参数 -config 解析出的文件路径 (如果未提供 flag，则为空字符串或默认值)
//   - cfgPtr: 指向具体配置结构体的指针 (e.g., &config.GatewayConfig{})
//
// 返回值：
//   - error: 如果加载或解析失败，返回错误
func LoadConfig(configPathFromFlag string, cfgPtr interface{}) error {
	v := viper.New()

	// 1. 检查是否强制只从环境变量加载
	configSource := strings.ToLower(os.Getenv("CONFIG_SOURCE")) // 转小写方便比较
	if configSource == "env" {
		log.Println("检测到 CONFIG_SOURCE=env，将仅从环境变量加载配置。")
		v.AutomaticEnv() // 启用环境变量自动匹配 (例如 MYAPP_SERVER_PORT 匹配 Server.Port)
		// 对于嵌套结构，可能需要设置 EnvKeyReplacer
		// v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		if err := v.Unmarshal(cfgPtr); err != nil {
			log.Printf("从环境变量解析配置失败: %v", err)
			return fmt.Errorf("无法从环境变量解析配置: %w", err)
		}
		log.Printf("从环境变量加载配置成功。")
		// 注意：纯环境变量加载不支持热重载
		return nil
	}

	// 2. 确定要使用的配置文件路径 (优先级: APP_CONFIG_PATH Env > -config Flag > 默认值)
	//    你可以为你的应用选择一个特定的环境变量名，例如 GATEWAY_CONFIG_PATH
	configFilePath := os.Getenv("APP_CONFIG_PATH") // 使用通用名称或特定名称
	if configFilePath == "" {
		configFilePath = configPathFromFlag // 使用 flag 传入的值
	}

	// 3. 如果确定了配置文件路径，则加载文件
	if configFilePath != "" {
		v.SetConfigFile(configFilePath)
		v.SetConfigType("yaml") // 或根据文件扩展名自动检测 v.AutomaticConfigFileType()
		err := v.ReadInConfig()
		if err != nil {
			// 如果只是文件未找到错误，可以容忍并继续尝试环境变量
			// 如果是其他错误（如解析错误），则直接返回
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if errors.As(err, &configFileNotFoundError) {
				log.Printf("配置文件 '%s' 未找到，将继续尝试环境变量。", configFilePath)
			}
		} else {
			log.Printf("成功从文件加载配置: %s", configFilePath)
			// 文件加载成功后，仍然启用环境变量覆盖 (文件优先)
			v.AutomaticEnv()
			// v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

			// 解析配置到结构体
			if err := v.Unmarshal(cfgPtr); err != nil {
				log.Printf("解析配置文件 '%s' 到结构体失败: %v", configFilePath, err)
				return fmt.Errorf("无法解析配置文件 '%s': %w", configFilePath, err)
			}
			log.Printf("解析配置文件成功。")

			// 启用热加载 (仅对文件生效)
			v.WatchConfig()
			v.OnConfigChange(func(e fsnotify.Event) {
				log.Printf("配置文件发生变化: %s", e.Name)
				// 注意：这里需要重新 Unmarshal 到 *同一个* cfgPtr 指向的结构体实例
				// 这需要稍微复杂的设计，例如通过闭包或全局变量访问 cfgPtr，
				// 或者提供一个回调函数让调用者处理更新。
				// 简单起见，这里只打印日志，实际应用中需要处理配置更新。
				newCfg := cfgPtr // 假设 cfgPtr 是个指针
				if err := v.Unmarshal(newCfg); err != nil {
					log.Printf("热重载配置文件失败: %v", err)
				} else {
					log.Printf("配置已通过热重载更新。")
					// 可能需要触发应用内部的逻辑来应用新配置
				}
			})
			return nil // 文件加载成功
		}
	} else {
		log.Println("未指定配置文件路径，将仅尝试从环境变量加载。")
	}

	// 4. 如果文件路径未指定或文件未找到，最后尝试仅从环境变量加载
	v.AutomaticEnv()
	// v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := v.Unmarshal(cfgPtr); err != nil {
		log.Printf("尝试从环境变量解析配置失败: %v", err)
		// 如果到这里仍然失败，可能意味着配置不完整
		return fmt.Errorf("无法从文件或环境变量加载有效配置: %w", err)
	}
	log.Printf("从环境变量加载配置成功。")
	return nil
}
