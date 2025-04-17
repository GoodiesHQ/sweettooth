package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/goodieshq/sweettooth/internal/client/system"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/goodieshq/sweettooth/pkg/info"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows/registry"
)

const (
	ENV_REGISTRY_PATH = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
	MAX_PATH_LENGTH   = 4096 // Adjust as needed
)

// Create a new SweetTooth YAML configuration file
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

// Ensures that the process is running as an administrator, fatal out if not
func mustBeAdmin() {
	// invoking chocolatey to manage installations or install the service requires admin permissions
	// it may be POSSIBLE to monitor software without admin permissions, but I thought it would be easier to just simply force it
	log.Trace().Msg("checking for administrator privileges")
	if !system.IsAdmin() {
		log.Fatal().Msg(info.APP_NAME + " must be run as administrator")
	}
}

func getSystemPath() (string, error) {
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		ENV_REGISTRY_PATH,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return "", fmt.Errorf("failed to open system environment key: %w", err)
	}
	defer key.Close()

	path, _, err := key.GetStringValue("Path")
	if err != nil {
		if err == registry.ErrNotExist {
			return "", nil
		}
		return "", fmt.Errorf("failed to get system PATH: %w", err)
	}
	return path, nil
}

func setSystemPath(pathNew string) error {
	if len(pathNew) > MAX_PATH_LENGTH {
		return fmt.Errorf("new PATH length (%d) exceeds maximum allowed (%d)", len(pathNew), MAX_PATH_LENGTH)
	}

	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		ENV_REGISTRY_PATH,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("failed to open system environment key for writing: %w", err)
	}
	defer key.Close()

	err = key.SetStringValue("Path", pathNew)
	if err != nil {
		return fmt.Errorf("failed to set system PATH: %w", err)
	}

	return nil
}

// Remove the SweetTooth bin directory from the PATH environment variable
func pathDel() {
	log.Trace().Msg("removing the " + info.APP_NAME + " bin directory to the PATH")

	// get the clean directory of the SweetTooth bin file
	binDir := filepath.Clean(filepath.FromSlash(filepath.Dir(config.BinFile())))

	// get the existing PATH
	pathOld, err := getSystemPath()
	if err != nil {
		log.Error().Err(err).Msg("failed to get the system path")
		return
	}

	log.Trace().Str("path", pathOld).Msg("old system PATH")

	// separate the PATH into individual paths and create a new empty path
	pathOldParts := strings.Split(pathOld, ";")
	pathNewParts := []string{}

	log.Trace().Msgf("looking for \"%s\" in \"%s\"", binDir, pathOld)

	// iterate over the split path entries
	for _, path := range pathOldParts {
		// check if the path includes the bin directory using unicode case-folding
		if !strings.EqualFold(filepath.Clean(path), binDir) {
			pathNewParts = append(pathNewParts, path)
		} else {
			log.Info().Msg("Found bin directory in PATH. Removing it.")
		}
	}

	// check if the size of path parts has changed. If not, the directory was not found.
	if len(pathNewParts) == len(pathOldParts) {
		log.Info().Msg("The bin directory was not found in PATH. Not modifying.")
		return
	}

	// join the remaining paths back together
	pathNew := strings.Join(pathNewParts, ";")

	log.Trace().Str("path", pathNew).Msg("new system PATH")

	// system PATH has been modified, save it
	if err := setSystemPath(pathNew); err != nil {
		log.Error().Err(err).Msg("failed to get the system path")
		return
	}
}

// Add the SweetTooth bin directory into the PATH environment variable
func pathAdd() {
	log.Trace().Msgf("Adding the %s bin directory to the PATH", info.APP_NAME)
	binDir := filepath.Clean(filepath.FromSlash(filepath.Dir(config.BinFile()))) // get the directory of the SweetTooth bin file

	// get the current system PATH
	pathOld, err := getSystemPath()
	if err != nil {
		log.Error().Err(err).Msg("failed to get the system path")
		return
	}

	log.Trace().Msgf("looking for \"%s\" in \"%s\"", binDir, pathOld)

	// split the PATH into individual paths
	pathOldParts := strings.Split(pathOld, ";")

	for _, path := range pathOldParts {
		// check if the path includes the bin directory using unicode case-folding
		if strings.EqualFold(filepath.Clean(path), binDir) {
			log.Debug().Msgf("%s directory already found in the PATH", info.APP_NAME)
			return
		}
	}

	// set the new system PATH with the bin directory added
	log.Debug().Msgf("%s directory not found in the PATH, adding...", info.APP_NAME)
	pathNew := pathOld + ";" + binDir

	log.Trace().Str("path", pathNew).Msg("new system PATH")
	if err := setSystemPath(pathNew); err != nil {
		log.Error().Err(err).Msg("failed to set the system path")
		return
	}
}

// load the client configuration file from the sweettooth folder
func loadConfig() *config.Configuration {
	// load the configuration file
	log.Trace().Msg("loading configuration file")
	cfg, err := config.LoadConfigFile(config.ClientConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load the configuration file")
	}

	// if insecure is used, then ignore SSL errors with the URL (not recommended for production)
	if cfg.Server.Insecure {
		log.Warn().Msgf("ignoring SSL transport errors with server '%s' due to insecure flag", cfg.Server.Url)
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return cfg
}
