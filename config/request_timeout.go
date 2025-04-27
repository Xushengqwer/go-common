package config

import "time"

type RequestTimeout struct {
	requestTimeout time.Duration `mapstructure:"request_timeout" yaml:"request_timeout"`
}
