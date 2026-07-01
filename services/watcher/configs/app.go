package configs

import (
	"github.com/rs/zerolog"
)

type App struct {
	LogLevel zerolog.Level `yaml:"log_level" envconfig:"LOG_LEVEL" default:"info"`
}
