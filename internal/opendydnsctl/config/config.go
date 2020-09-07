package config

import (
	"fmt"
	"github.com/creekorful/open-dydns/internal/common"
)

//go:generate mockgen -source config.go -destination=../config_mock/config_mock.go -package=config_mock

// DefaultConfig is the OpenDyDNS-CLI default configuration
var DefaultConfig = Config{
	APIAddr: "http://127.0.0.1:8888",
}

// Provider represent the way of storing read the configuration
// used to abstract & ease unit testing without having to write a real file
type Provider interface {
	Load() (Config, error)
	Save(config Config) error
}

type fileProvider struct {
	filePath string
}

func (fp *fileProvider) Load() (Config, error) {
	var config Config
	if err := common.LoadToml(fp.filePath, &config); err != nil {
		return Config{}, err
	}

	if !config.Valid() {
		return Config{}, fmt.Errorf("invalid config file `%s`", fp.filePath)
	}

	return config, nil
}

func (fp *fileProvider) Save(config Config) error {
	return common.SaveToml(fp.filePath, &config)
}

// NewFileProvider return a new config Provider using file for storage
func NewFileProvider(filepath string) Provider {
	return &fileProvider{
		filePath: filepath,
	}
}

// Config represent the OpenDyDNS-CLI configuration
type Config struct {
	APIAddr string
	Token   string
	Aliases map[string]AliasConfig
}

// AliasConfig represent the aliases part of the configuration file
type AliasConfig struct {
	Synchronize bool
}

// Valid determinate if current configuration is valid one
func (c Config) Valid() bool {
	return c.APIAddr != ""
}
