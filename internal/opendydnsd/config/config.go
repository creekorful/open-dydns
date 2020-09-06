package config

import (
	"fmt"
	"github.com/creekorful/open-dydns/internal/common"
)

// DefaultConfig is the OpenDyDNSD default configuration
var DefaultConfig = Config{
	APIConfig: APIConfig{
		ListenAddr: "127.0.0.1:8888",
		SigningKey: "",
	},
	DaemonConfig: DaemonConfig{},
	DatabaseConfig: DatabaseConfig{
		Driver: "sqlite",
		DSN:    "test.db",
	},
}

// Config is the global Daemon configuration
type Config struct {
	APIConfig      APIConfig `toml:"ApiConfig"`
	DaemonConfig   DaemonConfig
	DatabaseConfig DatabaseConfig
}

// Valid determinate if config is valid one
func (c Config) Valid() bool {
	return c.APIConfig.Valid() && c.DaemonConfig.Valid() && c.DatabaseConfig.Valid()
}

// APIConfig represent the API configuration
type APIConfig struct {
	ListenAddr   string
	SigningKey   string
	CertCacheDir string
	Hostname     string
	AutoTLS      bool
}

// Valid determinate if config is valid one
func (ac APIConfig) Valid() bool {
	return ac.ListenAddr != "" && ac.SigningKey != ""
}

// SSLEnabled determinate if SSL (HTTPS) is enabled for the API
func (ac APIConfig) SSLEnabled() bool {
	return ac.CertCacheDir != ""
}

// DaemonConfig represent the daemon configuration
type DaemonConfig struct {
	DNSProvisioners []DNSProvisionerConfig `toml:"DnsProvisioners"`
}

// DNSProvisionerConfig represent the configuration of a DNS provisioner
type DNSProvisionerConfig struct {
	Name    string
	Config  map[string]string
	Domains []string
}

// Valid determinate if config is valid one
func (dc DaemonConfig) Valid() bool {
	return true
}

// DatabaseConfig represent the database configuration
type DatabaseConfig struct {
	Driver string
	DSN    string
}

// Valid determinate if config is valid one
func (dc DatabaseConfig) Valid() bool {
	return dc.Driver != "" && dc.DSN != ""
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
