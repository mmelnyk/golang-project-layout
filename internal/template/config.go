package template

import (
	"fmt"
	"os"
)

// Config is package configuration structure
type Config struct {
	FilePath string `yaml:"path"`
}

// Validate provides config structure validation
func (conf *Config) Validate() error {
	// Demo check
	if _, err := os.Stat(conf.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("path: %w", err)
	}

	// All checks passed
	return nil
}

// Reset fills config structure with default values
func (conf *Config) Reset() {
	conf.Cleanup()
	// Demo default
	conf.FilePath = "<path/to/file>"
}

// Cleanup releases all allocated objects
func (conf *Config) Cleanup() {
	// Do nothing here
}
