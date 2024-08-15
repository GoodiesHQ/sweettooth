package main

import (
	"net/http"

	"github.com/goodieshq/sweettooth/pkg/api"
	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/goodieshq/sweettooth/pkg/choco"
	"github.com/goodieshq/sweettooth/pkg/crypto"
	"github.com/goodieshq/sweettooth/pkg/system"
	"github.com/goodieshq/sweettooth/pkg/util"
	"github.com/rs/zerolog/log"
)

// ensure the node is registered with the server
func register(cli *client.SweetToothClient) bool {
	// perform an initial check for the current status and determine if a registration is required
	if !cli.Registered { // assume it is not registered by default, subsequent calls to loop will skip this
		err := cli.Check()
		switch err {
		case nil:
			cli.Registered = true
			log.Info().Msg("node is registered and approved")
		case client.ErrNodeNotApproved:
			cli.Registered = true
			log.Warn().Msg("node is registered, but not yet approved")
		case client.ErrNodeNotRegistered:
			cli.Registered = false
			log.Warn().Msg("node has not yet been registered")
		default:
			cli.Registered = false
			if !client.LogIfApiErr(err) {
				log.Error().Err(err).Msg("check failed")
			}
		}
	}

	if !cli.Registered {
		var registration api.RegistrationRequest

		info, err := system.GetSystemInfo()
		if err != nil {
			log.Error().Err(err).Send()
			return false
		}

		pkgChoco, pkgSystem, err := choco.ListAllInstalled()
		if err != nil {
			log.Error().Err(err).Send()
			return false
		}

		// set the node registration information
		registration.Hostname = info.Hostname
		registration.OrganizationID = nil
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
		registration.PackagesChoco = pkgChoco
		registration.PackagesSystem = pkgSystem

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
			log.Info().Msg("client was previously registered")
		case http.StatusNotFound:
			log.Info().Msg("organization ID was not found")
		case http.StatusForbidden:
			log.Warn().Msg("client is registered but not yet approved")
		default:
			log.Panic().Int("status_code", code).Str("status", http.StatusText(code)).Err(err).Msg("unexpected status code")
		}

		cli.Registered = true
	}

	// if a panic did not occur, then the registration was successful
	return cli.Registered

}
