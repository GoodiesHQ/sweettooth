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
	Description: config.APP_NAME + " is a centrally managed application developed by " +
		"Austin Archer to monitor, report on, and manage 3rd party software at scale.",
	Executable: config.BinFile(),
	Arguments:  []string{"run"},
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
