package common

import (
	"github.com/pelletier/go-toml"
	"os"
)

// LoadToml load given TOML file and decode it into given structure
func LoadToml(path string, value interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	err = toml.NewDecoder(file).Decode(value)
	if err != nil {
		return err
	}

	return nil
}

// SaveToml save given structure in toml format into file located at given path
func SaveToml(path string, value interface{}) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}

	return toml.NewEncoder(file).Encode(value)
}
