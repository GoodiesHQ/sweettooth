package main

import (
	"strings"

	"github.com/goodieshq/sweettooth/internal/engine"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/google/uuid"
	"github.com/kardianos/service"
	"github.com/rs/zerolog/log"
)

var ServiceConfig = &service.Config{
	Name:        strings.ToLower(config.APP_NAME),
	DisplayName: config.APP_NAME + " v" + util.VERSION,
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
func runInstall(svc service.Service, eng *engine.SweetToothEngine, token string, notoken bool) {
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

		if !eng.Register(tok) {
			log.Fatal().Msg("failed to register with the server")
		}

		if err := svc.Install(); err != nil {
			log.Fatal().Err(err).Msg("failed to install the service")
		}
		log.Info().Msg("installed service")
		if err := svc.Start(); err != nil {
			log.Fatal().Err(err).Msg("failed to start the service")
		}
		log.Info().Msg("started the service")
	}
}

func runUninstall(svc service.Service) {
	if err := svc.Stop(); err != nil {
		log.Fatal().Err(err).Msg("failed to stop the service")
	}
	log.Info().Msg("stopped the service")

	if err := svc.Uninstall(); err != nil {
		log.Fatal().Err(err).Msg("failed to uninstall the service")
	}
	log.Info().Msg("uninstalled the service")
}
