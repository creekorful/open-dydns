package config

import (
	"fmt"
	"github.com/creekorful/open-dydns/internal/common"
)

type Config struct {
	APIAddr string
	Token   string
}

func (c Config) Valid() bool {
	return c.APIAddr != "" && c.Token != ""
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
