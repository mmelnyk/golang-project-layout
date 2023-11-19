package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.melnyk.org/mlog"
	"go.melnyk.org/mlog/nolog"
	"gopkg.in/yaml.v3"
)

const (
	configExt    string = ".yaml"
	configDirExt string = ".d"

	configKind string = "config"
)

var (
	errConfigNotFound     = errors.New("config not found")
	errConfigNoConfigKind = errors.New("file is not config kind")
	errConfigNoConfig     = errors.New("config section not found")
	errConfigWrongApp     = errors.New("config is not for app")
	errConfigNewer        = errors.New("config version is newer than supported")
)

var (
	log mlog.Logger
)

type config struct {
	App     string    `yaml:"app"`
	Version int       `yaml:"version"`
	Kind    string    `yaml:"kind"`
	Config  yaml.Node `yaml:"config"`
}

func defaultConfigFile() (string, error) {
	cf := filepath.Base(os.Args[0])
	bd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	log.Verbose("Binary file path:" + bd)
	n := strings.LastIndexByte(cf, '.')
	if n > 0 {
		cf = cf[:n]
	}
	cf += configExt
	log.Verbose("Expected config file name: " + cf)

	// File selection
	// 1st, check config in local dir
	if _, err := os.Stat(cf); !os.IsNotExist(err) {
		return cf, nil
	}

	// 2nd, check directory with binary
	bd = filepath.Join(bd, cf)
	if _, err := os.Stat(bd); !os.IsNotExist(err) {
		return bd, nil
	}

	// 3rd option - system config directory
	for _, v := range otherDirs {
		bd = filepath.Join(v, cf)
		if _, err := os.Stat(bd); !os.IsNotExist(err) {
			return bd, nil
		}
	}

	return cf, errConfigNotFound
}

func defaultConfigDropDir(file string) (string, error) {
	// Drop-inds directory should be located same directory as config file
	n := strings.LastIndexByte(file, '.')
	if n > 0 {
		file = file[:n]
	}
	file += configDirExt
	log.Verbose("Expected config drop-ins dir name: " + file)

	// Drop-ins directory check
	if fi, err := os.Stat(file); !os.IsNotExist(err) && fi.IsDir() {
		return file, nil
	}

	return file, errConfigNoConfig
}

func checkYaml(cont []byte) error {
	var val interface{}
	return yaml.Unmarshal(cont, &val)
}

// GetConfig fills app config (conf) structure from config file
func GetConfig(conf interface{}, app string, version int) error {
	cf, err := defaultConfigFile()

	if err != nil {
		return err
	}

	var content []byte

	if dir, err := defaultConfigDropDir(cf); err == nil {
		var files []string

		log.Info("Config drop-ins dir: " + dir)

		// Walk through drop-ins directory
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			// ... and select only config files
			if info.Mode().IsRegular() {
				if strings.HasSuffix(path, configExt) {
					files = append(files, path)
					log.Info("Drop-ins config found: " + path)
				}
			}
			return nil
		})

		if err == nil {
			// Read all drop-ins and concatenize them
			for _, file := range files {
				cont, err := os.ReadFile(file)
				if err == nil {
					if err = checkYaml(cont); err != nil {
						return fmt.Errorf("%s: %w", file, err)
					}
					content = append(content, cont...)
					content = append(content, '\n')
					if err = checkYaml(content); err != nil {
						return fmt.Errorf("%s: %w", file, err)
					}
					log.Verbose("Drop-ins config added: " + file)
				}
			}
		}
	}

	log.Info("Config file: " + cf)
	cont, err := os.ReadFile(cf)
	if err != nil {
		return err
	}
	content = append(content, cont...)

	var localcfg config

	if err = yaml.Unmarshal(content, &localcfg); err != nil {
		return err
	}

	if localcfg.App != app {
		return errConfigWrongApp
	}

	if localcfg.Kind != configKind {
		return errConfigNoConfigKind
	}

	if localcfg.Version > version {
		return errConfigNewer
	}

	if localcfg.Config.Kind == 0 {
		return errConfigNoConfig
	}

	if err = localcfg.Config.Decode(conf); err != nil {
		return err
	}

	return nil
}

// Marshal returns content of config file
func Marshal(conf interface{}, app string, version int) ([]byte, error) {
	localcfg := struct {
		App     string      `yaml:"app"`
		Version int         `yaml:"version"`
		Kind    string      `yaml:"kind"`
		Config  interface{} `yaml:"config"`
	}{App: app, Version: version, Kind: configKind, Config: conf}

	return yaml.Marshal(&localcfg)
}

// SetLogger allows to change logger
func SetLogger(joiner mlog.Joiner) {
	log = joiner.Join("cfg")
}

func init() {
	log = nolog.NewLogbook().Joiner().Join("cfg")
}
