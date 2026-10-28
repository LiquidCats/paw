package config

import "github.com/LiquidCats/paw/internal/modules/rater/internal/core/domain/values"

type Config struct {
	CollectSchedule string `envconfig:"COLLECT_SCHEDULE" default:"*/5 * * * *"`

	Providers []ProviderConfig `yaml:"providers"`
}

type ProviderConfig struct {
	Enabled bool                `yaml:"enabled"`
	Name    values.ProviderName `yaml:"name"`
	URL     string              `yaml:"url"`
	Secret  *string             `yaml:"secret"`
}
