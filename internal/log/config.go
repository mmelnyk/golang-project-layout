package log

import (
	"errors"
	"os"

	"go.melnyk.org/mlog"
	"go.melnyk.org/mlog/console"
	"go.melnyk.org/mlog/nolog"
)

// Config stuct privides configuration for logger
type Config struct {
	Provider string `yaml:"provider"`
	Level    string `yaml:"level"`
}

// Validate provides config structure validation
func (conf *Config) Validate() error {
	providers := []string{"none", "console"}
	levels := []string{"verbose", "info", "warning", "error", "fatal"}

	if !conf.check(providers, conf.Provider) {
		return errors.New("Config parameter log.provider is not set to correct value")
	}

	if !conf.check(levels, conf.Level) {
		return errors.New("Config parameter log.level is not set to correct value")
	}

	// All checks passed
	return nil
}

// Reset fills config structure with default values
func (conf *Config) Reset() {
	conf.Cleanup()
	conf.Provider = "none"
	conf.Level = "fatal"
}

// Cleanup releases all allocated objects
func (conf *Config) Cleanup() {
	// Do nothing here
}

func (conf *Config) check(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// NewLogbook returns configured logbook
func NewLogbook(conf Config) mlog.Logbook {
	res := nolog.NewLogbook()

	switch conf.Provider {
	case "console":
		res = console.NewLogbook(os.Stdout)
	}

	lv := mlog.Info
	lvm := map[string]mlog.Level{
		"verbose": mlog.Verbose,
		"info":    mlog.Info,
		"warning": mlog.Warning,
		"error":   mlog.Error,
		"fatal":   mlog.Fatal,
	}

	if l, ok := lvm[conf.Level]; ok {
		lv = l
	}

	res.SetLevel(mlog.Default, lv)

	return res
}
