package configs

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

func Load(prefix string) (Config, error) {
	var cfg Config

	err := envconfig.Process(prefix, &cfg)
	if err != nil {
		return Config{}, err
	}

	file, err := os.OpenFile(".app.cfg.yaml", os.O_RDONLY, 0677)
	if err != nil {
		return Config{}, err
	}

	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(&cfg); err != nil { //nolint:musttag
		return Config{}, err
	}

	return cfg, nil
}
