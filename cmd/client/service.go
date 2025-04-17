package main

import (
	"context"
	"strings"
	"time"

	"github.com/goodieshq/sweettooth/internal/client/engine"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/goodieshq/sweettooth/pkg/info"
	"github.com/google/uuid"
	"github.com/kardianos/service"
	"github.com/rs/zerolog/log"
)

var ServiceConfig = &service.Config{
	Name:        strings.ToLower(info.APP_NAME),
	DisplayName: info.APP_NAME + " v" + info.APP_VERSION,
	Description: "Centrally managed application by " +
		"Datalink Networks to manage and monitor 3rd party software at scale.",
	Executable: config.BinFile(),
	Arguments:  []string{"run"},
	Option: service.KeyValue{
		"Restart":                "always",        // restart always
		"SuccessExitCode":        0,               // consider exit code 0 as a successful exit
		"LogDirectory":           config.LogDir(), // sweettooth log directory
		"StartType":              "automatic",     // start up the service with the system
		"OnFailure":              "restart",       // restart the service if it ever panics/fails
		"OnFailureDelayDuration": "5s",            // sleep for 5s after failure
		"OnFailureResetPeriod":   10,              // sleep for 10s before reset
	},
}

type SweetToothService struct {
	engine *engine.SweetToothEngine
}

func (p *SweetToothService) Start(s service.Service) error {
	p.engine.Start()
	return nil
}

func (p *SweetToothService) Stop(s service.Service) error {
	p.engine.Stop()
	return nil
}

// get the status of the current service
func getStatus(svc service.Service) service.Status {
	status, err := svc.Status()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to check the service status")
	}
	return status
}

// start the sweettooth service
func runStart(svc service.Service) {
	if getStatus(svc) == service.StatusRunning {
		log.Warn().Msg("service is already running")
		return
	}
	if err := svc.Start(); err != nil {
		log.Fatal().Err(err).Msg("failed to start the service")
	}
	log.Info().Msg("started the service")
}

// stop the sweettooth service
func runStop(svc service.Service) {
	if getStatus(svc) == service.StatusStopped {
		log.Warn().Msg("service is already stopped")
		return
	}
	if err := svc.Stop(); err != nil {
		log.Fatal().Err(err).Msg("failed to stop the service")
	}
	log.Info().Msg("stopped the service")
}

// get the sweettooth service status
func runStatus(svc service.Service) {
	switch getStatus(svc) {
	case service.StatusRunning:
		log.Info().Str("status", "running").Send()
	case service.StatusStopped:
		log.Info().Str("status", "stopped").Send()
	case service.StatusUnknown:
		fallthrough
	default:
		log.Info().Str("status", "unknown").Send()
	}
}

// register the node with the central server using a registration token
func runRegister(eng *engine.SweetToothEngine, token string) {
	defer util.Recoverable(false)
	if token == "" {
		log.Fatal().Msg("no registration token provided")
	}

	tok, err := uuid.Parse(token)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid registration token provided")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	if !eng.Register(ctx, tok) {
		log.Fatal().Msg("failed to register with the server")
	}

	log.Info().Msg("successfully registered with the server")
}

// install the sweettooth service
func runInstall(svc service.Service, nopath bool) {
	if err := svc.Install(); err != nil {
		log.Error().Err(err).Msg("failed to install the service")
		return
	} else {
		log.Info().Msg("installed service")
		if err := svc.Start(); err != nil {
			log.Error().Err(err).Msg("failed to start the service")
		} else {
			log.Info().Msg("started the service")
		}
	}

	if !nopath {
		pathAdd() // add the sweettooth directory to the %PATH%
	}
}

// uninstall and remove the service
func runUninstall(svc service.Service) {
	status, err := svc.Status()
	if err != nil {
		log.Error().Err(err).Msg("failed to get the service status while uninstalling")
		// not necessarily fatal, but not sure when this would fail and uninstall would succeed
	}

	if status == service.StatusRunning {
		if err := svc.Stop(); err != nil {
			log.Fatal().Err(err).Msg("failed to stop the service")
		}
		log.Info().Msg("stopped the service")
	}

	if err := svc.Uninstall(); err != nil {
		log.Fatal().Err(err).Msg("failed to uninstall the service")
	}

	pathDel() // remove sweettooth from %PATH% even if -nopath is provided

	log.Info().Msg("uninstalled the service and cleaned up the PATH")
}
