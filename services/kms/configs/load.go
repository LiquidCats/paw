package configs

import (
	"os"

	"github.com/kelseyhightower/envconfig"
)

func Load(prefix string) (Config, error) {
	defer os.Clearenv()

	var config Config

	err := envconfig.Process(prefix, &config)

	return config, err
}
