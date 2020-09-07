package config

import "testing"

func TestIsValid(t *testing.T) {
	config := Config{}

	if config.Valid() {
		t.Error("Valid() should have returned false")
	}

	if !DefaultConfig.Valid() {
		t.Error("DefaultConfig should be valid")
	}
}
