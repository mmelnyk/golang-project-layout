package main

import (
	"fmt"

	"github.com/mmelnyk/golang-project-layout/internal/config"
	"github.com/mmelnyk/golang-project-layout/internal/log"
	"github.com/mmelnyk/golang-project-layout/internal/template"
)

const (
	currentConfigVersion = 1
)

type appconfig struct {
	Log log.Config      `yaml:"log"`
	Lib template.Config `yaml:"lib"`
}

func getConfig() (*appconfig, error) {
	cfg := &appconfig{}

	err := config.GetConfig(cfg, appshortname, currentConfigVersion)
	if err != nil {
		return cfg, err
	}

	// Do all required configuration stuff

	return cfg, err
}

func (cfg *appconfig) check() error {
	// Do config check here
	if err := cfg.Log.Validate(); err != nil {
		return fmt.Errorf("config:log:%w", err)
	}

	if err := cfg.Lib.Validate(); err != nil {
		return fmt.Errorf("config:lib:%w", err)
	}

	return nil
}

func (cfg *appconfig) cleanup() {
	// Do config cleanup here
	cfg.Log.Cleanup()
	cfg.Lib.Cleanup()
}
