package system

import (
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows/registry"
)

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
			continue
		}
		defer key.Close()

		subpaths, err := key.ReadSubKeyNames(-1)
		if err != nil {
			return nil, err
		}

		for _, subpath := range subpaths {
			// Open the subkey for reading
			subkey, err := registry.OpenKey(key, subpath, registry.READ)
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to open subpath %s", subpath)
				continue
			}

			name, _, err := subkey.GetStringValue("DisplayName")
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to get name of %s", subpath)
				subkey.Close()
				continue
			}

			version, _, err := subkey.GetStringValue("DisplayVersion")
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to get version of %s", subpath)
				subkey.Close()
				continue
			}

			packages = append(packages, util.Software{
				Name:    name,
				Version: version,
			})

			subkey.Close()
		}
	}

	return packages, nil
}
