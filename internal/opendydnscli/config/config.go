package config

import (
	"fmt"
	"github.com/creekorful/open-dydns/internal/common"
)

// DefaultConfig is the OpenDyDNS-CLI default configuration
var DefaultConfig = Config{
	APIAddr: "http://127.0.0.1:8888",
}

// Config represent the OpenDyDNS-CLI configuration
type Config struct {
	APIAddr string
	Token   string
}

// Valid determinate if current configuration is valid one
func (c Config) Valid() bool {
	return c.APIAddr != ""
}

// Load load configuration from given path
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

// Save configuration in file located at path
func Save(config Config, path string) error {
	return common.SaveToml(path, &config)
}
