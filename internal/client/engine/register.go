package engine

import (
	"context"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/client/keys"
	"github.com/goodieshq/sweettooth/internal/client/system"
	"github.com/goodieshq/sweettooth/internal/client/tracker"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/goodieshq/sweettooth/pkg/info"
	"github.com/google/uuid"
)

// client routine to ensure the node is registered with the server
func (engine *SweetToothEngine) Register(ctx context.Context, token uuid.UUID) bool {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	log := util.Logger("engine.Register")
	log.Trace().Msg("called")
	defer log.Trace().Msg("finish")

	// perform an initial check for the current status and determine if a registration is required
	if !engine.client.Registered {
		log.Debug().Msg("running a check to determine registration status")
		err := engine.client.Check()
		switch err {
		case nil:
			engine.client.Registered = true
			log.Info().Msg("Node Status: ✅ registered, approved")
		case client.ErrNodeNotApproved:
			engine.client.Registered = true
			log.Warn().Msg("Node Status: ⛔ registered, not yet approved")
		case client.ErrNodeNotRegistered:
			engine.client.Registered = false
			log.Warn().Msg("Node Status: ⚠️ not yet registered")
		default:
			engine.client.Registered = false
			if !client.LogIfApiErr(err) {
				log.Error().Err(err).Msg("check failed")
			}
		}
	}

	// if it's still not registered
	if !engine.client.Registered {
		log.Trace().Msg("generating registration request")

		log.Trace().Msg("gathering system information")
		var registration api.RegistrationRequest
		sysinfo, err := system.GetSystemInfo()
		if err != nil {
			log.Error().Err(err).Send()
			return false
		}

		log.Trace().Msg("gathering software information")
		pkg, _, err := tracker.Track(ctx)
		if err != nil {
			log.Error().Err(err).Send()
			return false
		}

		// set the node registration information
		registration.Hostname = sysinfo.Hostname
		registration.Token = token
		registration.ClientVersion = info.APP_VERSION
		registration.PublicKey = keys.GetPublicKeyBase64()
		registration.PublicKeySig = keys.GetPublicKeySigBase64()
		registration.Label = nil
		registration.OSKernel = sysinfo.OSInfo.Kernel
		registration.OSName = sysinfo.OSInfo.Name
		registration.OSMajor = sysinfo.OSInfo.Major
		registration.OSMinor = sysinfo.OSInfo.Minor
		registration.OSBuild = sysinfo.OSInfo.Build
		registration.PackagesChoco = pkg.PackagesChoco
		registration.PackagesSystem = pkg.PackagesSystem
		registration.PackagesOutdated = pkg.PackagesOutdated

		code, err := engine.client.Register(&registration)
		if err != nil {
			log.Panic().Err(err).Msg("failed to register the client")
		}

		switch code {
		case http.StatusCreated:
			log.Info().Msg("client was successfully registered")
		case http.StatusNoContent:
			fallthrough
		case http.StatusOK:
			log.Info().Msg("client was already registered")
		case http.StatusForbidden:
			log.Warn().Msg("client is registered but not yet approved")
		case http.StatusUnauthorized:
			log.Panic().Msg("registration token is not authorized")
		default:
			log.Panic().Int("status_code", code).Str("status", http.StatusText(code)).Err(err).Msg("unexpected status code")
		}
		engine.client.Registered = true
	}

	// if a panic did not occur, then the registration was successful
	return engine.client.Registered
}
