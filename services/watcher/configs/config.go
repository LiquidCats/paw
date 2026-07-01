package configs

type Config struct {
	App    App          `yaml:"app" envconfig:"APP"`
	Chains ChainsConfig `yaml:"chains"`
	DB     DBConfig     `yaml:"db" envconfig:"DB"`
	Redis  RedisConfig  `yaml:"redis" envconfig:"REDIS"`

	HTTP    HTTPConfig `yaml:"http" envconfig:"HTTP" default:"8080"`
	Metrics HTTPConfig `yaml:"metrics" envconfig:"METRICS" default:"9100"`
}
