package system

import (
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows/registry"
)

func getInstalledRegistry(key registry.Key, subpath string) (*util.Software, error) {
	// Open the subkey for reading
	subkey, err := registry.OpenKey(key, subpath, registry.READ)
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to open subpath %s", subpath)
		return nil, err
	}

	name, _, err := subkey.GetStringValue("DisplayName")
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to get name of %s", subpath)
		subkey.Close()
		return nil, err
	}

	version, _, err := subkey.GetStringValue("DisplayVersion")
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to get version of %s", subpath)
		subkey.Close()
		return nil, err
	}

	subkey.Close()
	return &util.Software{
		Name:    name,
		Version: version,
	}, nil
}

func ListInstalled() ([]util.Software, error) {
	paths := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	var packages []util.Software
	var err error

	for _, path := range paths {
		var key registry.Key
		key, err = registry.OpenKey(registry.LOCAL_MACHINE, path, registry.READ)
		if err != nil {
			log.Warn().Str("key", path).Msg("failed to open registry key")
			continue
		}
		defer key.Close()

		subpaths, err := key.ReadSubKeyNames(-1)
		if err != nil {
			return nil, err
		}

		for _, subpath := range subpaths {
			software, err := getInstalledRegistry(key, subpath)
			if err != nil {
				continue
			}

			packages = append(packages, *software)
		}
	}

	return packages, nil
}
