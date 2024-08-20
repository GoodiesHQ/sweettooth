package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/goodieshq/sweettooth/pkg/util"
	"github.com/rs/zerolog"
	yaml "gopkg.in/yaml.v3"
)

var (
	CONFIG_DEFAULT_LOGLEVEL = zerolog.LevelInfoValue
)

type Config struct {
	Server struct {
		Url      string `yaml:"url"`
		Insecure bool   `yaml:"insecure,omitempty"`
	} `yaml:"server"`
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
}

func (cfg *Config) Save(filename string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return err
	}

	return nil
}

func LoadConfig(filename string) (*Config, error) {
	var conf Config

	// requires a configuration
	if !util.IsFile(filename) {
		return nil, fmt.Errorf("client configuration file '%s' does not exist", filename)
	}

	// open the yaml config file
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	if err := yaml.NewDecoder(f).Decode(&conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func loadConf() (*Config, error) {
	fname := config.ClientConfig()

	// if there is already a config file, load it and return
	if util.IsFile(fname) {
		return LoadConfig(fname)
	}

	// if there is no config file, extract it from the environment variable
	url := os.Getenv("SWEETTOOTH_SERVER_URL")

	// if the environment variable is not empty...
	if url != "" {
		// ... then create a default config which will be used on future executions
		var cfg Config
		cfg.Server.Url = url
		cfg.Logging.Level = CONFIG_DEFAULT_LOGLEVEL
		if err := cfg.Save(fname); err != nil {
			return nil, err
		}
		return &cfg, nil
	}

	return nil, errors.New("server URL was not found in configuration file or environment variable")
}
