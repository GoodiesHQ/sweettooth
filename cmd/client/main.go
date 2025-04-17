package main

import (
	"flag"
	"os"

	"github.com/goodieshq/sweettooth/internal/client/engine"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/goodieshq/sweettooth/pkg/info"
	"github.com/kardianos/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type SweetToothCommand string

const (
	CMD_INSTALL   SweetToothCommand = "install"
	CMD_REGISTER  SweetToothCommand = "register"
	CMD_UNINSTALL SweetToothCommand = "uninstall"
	CMD_UPDATE    SweetToothCommand = "update"
	CMD_RUN       SweetToothCommand = "run"

	// status commands
	CMD_STATUS  SweetToothCommand = "status"
	CMD_START   SweetToothCommand = "start"
	CMD_STOP    SweetToothCommand = "stop"
	CMD_RESTART SweetToothCommand = "restart"
)

func main() {
	// process flags and args
	args := processArgs()

	// immediately initialize the terminal logger for human-friendly output
	engine.LoggingBasic()

	// ensure the process is run with administrative privileges
	mustBeAdmin()

	// CMD: update the sweettooth executable
	if args.Command == "update" {
		updated, err := update(info.APP_NAME)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to update")
		}
		if updated {
			return
		}
	}

	// create the engine without a configuration to prevent the need for bootstrapping yet
	eng := engine.NewSweetToothEngine(nil)

	// create a startup service configuration that should run when the system boots up
	stsvc := &SweetToothService{
		engine: eng,
	}

	// create the kardianos/service object from the program and service configuration
	svc, err := service.New(stsvc, ServiceConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register the service")
	}

	if cmdSvc(args.Command, svc) {
		os.Exit(0)
	}

	// an obligatory old school banner :)
	if !*args.Quiet {
		banner()
	}

	log.Info().Msg("Initializing " + info.APP_NAME + "...")

	// initialize the configuration directory which stores the keys, logs, binary, config, etc
	if err := config.Bootstrap(*args.Force); err != nil {
		log.Fatal().Err(err).Msg("failed to bootstrap the local config directory")
	}
	log.Info().Msg("✅ Bootstrapped the config directory")

	// CMD: install, create a new configuration file from the CLI flags (only on install)
	if args.Command == CMD_INSTALL {
		buildNewConfig(*args.ServerURL, *args.Insecure, *args.LogLevel)
	}

	// load the configuration file which should exist at this point
	cfg := loadConfig()
	if cfg == nil {
		log.Fatal().Msg("failed to load the " + info.APP_NAME + " configuration file")
		os.Exit(-1)
	}

	// initialize the engine with the loaded configuration
	eng.LoadConfig(cfg)

	// set logfile output
	log.Debug().Msg("enabling file logging (" + cfg.Logging.Level + ")")
	engine.LoggingFile(config.LogFile(), cfg.Logging.Level)

	// create a new SweetTooth engine and client instance
	log.Debug().Msg("bootstrapping the " + info.APP_NAME + " engine")
	eng.Bootstrap()

	switch args.Command {
	case CMD_INSTALL:
		runInstall(svc, *args.NoPath)
		if *args.Token == "" {
			break
		}
		fallthrough
		// register the node on installation if the token is provided
	case CMD_REGISTER:
		runRegister(eng, *args.Token)
	case CMD_UNINSTALL:
		runUninstall(svc)
	case CMD_RUN:
		svc.Run()
	default:
		flag.Usage()
	}
}

// process commands for interacting with the service. returns true if the command is service related
func cmdSvc(command SweetToothCommand, svc service.Service) bool {
	// check if the command is related to the service configuration
	switch command {
	case CMD_STATUS:
		runStatus(svc)
		return true
	case CMD_START:
		runStart(svc)
		return true
	case CMD_STOP:
		runStop(svc)
		return true
	case CMD_RESTART:
		runStop(svc)
		runStart(svc)
		return true
	default:
		return false
	}
}

type SweetToothArgs struct {
	Command   SweetToothCommand
	ServerURL *string
	Token     *string
	LogLevel  *string
	Insecure  *bool
	NoPath    *bool
	Force     *bool
	Quiet     *bool
}

func processArgs() *SweetToothArgs {
	var args = new(SweetToothArgs)
	args.ServerURL = flag.String("url", "", "The Server URL for the SweetToooth server (e.g. \"https://sweettooth.example.com\")")
	args.Token = flag.String("token", "", "The token used to register the node with the server")
	args.LogLevel = flag.String("loglevel", zerolog.LevelInfoValue, "The logging level to use for sweettooth")
	args.Insecure = flag.Bool("insecure", false, "Disable HTTPS certificate verification (not recommended)")
	args.NoPath = flag.Bool("nopath", false, "Disables modifications to %PATH%")
	args.Force = flag.Bool("force", false, "Copy the current executable to the sweettooth base directory even if it exists")
	args.Quiet = flag.Bool("q", false, "Quiet")

	flag.Parse()

	// proces command from positional argument
	positional := flag.Args()
	if len(positional) <= 0 {
		flag.Usage()
		os.Exit(1)
	}
	args.Command = SweetToothCommand(positional[0])

	return args
}
