package config

import (
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/rs/zerolog/log"
)

const APP_NAME = "SweetTooth"

// create the base directory to store configs and cache
func Bootstrap() (err error) {
	for _, p := range []string{dirBase(), dirLogs()} {
		err = os.MkdirAll(p, 0600)
		if err != nil {
			return
		}
	}
	return
}

// by default, use PROGRAMDATA on Windows and /etc on nix
func dirBase() string {
	switch runtime.GOOS {
	case "windows":
		programdata := os.Getenv("PROGRAMDATA")
		if programdata == "" {
			log.Warn().Msg("%PROGRAMDATA% not found, using C:/" + APP_NAME)
			return path.Join("C:", APP_NAME)
		} else {
			return path.Join(programdata, APP_NAME)
		}
	default:
		return path.Join("/etc", APP_NAME)
	}
}

// get the base log directory
func dirLogs() string {
	return configPath("logs")
}

func dirKeys() string {
	return configPath("keys")
}

// return the full configuration file path given a filename
func configPath(names ...string) string {
	return path.Join(append([]string{dirBase()}, names...)...)
}

func ClientConfig() string {
	return configPath(strings.ToLower(APP_NAME) + ".yaml")
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

func LogFile() string {
	return path.Join(dirLogs(), strings.ToLower(APP_NAME)+".log")
}
