package configs

import (
	"net"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host string `yaml:"host" envconfig:"HOST"`
	Port string `yaml:"port" envconfig:"PORT"`
	DB   int    `yaml:"db" envconfig:"DB"`
}

func (r RedisConfig) ToConfig(appName string) *redis.Options {
	return &redis.Options{
		Addr:       net.JoinHostPort(r.Host, r.Port),
		ClientName: appName,
		DB:         r.DB,
	}
}
