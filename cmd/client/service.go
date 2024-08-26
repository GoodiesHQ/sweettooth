package main

import (
	"strings"

	"github.com/goodieshq/sweettooth/internal/engine"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/kardianos/service"
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
