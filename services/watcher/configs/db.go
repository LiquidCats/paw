package configs

import (
	"fmt"
)

type DBConfig struct {
	Driver   string `yaml:"driver" envconfig:"DRIVER" default:"postgres"`
	Host     string `yaml:"host" envconfig:"HOST"`
	Port     string `yaml:"port" envconfig:"PORT"`
	Database string `yaml:"database" envconfig:"DATABASE"`
	User     string `yaml:"user" envconfig:"USER"`
	Password string `yaml:"password" envconfig:"PASSWORD"`
}

func (d *DBConfig) ToDSN() string {
	return fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s?sslmode=disable",
		d.Driver,
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
	)
}
