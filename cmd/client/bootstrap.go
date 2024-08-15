package main

import (
	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/goodieshq/sweettooth/pkg/choco"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/goodieshq/sweettooth/pkg/crypto"
	"github.com/goodieshq/sweettooth/pkg/crypto/dpapi"
	"github.com/goodieshq/sweettooth/pkg/schedule"
	"github.com/goodieshq/sweettooth/pkg/tracker"
	"github.com/rs/zerolog/log"
)

var isBootstrapped = false

// returns true if all bootstrap procedures succeed, panics otherwise.
func bootstrap(cli *client.SweetToothClient) bool {
	if isBootstrapped {
		return true
	}

	// initialize the crypto module by generating new keys or loading existing keys from disk
	if err := crypto.Bootstrap(dpapi.CipherDPAPI{}); err != nil {
		log.Panic().Err(err).Msg("Failed to bootstrap the cryptosystem")
	}
	log.Info().Msg("✅ Bootstrapped crypto system")

	// Bootstrap the scheduler (reset to an empty schedule and let the server provide the schedules)
	if err := schedule.Bootstrap(); err != nil {
		log.Panic().Err(err).Msg("Failed to bootstrap the scheduler.")
	}
	log.Info().Msg("✅ Bootstrapped scheduler")

	// initialize the chocolatey binary or perform installation
	if err := choco.Bootstrap(); err != nil {
		log.Panic().Err(err).Msg("Failed to bootstrap chocolatey")
	}
	log.Info().Msg("✅ Bootstrapped chocolatey")

	// initialize the chocolatey software tracker
	if err := tracker.Bootstrap(); err != nil {
		log.Panic().Err(err).Msg("Failed to bootstrap software tracker")
	}
	log.Info().Msg("✅ Bootstrapped software tracker")

	// once initialized, the node ID should be permanent
	setLogKey("nodeid", cli.NodeID())

	log.Info().
		Str("public_key", crypto.GetPublicKeyBase64()).
		Msgf("Initialized " + config.APP_NAME + " Client")

	isBootstrapped = true
	return true
}
