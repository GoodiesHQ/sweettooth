package main

import (
	"context"
	"strings"
	"time"

	"github.com/goodieshq/sweettooth/internal/client/engine"
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

type SweetToothProgram struct {
	engine *engine.SweetToothEngine
}

func (p *SweetToothProgram) Start(s service.Service) error {
	p.engine.Start()
	return nil
}

func (p *SweetToothProgram) Stop(s service.Service) error {
	p.engine.Stop()
	return nil
}

// start the sweettooth service
func runStart(svc service.Service) {
	if err := svc.Start(); err != nil {
		log.Fatal().Err(err).Msg("failed to start the service")
	}
	log.Info().Msg("started the service")
}

// stop the sweettooth service
func runStop(svc service.Service) {
	if err := svc.Stop(); err != nil {
		log.Fatal().Err(err).Msg("failed to stop the service")
	}
	log.Info().Msg("stopped the service")
}

// get the sweettooth service status
func runStatus(svc service.Service) {
	status, err := svc.Status()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to check the service status")
	}
	switch status {
	case service.StatusRunning:
		log.Info().Str("status", "running").Send()
	case service.StatusStopped:
		log.Info().Str("status", "stopped").Send()
	case service.StatusUnknown:
		log.Info().Str("status", "unknown").Send()
	}
}

// install the sweettooth service
func runInstall(svc service.Service, eng *engine.SweetToothEngine, token string, notoken, nopath bool) {
	if token == "" {
		if !notoken {
			log.Fatal().Msg("a registration token should be provided. to ignore, use \"-notoken\"")
		}
		// -notoken provided, you better know what you're doing!
	} else {
		tok, err := uuid.Parse(token)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid registration token provided")
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if !eng.Register(ctx, tok) {
			log.Fatal().Msg("failed to register with the server")
		}

		if err := svc.Install(); err != nil {
			log.Fatal().Err(err).Msg("failed to install the service")
		}
		log.Info().Msg("installed service")
		if err := svc.Start(); err != nil {
			log.Error().Err(err).Msg("failed to start the service")
		} else {
			log.Info().Msg("started the service")
		}
	}

	if !nopath {
		addPath() // add the sweettooth directory to the %PATH%
	}
}

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

	delPath() // remove sweettooth from %PATH% even if -nopath is provided

	log.Info().Msg("uninstalled the service and cleaned up the PATH")
}
