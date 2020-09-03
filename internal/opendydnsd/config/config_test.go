package config

import "testing"

func TestIsValid(t *testing.T) {
	c := Config{}

	if validate(c) {
		t.Error("validate() should have failed")
	}
}