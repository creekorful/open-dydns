package config

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"os"
)

var DefaultConfig = Config{
	ApiConfig: ApiConfig{
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
	ApiConfig      ApiConfig
	DaemonConfig   DaemonConfig
	DatabaseConfig DatabaseConfig
}

func (c Config) Valid() bool {
	return c.ApiConfig.Valid() && c.DaemonConfig.Valid() && c.DatabaseConfig.Valid()
}

type ApiConfig struct {
	ListenAddr string
	SigningKey string
}

func (ac ApiConfig) Valid() bool {
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
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = toml.NewDecoder(file).Decode(&config)
	if err != nil {
		return Config{}, err
	}

	if !config.Valid() {
		return Config{}, fmt.Errorf("invalid config file `%s`", path)
	}

	return config, nil
}

func Save(config Config, path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}

	return toml.NewEncoder(file).Encode(&config)
}