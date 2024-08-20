package main

import (
	"net/http"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/goodieshq/sweettooth/pkg/crypto"
	"github.com/goodieshq/sweettooth/pkg/system"
	"github.com/goodieshq/sweettooth/pkg/tracker"
	"github.com/goodieshq/sweettooth/pkg/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ensure the node is registered with the server
func doRegister(cli *client.SweetToothClient) bool {
	log.Trace().Str("routine", "doRegister").Msg("called")
	defer log.Trace().Str("routine", "doRegister").Msg("finished")

	// perform an initial check for the current status and determine if a registration is required
	if !cli.Registered {
		log.Debug().Msg("running a check to determine registration status")
		err := cli.Check()
		switch err {
		case nil:
			cli.Registered = true
			log.Info().Msg("Node Status: ✅ registered, approved")
		case client.ErrNodeNotApproved:
			cli.Registered = true
			log.Warn().Msg("Node Status: ⛔ registered, not yet approved")
		case client.ErrNodeNotRegistered:
			cli.Registered = false
			log.Warn().Msg("Node Status: ⚠️ not yet registered")
		default:
			cli.Registered = false
			if !client.LogIfApiErr(err) {
				log.Error().Err(err).Msg("check failed")
			}
		}
	}

	if !cli.Registered {
		log.Trace().Msg("generating registration request")

		log.Trace().Msg("gathering system information")
		var registration api.RegistrationRequest
		info, err := system.GetSystemInfo()
		if err != nil {
			log.Error().Err(err).Send()
			return false
		}

		log.Trace().Msg("gathering software information")
		pkg, _, err := tracker.Track()
		if err != nil {
			log.Error().Err(err).Send()
			return false
		}

		// set the node registration information
		registration.Hostname = info.Hostname
		registration.Token = uuid.MustParse("89e07b4e-3943-4ee1-8f06-e63b65892289")
		// registration.OrganizationID = organization_id
		registration.ClientVersion = util.VERSION
		registration.PublicKey = crypto.GetPublicKeyBase64()
		registration.PublicKeySig = crypto.GetPublicKeySigBase64()
		registration.Label = nil
		registration.OSKernel = info.OSInfo.Kernel
		registration.OSName = info.OSInfo.Name
		registration.OSMajor = info.OSInfo.Major
		registration.OSMinor = info.OSInfo.Minor
		registration.OSBuild = info.OSInfo.Build
		registration.PackagesChoco = pkg.PackagesChoco
		registration.PackagesSystem = pkg.PackagesSystem
		registration.PackagesOutdated = pkg.PackagesOutdated

		code, err := cli.Register(&registration)
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

		cli.Registered = true
	}

	// if a panic did not occur, then the registration was successful
	return cli.Registered

}
