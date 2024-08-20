package tracker

import (
	"reflect"
	"sync"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/goodieshq/sweettooth/pkg/choco"
	"github.com/rs/zerolog/log"
)

var mu sync.Mutex
var packages *api.Packages

func isDifferent(p1, p2 *api.Packages) bool {
	return (p1 == nil || p2 == nil) || !reflect.DeepEqual(*p1, *p2)
}

func IsEmpty() bool {
	return packages == nil
}

func Reset() {
	mu.Lock()
	defer mu.Unlock()

	packages = nil
}

func SetPackages(packagesNew api.Packages) {
	mu.Lock()
	defer mu.Unlock()

	packages = &packagesNew
}

func Track() (*api.Packages, bool, error) {
	log.Trace().Msg("tracker.Track called")

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

	mu.Lock()
	defer mu.Unlock()

	return &pkg, isDifferent(&pkg, packages), nil
}

func Bootstrap() error {
	// set an empty cache to force a sync with the server first
	Reset()
	return nil
}
