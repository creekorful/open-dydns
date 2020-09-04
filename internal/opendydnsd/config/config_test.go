package config

import "testing"

func TestIsValid(t *testing.T) {
	c := Config{}

	if c.Valid() {
		t.Error("validate() should have failed")
	}

	c = Config{
		ApiConfig: APIConfig{
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
