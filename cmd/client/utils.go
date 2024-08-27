package main

import (
	"crypto/tls"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/system"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/rs/zerolog/log"
)

func buildNewConfig(serverurl string, insecure bool, loglevel string) {
	// build a new configuration from cmd flags if installing
	var cfg config.Configuration

	if serverurl == "" {
		log.Fatal().Msg("a valid server URL must be provided")
	}

	cfg.Server.Url = serverurl
	cfg.Server.Insecure = insecure
	cfg.Logging.Level = loglevel

	// save the config from the flag parameters
	cfg.Save(config.ClientConfig())
}

func mustBeAdmin() {
	// invoking chocolatey to manage installations or install the service requires admin permissions
	// it may be POSSIBLE to monitor software without admin permissions, but I thought it would be easier to just simply force it
	log.Trace().Msg("checking for administrator privileges")
	if !system.IsAdmin() {
		log.Fatal().Msg(config.APP_NAME + " must be run as administrator")
	}
}

// load the client configuration file
func loadConfig() *config.Configuration {
	// load the configuration file
	log.Trace().Msg("loading configuration file")
	cfg, err := config.LoadConfigFile(config.ClientConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load the configuration file")
	}

	// if insecure is used, then ignore SSL errors with the URL (not recommended for production)
	if cfg.Server.Insecure {
		log.Warn().Msg("ignoring SSL transport errors with server " + cfg.Server.Url)
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return cfg
}
