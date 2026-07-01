package configs

import (
	"fmt"
	"time"
)

type CommonConfig struct {
	Logging LoggingConfig `yaml:"logging" envconfig:"LOGGING"`
	GRPC    GRPCConfig    `yaml:"grpc" envconfig:"GRPC"`
	DB      DBConfig      `yaml:"db" envconfig:"DB"`
}

type GRPCConfig struct {
	Port        string        `yaml:"port" envconfig:"PORT" default:"50051"`
	ConnTimeout time.Duration `yaml:"conn_timeout" envconfig:"CONN_TIMEOUT" default:"120s"`
}

type DBConfig struct {
	Driver   string `yaml:"driver" envconfig:"DRIVER" default:"postgres"`
	Host     string `yaml:"host" envconfig:"HOST"`
	Port     string `yaml:"port" envconfig:"PORT"`
	Database string `yaml:"database" envconfig:"DATABASE"`
	User     string `yaml:"user" envconfig:"USER"`
	Password string `yaml:"password" envconfig:"PASSWORD"`
	SSL      bool   `yaml:"ssl" envconfig:"SSL" default:"false"`
}

func (d *DBConfig) ToDSN() string {
	sslmode := "disable"
	if d.SSL {
		sslmode = "enable"
	}

	return fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s?sslmode=%s",
		d.Driver,
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
		sslmode,
	)
}
