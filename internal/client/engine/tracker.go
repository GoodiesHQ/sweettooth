package engine

import (
	"time"

	"github.com/goodieshq/sweettooth/internal/client/tracker"
	"github.com/goodieshq/sweettooth/internal/util"
)

const (
	TIMEOUT_TRACKER = time.Second * 30 // plenty of time to run choco list and choco outdated commands
)

// client routine which invokes the software tracker and reports changes to the server
func (engine *SweetToothEngine) Tracker() {
	log := util.Logger("engine.Tracker")
	log.Trace().Msg("called")
	defer log.Trace().Msg("finish")

	log.Trace().Msg("doTracker called")

	// get a list of packages from the server, really only needs to be done once per execution
	if tracker.IsEmpty() {
		pkg, err := engine.client.GetPackages()
		if err != nil {
			log.Error().Err(err).Msg("failed to get existing packages")
			panic(err)
		}

		// output the inventory
		log.Debug().
			Int("choco_count", len(pkg.PackagesChoco)).
			Int("outdated_count", len(pkg.PackagesOutdated)).
			Int("system_count", len(pkg.PackagesSystem)).
			Msg("inventory received from server")

		log.Trace().Msg("cacheing the package inventory in the software tracker")
		// update the tracker's local cache with the existing packages
		tracker.SetPackages(*pkg)
	}

	ctx, cancel := engine.commandContext("tracker.Track", TIMEOUT_TRACKER)
	defer cancel()

	// track package changes based on the cache file to determine if any software changes have occurred
	pkg, changed, err := tracker.Track(ctx)
	if err != nil {
		log.Panic().Err(err).Msg("software tracker failed to run")
	}

	log.Debug().
		Int("choco_count", len(pkg.PackagesChoco)).
		Int("outdated_count", len(pkg.PackagesOutdated)).
		Int("system_count", len(pkg.PackagesSystem)).
		Msg("inventory from client")

	// if the software has changed (server's inventory is out of date)...
	if changed {
		// ... then update tracker cache with the current packages
		log.Debug().Msg("tracker has identified software changes, updating cache")

		// and update the server's packages
		if err := engine.client.UpdatePackages(pkg); err != nil {
			log.Panic().Err(err).Msg("failed to update server package inventory")
		}

		// update the cached packages only if the server update succeeded
		tracker.SetPackages(*pkg)

		log.Trace().Msg("successfully updated server's package inventory")
	} else {
		log.Trace().Msg("software tracker identified no software changes")
	}
}
