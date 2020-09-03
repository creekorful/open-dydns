package config

import "testing"

func TestIsValid(t *testing.T) {
	c := Config{}

	if c.Valid() {
		t.Error("validate() should have failed")
	}
}