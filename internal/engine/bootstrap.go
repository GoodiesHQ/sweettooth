package engine

import (
	"github.com/goodieshq/sweettooth/internal/choco"
	"github.com/goodieshq/sweettooth/internal/crypto"
	"github.com/goodieshq/sweettooth/internal/crypto/dpapi"
	"github.com/goodieshq/sweettooth/internal/schedule"
	"github.com/goodieshq/sweettooth/internal/tracker"
	"github.com/goodieshq/sweettooth/pkg/info"
	"github.com/rs/zerolog/log"
)

// returns true if all bootstrap procedures succeed, panics otherwise.
func (engine *SweetToothEngine) Bootstrap() {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	if engine.bootstrapped {
		return
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

	log.Info().
		Str("public_key", crypto.GetPublicKeyBase64()).
		Msgf("Initialized " + info.APP_NAME + " Client")

	// once initialized, the node ID should be permanent
	AddLogKey("nodeid", engine.client.NodeID())

	engine.bootstrapped = true
}
