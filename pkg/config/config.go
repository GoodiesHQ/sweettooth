package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/info"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var (
	DEFAULT_LOGLEVEL string = zerolog.LevelInfoValue
)

// create the base directory to store configs and cache
func Bootstrap(force bool) error {
	log := util.Logger("config::Bootstrap")
	log.Trace().Msg("called")
	defer log.Trace().Msg("finish")

	var err error
	for _, directory := range []string{dirBase(), dirLogs(), dirKeys(), dirBin()} {
		err = os.MkdirAll(directory, 0600) // create all necessary directories
		if err != nil {
			return err
		}
	}

	selfPath, err := os.Executable()
	if err != nil {
		return err
	}

	// get the path of the current exe
	exePath, err := filepath.Abs(selfPath)
	if err != nil {
		return err
	}

	// get the path of the exe in the base directory
	binPath, err := filepath.Abs(BinFile())
	if err != nil {
		return err
	}

	if exePath == binPath {
		// the current process is already the target binPath, no need
		return nil
	}

	log.Warn().Msg(info.APP_NAME + " is not running from the target directory")

	// check if the file exists
	binExists := util.IsFile(binPath)
	if !binExists {
		log.Warn().Msg(info.APP_NAME + " missing from the target directory")
	}

	// if it doesn't, or if -override is supplied, copy the file anyways
	if force || !binExists {
		err := util.CopyFile(exePath, binPath)
		if err != nil {
			return err
		}
		log.Info().Msg("copied the " + info.APP_NAME + " executable")
	}

	return nil
}

// by default, use PROGRAMDATA on Windows and /etc on nix
func dirBase() string {
	switch runtime.GOOS {
	case "windows":
		programdata := os.Getenv("PROGRAMDATA")

		if programdata == "" {
			log.Warn().Msg("%PROGRAMDATA% not found, using C:/" + info.APP_NAME)
			return path.Join("C:", info.APP_NAME)
		} else {
			return path.Join(filepath.Clean(programdata), info.APP_NAME)
		}
	default:
		return path.Join("/etc", info.APP_NAME)
	}
}

// get the base log directory
func dirLogs() string {
	return configPath("logs")
}

func dirKeys() string {
	return configPath("keys")
}

func dirBin() string {
	return configPath("bin")
}

// return the full configuration file path given a filename
func configPath(names ...string) string {
	return path.Join(append([]string{dirBase()}, names...)...)
}

func ClientConfig() string {
	return configPath(strings.ToLower(info.APP_NAME) + ".yaml")
}

// client private key location
func SecretKey() string {
	return path.Join(dirKeys(), "secret.pem")
}

// client public key location
func PublicKey() string {
	return path.Join(dirKeys(), "public.pem")
}

// the JSON cache storage which keeps recent state information
func Cache() string {
	return configPath("cache.json")
}

func LogDir() string {
	return dirLogs()
}

func LogFile() string {
	return path.Join(dirLogs(), strings.ToLower(info.APP_NAME)+".log")
}

func BinFile() string {
	return path.Join(dirBin(), strings.ToLower(info.APP_NAME)+".exe")
}

type Configuration struct {
	Server struct {
		Url      string `yaml:"url"`
		Insecure bool   `yaml:"insecure,omitempty"`
	} `yaml:"server"`
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
}

func (cfg *Configuration) Save(filename string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return err
	}

	return nil
}

func LoadConfigFile(filename string) (*Configuration, error) {
	var conf Configuration

	// requires a configuration file
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

func LoadConfig() (*Configuration, error) {
	fname := ClientConfig()

	// if there is already a config file, load it and return
	if util.IsFile(fname) {
		return LoadConfigFile(fname)
	}

	// if there is no config file, extract it from the environment variable
	url := os.Getenv("SWEETTOOTH_SERVER_URL")

	// if the environment variable is not empty...
	if url != "" {
		// ... then create a default config which will be used on future executions
		var cfg Configuration
		cfg.Server.Url = url
		cfg.Logging.Level = DEFAULT_LOGLEVEL
		if err := cfg.Save(fname); err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	return nil, errors.New("server URL was not found in configuration file or environment variable")
}
