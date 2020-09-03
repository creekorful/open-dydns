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
	DaemonConfig:   DaemonConfig{},
	DatabaseConfig: DatabaseConfig{},
}

type Config struct {
	ApiConfig      ApiConfig
	DaemonConfig   DaemonConfig
	DatabaseConfig DatabaseConfig
}

type ApiConfig struct {
	ListenAddr string
	SigningKey string
}

type DaemonConfig struct {
}

type DatabaseConfig struct {
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

	if !validate(config) {
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

func validate(conf Config) bool {
	// TODO complete / improve tests
	return conf.ApiConfig.ListenAddr != "" && conf.ApiConfig.SigningKey != ""
}
