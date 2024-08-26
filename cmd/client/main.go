package main

import (
	"crypto/tls"
	"flag"
	"net/http"

	"github.com/goodieshq/sweettooth/internal/engine"
	"github.com/goodieshq/sweettooth/internal/system"
	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/google/uuid"
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

	log.Info().Msg("✅ Bootstrapped config directory")

	// invoking chocolatey to manage installations requires admin permissions
	// it may be POSSIBLE to monitor software without admin permissions, but I thought it would be easier to just simply force it
	if !system.IsAdmin() {
		log.Fatal().Msg(config.APP_NAME + " must be run as administrator")
	}
	log.Trace().Msg(config.APP_NAME + " is running as administrator")

	if args[0] == "install" {
		// build a new configuration from cmd flags if installing
		var cfg config.Configuration

		if *serverurl == "" {
			log.Fatal().Msg("a valid server URL must be provided")
		}

		cfg.Server.Url = *serverurl
		cfg.Server.Insecure = *insecure
		cfg.Logging.Level = *loglevel

		// save the config from the flag parameters
		cfg.Save(config.ClientConfig())
	}

	// load the configuration file
	log.Trace().Msg("loading configuration file")
	cfg, err := config.LoadConfigFile(config.ClientConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load the configuration file")
	}

	// set logfile output
	log.Debug().Msg("enabling file logging...")
	engine.LoggingFile(config.LogFile(), cfg.Logging.Level)

	// if insecure is used, then ignore SSL errors with the URL (not recommended for production)
	if cfg.Server.Insecure {
		log.Warn().Msg("ignoring SSL transport errors with server " + cfg.Server.Url)
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// create a temporary client and engine
	cli := client.NewSweetToothClient(cfg.Server.Url)
	eng := engine.NewSweetToothEngine(cli)

	// create a new SweetTooth engine and client instance
	log.Debug().Msg("bootstrapping the " + config.APP_NAME + " engine")
	eng.Bootstrap()

	prog := &SweetToothProgram{
		engine: eng,
	}

	svc, err := service.New(prog, ServiceConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register the service")
	}

	switch args[0] {
	case "status":
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
	case "install":
		/* bootstrap the system, install and run the service */
		if *token == "" {
			if !*notoken {
				log.Fatal().Msg("a registration token should be provided. to ignore, use \"-notoken\"")
			}
			// -notoken provided, you better know what you're doing
		} else {
			tok, err := uuid.Parse(*token)
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
	case "start":
		if err := svc.Start(); err != nil {
			log.Fatal().Err(err).Msg("failed to start the service")
		}
		log.Info().Msg("started the service")
	case "stop":
		if err := svc.Stop(); err != nil {
			log.Fatal().Err(err).Msg("failed to stop the service")
		}
		log.Info().Msg("stopped the service")
	case "uninstall":
		if err := svc.Stop(); err != nil {
			log.Fatal().Err(err).Msg("failed to stop the service")
		}
		log.Info().Msg("stopped the service")

		if err := svc.Uninstall(); err != nil {
			log.Fatal().Err(err).Msg("failed to uninstall the service")
		}
		log.Info().Msg("uninstalled the service")
	case "run":
		svc.Run()
	default:
		usage()
	}
}
