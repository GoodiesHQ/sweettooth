package main

import (
	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/goodieshq/sweettooth/pkg/tracker"
	"github.com/rs/zerolog/log"
)

func doTracker(cli *client.SweetToothClient) {
	log.Trace().Str("routine", "doTracker").Msg("called")
	defer log.Trace().Str("routine", "doTracker").Msg("finished")

	log.Trace().Msg("doTracker called")

	// get a list of packages from the server, really only needs to be done once per execution
	if tracker.IsEmpty() {
		pkg, err := cli.GetPackages()
		if err != nil {
			log.Panic().Err(err).Msg("failed to get existing packages")
		}

		log.Debug().
			Int("choco_count", len(pkg.PackagesChoco)).
			Int("outdated_count", len(pkg.PackagesOutdated)).
			Int("system_count", len(pkg.PackagesSystem)).
			Msg("inventory received from server")

		log.Trace().Msg("cacheing the package inventory in the software tracker")
		// update the tracker's local cache with the existing packages
		tracker.SetPackages(*pkg)
	}

	// track package changes based on the cache file to determine if any software changes have occurred
	pkg, changed, err := tracker.Track()
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
		if err := cli.UpdatePackages(pkg); err != nil {
			log.Panic().Err(err).Msg("failed to update server package inventory")
		}

		// update the cached packages
		tracker.SetPackages(*pkg)

		log.Trace().Msg("successfully updated server's package inventory")
	} else {
		log.Trace().Msg("software tracker identified no software changes")
	}
}
