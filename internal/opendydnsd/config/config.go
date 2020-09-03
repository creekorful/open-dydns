package config

import (
	"fmt"
	"github.com/creekorful/open-dydns/internal/common"
)

var DefaultConfig = Config{
	ApiConfig: APIConfig{
		ListenAddr: "127.0.0.1:8888",
		SigningKey: "",
	},
	DaemonConfig: DaemonConfig{},
	DatabaseConfig: DatabaseConfig{
		Driver: "sqlite",
		DSN:    "test.db",
	},
}

type Config struct {
	ApiConfig      APIConfig
	DaemonConfig   DaemonConfig
	DatabaseConfig DatabaseConfig
}

func (c Config) Valid() bool {
	return c.ApiConfig.Valid() && c.DaemonConfig.Valid() && c.DatabaseConfig.Valid()
}

type APIConfig struct {
	ListenAddr string
	SigningKey string
}

func (ac APIConfig) Valid() bool {
	return ac.ListenAddr != "" && ac.SigningKey != ""
}

type DaemonConfig struct {
}

func (dc DaemonConfig) Valid() bool {
	return true
}

type DatabaseConfig struct {
	Driver string
	DSN    string
}

func (dc DatabaseConfig) Valid() bool {
	return dc.Driver != "" && dc.DSN != ""
}

func Load(path string) (Config, error) {
	var config Config
	if err := common.LoadToml(path, &config); err != nil {
		return Config{}, err
	}

	if !config.Valid() {
		return Config{}, fmt.Errorf("invalid config file `%s`", path)
	}

	return config, nil
}

func Save(config Config, path string) error {
	return common.SaveToml(path, &config)
}
