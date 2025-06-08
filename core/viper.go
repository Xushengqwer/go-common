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

func LoadConfig(configPathFromFlag string, cfgPtr interface{}) error {
	v := viper.New()

	// 1. 统一设置Viper，无论配置来源如何
	//    这是让环境变量能正确覆盖嵌套字段的关键！
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 2. 确定要使用的配置文件路径
	//    优先级: APP_CONFIG_PATH 环境变量 > -config 命令行标志
	configFilePath := os.Getenv("APP_CONFIG_PATH")
	if configFilePath == "" {
		configFilePath = configPathFromFlag
	}

	// 3. 如果有配置文件路径，就尝试加载它
	if configFilePath != "" {
		v.SetConfigFile(configFilePath)
		v.SetConfigType("yaml")
		if err := v.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if errors.As(err, &configFileNotFoundError) {
				// 文件没找到是正常情况，程序可以继续，仅从环境变量加载
				log.Printf("信息: 配置文件 '%s' 未找到，将仅从环境变量加载配置。", configFilePath)
			} else {
				// 如果是其他错误（如YAML格式错误），则是一个严重问题
				return fmt.Errorf("无法读取或解析配置文件 '%s': %w", configFilePath, err)
			}
		} else {
			log.Printf("成功从文件加载配置: %s。环境变量（如果存在）将覆盖文件中的值。", configFilePath)
		}
	} else {
		log.Println("信息: 未提供配置文件路径，将仅从环境变量加载配置。")
	}

	// 4. 将所有配置（已合并文件和环境变量）解析到结构体中
	if err := v.Unmarshal(cfgPtr); err != nil {
		return fmt.Errorf("无法将最终配置解析到结构体: %w", err)
	}

	log.Println("配置加载和解析成功。")

	// 5. 处理热加载
	if v.ConfigFileUsed() != "" {
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			log.Printf("配置文件发生变化: %s。尝试热重载...", e.Name)
			if err := v.Unmarshal(cfgPtr); err != nil {
				log.Printf("热重载配置文件失败: %v", err)
			} else {
				log.Printf("配置已通过热重载更新。")
			}
		})
	}

	return nil
}
