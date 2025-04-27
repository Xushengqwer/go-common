package config

import "time"

type ServerConfig struct {
	ListenAddr     string        `mapstructure:"listen_addr" yaml:"listen_addr"` // 添加监听地址
	Port           string        `mapstructure:"port" yaml:"port"`
	RequestTimeout time.Duration `mapstructure:"requestTimeout" yaml:"requestTimeout"`
}
