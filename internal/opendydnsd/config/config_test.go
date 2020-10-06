package config

import "testing"

func TestConfig_Valid(t *testing.T) {
	c := Config{}

	if c.Valid() {
		t.Error("validate() should have failed")
	}

	c = Config{
		APIConfig: APIConfig{
			ListenAddr: "127.0.0.1:8080",
			SigningKey: "TEST",
		},
		DaemonConfig: DaemonConfig{},
		DatabaseConfig: DatabaseConfig{
			Driver: "sqlite",
			DSN:    "test.db",
		},
	}

	if !c.Valid() {
		t.Error("validate() should have work")
	}
}

func TestAPIConfig_SSLEnabled(t *testing.T) {
	c := APIConfig{}

	if c.SSLEnabled() {
		t.Error()
	}

	c.Hostname = "foo.bar.baz"
	if c.SSLEnabled() {
		t.Error()
	}

	c.CertCacheDir = "/var/lib/opendydns/certificates"
	if !c.SSLEnabled() {
		t.Error()
	}
}

func TestAPIConfig_Valid(t *testing.T) {
	c := APIConfig{}

	if c.Valid() {
		t.Error()
	}

	c.ListenAddr = "127.0.0.1:8080"
	if c.Valid() {
		t.Error()
	}

	c.SigningKey = "todo"
	if !c.Valid() {
		t.Error()
	}
}

func TestDatabaseConfig_Valid(t *testing.T) {
	c := DatabaseConfig{}

	if c.Valid() {
		t.Error()
	}

	c.Driver = "driver"
	if c.Valid() {
		t.Error()
	}

	c.DSN = "test.db"
	if !c.Valid() {
		t.Error()
	}
}

func TestDomainConfig_String(t *testing.T) {
	c := DomainConfig{Domain: "example.org"}
	if c.String() != "example.org" {
		t.Error()
	}

	c.Host = "demo"
	if c.String() != "demo.example.org" {
		t.Error()
	}
}
