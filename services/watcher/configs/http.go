package configs

import "time"

type HTTPConfig struct {
	Port         string        `yaml:"ports" envconfig:"PORTS"`
	ReadTimeout  time.Duration `yaml:"read_timeout" envconfig:"READ_TIMEOUT"`
	WriteTimeout time.Duration `yaml:"write_timeout" envconfig:"WRITE_TIMEOUT"`
}
