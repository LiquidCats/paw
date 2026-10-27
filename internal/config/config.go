package config

import (
	"github.com/go-playground/sensitive"
	"github.com/rs/zerolog"
)

type RedisConfig struct {
	Host     string           `envconfig:"HOST"`
	Port     string           `envconfig:"PORT"`
	DB       int              `envconfig:"DB"`
	Protocol int              `envconfig:"PROTOCOL" default:"3"`
	Password sensitive.String `envconfig:"PASSWORD"`
}

type PostgresConfig struct {
	Host     string           `envconfig:"HOST"`
	Port     string           `envconfig:"PORT"`
	Database string           `envconfig:"DATABASE"`
	User     string           `envconfig:"USER"`
	Password sensitive.String `envconfig:"PASSWORD"`
	SSL      bool             `envconfig:"SSL"`
}

type ZerologConfig struct {
	Level zerolog.Level `envconfig:"LEVEL" default:"info"`
}
