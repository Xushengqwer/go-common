package core

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// LoadConfig 加载配置的核心函数 (通用版本)
// cfgPtr: 必须是一个指向你的具体配置结构体的指针 (e.g., &config.PostConfig{})
func LoadConfig(configPath string, cfgPtr interface{}) error {
	v := viper.New()

	// 尝试从文件加载
	if configPath != "" {
		v.SetConfigFile(configPath)
		v.SetConfigType("yaml") // 或者根据需要支持更多类型
		if err := v.ReadInConfig(); err == nil {
			log.Printf("成功从文件加载配置: %s", configPath)
			// 使用传入的指针来解析
			if err := v.Unmarshal(cfgPtr); err != nil {
				log.Printf("解析配置文件失败: %v", err)
				return fmt.Errorf("无法解析配置: %v", err)
			}
			log.Printf("配置解析成功: %+v", cfgPtr)

			// 热加载 (注意: 热加载时需要更新传入的指针指向的结构体)
			v.WatchConfig()
			v.OnConfigChange(func(e fsnotify.Event) {
				log.Printf("配置文件发生变化: %s", e.Name)
				// 重新解析到传入的指针
				if err := v.Unmarshal(cfgPtr); err != nil {
					log.Printf("重新解析配置文件失败: %v", err)
				} else {
					log.Printf("配置已更新: %+v", cfgPtr)
					// 这里可能需要通知应用的其他部分配置已更新
				}
			})
			return nil // 文件加载成功
		} else {
			log.Printf("从文件加载配置失败: %s, 错误: %v", configPath, err)
			// 文件加载失败不一定是致命错误，可能回退到环境变量
		}
	}

	// 回退到环境变量
	v.AutomaticEnv()
	log.Println("尝试从环境变量加载配置")
	if err := v.Unmarshal(cfgPtr); err != nil {
		log.Printf("从环境变量解析配置失败: %v", err)
		return fmt.Errorf("无法从环境变量解析配置: %v", err)
	}
	log.Printf("环境变量配置解析成功: %+v", cfgPtr)
	return nil
}

// 注意: 原来的 unmarshalConfig 函数现在合并到 LoadConfig 中了
// 注意: ZapConfig 结构体的定义也应该放在共享库的某个地方，比如 core/logger.go 或一个新的 core/config_models.go
