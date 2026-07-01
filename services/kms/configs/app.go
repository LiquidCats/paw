package configs

import "github.com/LiquidCats/paw/services/litehsm/pkg/configuration"

type AppConfig struct {
	KeyManager KeyManagerConfig `yaml:"key_manager"`
}

type KeyManagerConfig struct {
	Seed KeyManagerSeedConfig `yaml:"seed"`
}

type KeyManagerSeedConfig struct {
	Passphrase configuration.SealedParam `yaml:"passphrase_source"`
	Sealing    configuration.LockedParam `yaml:"sealing"`
}
