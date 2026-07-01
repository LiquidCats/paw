package configs

type Config struct {
	App    AppConfig    `yaml:"app" envconfig:"APP"`
	Common CommonConfig `yaml:"common_config" envconfig:"COMMON"`
}
