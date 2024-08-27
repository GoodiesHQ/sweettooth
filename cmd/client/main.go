package main

import (
	"flag"

	"github.com/goodieshq/sweettooth/internal/engine"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/kardianos/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func usage() {
	log.Fatal().
		Str("install", "sweettooth.exe -url <server url> -token <registration token> [-insecure] [-loglevel (level)] install").
		Str("uninstall", "sweeettooth.exe uninstall").
		Str("run", "sweeettooth.exe run").
		Msg("usage")
}

func main() {
	/* these arguments are used for install only, but should be processed first, checked later */
	serverurl := flag.String("url", "", "The Server URL for the SweetToooth server (e.g. \"https://sweettooth.example.com\")")
	token := flag.String("token", "", "The token used to register the node with the server")
	loglevel := flag.String("loglevel", zerolog.LevelInfoValue, "The logging level to use for sweettooth")
	insecure := flag.Bool("insecure", false, "Disable HTTPS certificate verification (not recommended)")
	notoken := flag.Bool("notoken", false, "This node does not require a token (not common)")
	override := flag.Bool("override", false, "Copy the current executable to the sweettooth base directory even if it exists")

	flag.Parse()
	args := flag.Args()

	// an obligatory old school banner :)
	banner()

	// immediately initialize the terminal logger for human-friendly output
	engine.LoggingBasic()

	log.Info().Msg("Initializing " + config.APP_NAME + "...")

	// expects at least one argument: run, install, uninstall, start, stop
	if len(args) == 0 {
		usage()
	}

	// initialize the configuration directory which stores the keys, logs, binary, config, etc
	if err := config.Bootstrap(*override); err != nil {
		log.Fatal().Err(err).Msg("failed to bootstrap the local config directory")
	}
	log.Info().Msg("✅ Bootstrapped the config directory")

	mustBeAdmin() // ensure the process is run with administrative privileges

	// if installing, create a new configuration file from the CLI flags
	if args[0] == "install" {
		buildNewConfig(*serverurl, *insecure, *loglevel)
	}

	cfg := loadConfig() // load the configuration file which should exist at this point

	// set logfile output
	log.Debug().Msg("enabling file logging...")
	engine.LoggingFile(config.LogFile(), cfg.Logging.Level)

	// create the engine
	eng := engine.NewSweetToothEngine(cfg)

	// create a new SweetTooth engine and client instance
	log.Debug().Msg("bootstrapping the " + config.APP_NAME + " engine")
	eng.Bootstrap()

	// create a startup service that should run when the system boots up
	prog := &SweetToothProgram{
		engine: eng,
	}

	svc, err := service.New(prog, ServiceConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register the service")
	}

	switch args[0] {
	case "status":
		runStatus(svc)
	case "install":
		runInstall(svc, eng, *token, *notoken)
	case "start":
		runStart(svc)
	case "stop":
		runStop(svc)
	case "uninstall":
		runUninstall(svc)
	case "run":
		svc.Run()
	default:
		usage()
	}
}
