package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goodieshq/sweettooth/internal/client/system"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/goodieshq/sweettooth/pkg/info"
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
		log.Fatal().Msg(info.APP_NAME + " must be run as administrator")
	}
}

func setPath(path string) {
	err := exec.Command("setx", "PATH", path).Run()
	log.Err(err).Str("path", path).Msg("setting the system PATH")
	// not fatal, but PATH changes did not succeed
}

func delPath() {
	log.Trace().Msg("removing the " + info.APP_NAME + " bin directory to the PATH")
	binDir := filepath.FromSlash(filepath.Dir(config.BinFile())) // get the directory of the SweetTooth bin file
	pathOld := os.Getenv("PATH")

	pathNew := pathOld
	pathNew = strings.Replace(pathNew, binDir+";", "", -1)
	pathNew = strings.Replace(pathNew, ";"+binDir, "", -1)
	pathNew = strings.Replace(pathNew, binDir, "", -1)

	setPath(pathNew)
}

func addPath() {
	log.Trace().Msg("adding the " + info.APP_NAME + " bin directory to the PATH")
	binDir := filepath.FromSlash(filepath.Dir(config.BinFile())) // get the directory of the SweetTooth bin file
	pathOld := os.Getenv("PATH")
	log.Trace().Msgf("looking for \"%s\" in \"%s\"", binDir, pathOld)
	if !strings.Contains(pathOld, binDir) {
		log.Debug().Msg(info.APP_NAME + " directory not found in the PATH, adding now")
		setPath(pathOld + ";" + binDir)
	} else {
		log.Debug().Msg(info.APP_NAME + " directory already found in the PATH")
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
