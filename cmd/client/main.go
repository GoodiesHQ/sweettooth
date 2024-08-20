package main

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/goodieshq/sweettooth/pkg/schedule"
	"github.com/goodieshq/sweettooth/pkg/tracker"
	"github.com/goodieshq/sweettooth/pkg/util"
	"github.com/rs/zerolog/log"
	// "github.com/rs/zerolog/log"
)

const (
	DEFAULT_PERIOD_CHECKIN = time.Second * 60 // checkin period
	DEFAULT_PERIOD_LOOP    = time.Second * 10 // loop every 30 seconds
	DEFAULT_PERIOD_RECOVER = time.Second * 10 // recover after 30 seconds
)

func doWaitForCheck(cli *client.SweetToothClient) bool {
	log.Trace().Str("routine", "doWaitForCheck").Msg("called")
	defer log.Trace().Str("routine", "doWaitForCheck").Msg("finished")

	for {
		if err := cli.Check(); err == nil {
			return true
		} else if err == client.ErrNodeNotRegistered {
			cli.Registered = false
			log.Panic().Msg("node is no longer registered")
		}
		util.Countdown("Next check in attempt in ", 10, " seconds...")
	}
}

func repeat(cli *client.SweetToothClient, fn func(cli *client.SweetToothClient) bool, warning string) func() bool {
	var said bool
	return func() bool {
		for !fn(cli) {
			if !said {
				log.Warn().Msg(warning)
				said = true
			}
		}
		return true
	}
}

// this is the main logic app
func loop(cli *client.SweetToothClient) bool {
	var msg string

	// bootstrap the system if it hasn't already, otherwise silently continue
	msg = "Failed to bootstrap the node. Retrying until success or panic..."
	for !repeat(cli, bootstrap, msg)() {
		util.Countdown("Re-trying bootstrap procedure in ", 5, "s...")
	}

	// register the client if it is not already registered, otherwise silently continue
	msg = "Failed to register the node. Retrying until success or panic..."
	for !repeat(cli, doRegister, msg)() {
		util.Countdown("Re-trying registration procedure in ", 5, "s...")
	}

	// wait for the first successful check in (wait for an admin to approve the public key if necessary)
	doWaitForCheck(cli)
	log.Debug().Msg("successfully checked in!")

	// TODO: update sources

	// acquire the schedule for this node, just in case it has changed.
	doSchedule(cli)

	// check package jobs if it is currently in a maintenance schedule
	if schedule.Now() || /* TODO: remove this*/ true {
		doPackageJobs(cli)
	}

	// inventory local software and compare with server's inventory, update if needed
	doTracker(cli)

	return true
}

func loopRecoverable(cli *client.SweetToothClient) (successful bool) {
	defer util.Recoverable(false)
	loop(cli)
	return true
}

func main() {
	// initialize the terminal logger for human-friendly output
	initLoggingTerm()

	// display the obligatory banner
	banner()

	// initialize the configuration directory which stores the keys, schedule, and other information
	if err := config.Bootstrap(); err != nil {
		log.Panic().Err(err).Msg("Failed to bootstrap the local config directory")
	}
	log.Info().Msg("✅ Bootstrapped config directory")

	// load the configuration file (or create a default config)
	log.Trace().Msg("Loading configuration file")
	cfg, err := loadConf()
	if err != nil {
		log.Panic().Err(err).Send()
	}

	// set logfile output
	initLoggingFile(config.LogFile())

	// set the level if provided, use info by default
	setLogLevel(cfg.Logging.Level)

	// if insecure is used, then ignore SSL errors with the URL (not recommended)
	if cfg.Server.Insecure {
		log.Warn().Msg("ignoring SSL errors with server")
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// create a new SweetTooth client instance
	cli := client.NewSweetToothClient(cfg.Server.Url)

	// just keep checking in forever every checkin period regardless of what the result is, updates last_seen in the database
	go func() {
		for {
			func() {
				defer util.Recoverable(true)
				err := cli.Check()
				// only output on Trace debug level
				log.Trace().Err(err).Msg("background check in")
			}()
			time.Sleep(DEFAULT_PERIOD_CHECKIN)
		}
	}()

	log.Debug().Msg("entering client logic loop.")

	for {
		if !loopRecoverable(cli) {
			// a panic is by definition unexpected behavior
			log.Info().Msg("Recovered from an error. Re-starting application loop.")

			// reset the registration status
			cli.Registered = false
			tracker.Reset()

			// sleep for a short time before re-starting
			time.Sleep(DEFAULT_PERIOD_RECOVER)
			continue
		}
		// client loop completed successfully, wait some time for the next loop
		time.Sleep(DEFAULT_PERIOD_LOOP)
	}
}
