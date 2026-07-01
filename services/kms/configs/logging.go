package configs

import "log/slog"

type LoggingConfig struct {
	Level slog.Level `yaml:"level" envconfig:"LEVEL" default:"info"`
}
