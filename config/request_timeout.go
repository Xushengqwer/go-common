package config

import "time"

type RequestTimeout struct {
	RequestTimeout time.Duration `mapstructure:"request_timeout" yaml:"request_timeout"`
}
