package tracker

import (
	"context"
	"reflect"
	"sync"

	"github.com/goodieshq/sweettooth/internal/client/choco"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/rs/zerolog/log"
)

var mu sync.Mutex
var packages *api.Packages

func isDifferent(p1, p2 *api.Packages) bool {
	return (p1 == nil || p2 == nil) || !reflect.DeepEqual(*p1, *p2)
}

func IsEmpty() bool {
	defer util.Locker(&mu)()
	return packages == nil
}

func Reset() {
	defer util.Locker(&mu)()
	packages = nil
}

func SetPackages(packagesNew api.Packages) {
	defer util.Locker(&mu)()
	packages = &packagesNew
}

func Track(ctx context.Context) (*api.Packages, bool, error) {
	log.Trace().Msg("tracker.Track called")

	defer util.Locker(&mu)()

	pkgChoco, pkgSystem, err := choco.ListAllInstalled(ctx)
	if err != nil {
		return nil, false, err
	}

	pkgOutdated, err := choco.ListChocoOutdated(ctx)
	if err != nil {
		return nil, false, err
	}

	pkg := api.Packages{
		PackagesChoco:    pkgChoco,
		PackagesSystem:   pkgSystem,
		PackagesOutdated: pkgOutdated,
	}

	return &pkg, isDifferent(&pkg, packages), nil
}

func Bootstrap() error {
	// set an empty cache to force a sync with the server first
	Reset()
	return nil
}
