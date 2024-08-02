package config

import (
	"os"
	"path"
	"runtime"

	"github.com/rs/zerolog/log"
)

const APP_NAME = "SweetTooth"

// create the base directory to store configs and cache
func Bootstrap() error {
	return os.MkdirAll(baseDirectory(), 0600)
}

// by default, use PROGRAMDATA on Windows and /etc on nix
func baseDirectory() string {
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

// return the full configuration file path given a filename
func configFile(filename string) string {
	err := Bootstrap()
	if err != nil {
		log.Error().Err(err).Msg("Failed to bootstrap the configuration directory")
	}
	return path.Join(baseDirectory(), filename)
}

// client private key location
func SecretKey() string {
	return configFile("secret.pem.enc")
}

// client public key location
func PublicKey() string {
	return configFile("public.pem")
}

// the JSON cache storage which keeps recent state information
func Cache() string {
	return configFile("cache.json")
}
