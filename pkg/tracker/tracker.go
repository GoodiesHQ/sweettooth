package tracker

import (
	"encoding/json"
	"os"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/goodieshq/sweettooth/pkg/choco"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/goodieshq/sweettooth/pkg/util"
)

var cachedPackages api.Packages

func SetPackages(packages *api.Packages) error {
	cachedPackages = *packages
	return savePackages()
}

func loadPackages() error {
	data, err := os.ReadFile(config.Cache())
	if err != nil {
		return err
	}

	var packages api.Packages
	err = json.Unmarshal(data, &packages)
	if err != nil {
		return err
	}

	SetPackages(&packages)

	return nil
}

func savePackages() error {
	data, err := json.Marshal(&cachedPackages)
	if err != nil {
		return err
	}

	return os.WriteFile(config.Cache(), data, 0600)
}

func Track() (*api.Packages, bool, error) {
	pkgChoco, pkgSystem, err := choco.ListAllInstalled()
	if err != nil {
		return nil, false, err
	}

	pkgOutdated, err := choco.ListChocoOutdated()
	if err != nil {
		return nil, false, err
	}

	pkg := api.Packages{
		PackagesChoco:    pkgChoco,
		PackagesSystem:   pkgSystem,
		PackagesOutdated: pkgOutdated,
	}

	return &pkg, util.Dumps(&pkg) != util.Dumps(&cachedPackages), nil
}

func Bootstrap() error {
	// check if the cache file exists
	if !util.IsFile(config.Cache()) {
		if err := savePackages(); err != nil {
			return err
		}
	}
	if err := loadPackages(); err != nil {
		return err
	}
	return nil
}
