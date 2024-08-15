package main

import (
	"crypto/tls"
	"errors"
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

func waitForCheck(cli *client.SweetToothClient) bool {
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

func recoverable(silent bool) {
	if r := recover(); r != nil {
		if !silent {
			var evt = log.Error()
			switch r := r.(type) {
			case error:
				evt = evt.AnErr("recovered", r)
			case string:
				evt = evt.AnErr("recovered", errors.New(r))
			default:
				evt = evt.Any("recovered", r)
			}
			evt.Stack().Msg("client panicked")
		}
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
	for !repeat(cli, register, msg)() {
		util.Countdown("Re-trying registration procedure in ", 5, "s...")
	}

	// wait for the first successful check in (wait for an admin to approve the public key if necessary)
	waitForCheck(cli)
	log.Info().Msg("successfully checked in!")

	// acquire the schedule for this node, just in case it has changed.
	sched, err := cli.Schedule()
	if err != nil {
		log.Panic().Err(err).Msg("failed to get client schedule")
	}

	if sched != nil {
		// set the schedule on the system
		schedule.SetSchedule(sched)
	}

	// TODO: update sources

	// track package changes based on the cache file to determine if any software changes have occurred
	pkg, changed, err := tracker.Track()
	if err != nil {
		log.Panic().Err(err).Msg("software tracker failed to run")
	}

	// if the software has changed
	if changed {
		log.Info().Msg("software tracker identified software changes, updating cache")
		// update tracker cache with the current packages
		err = tracker.SetPackages(pkg)
		if err != nil {
			log.Panic().Err(err).Msg("failed to update the software tracker cache")
		}
	} else {
		log.Debug().Msg("software tracker identified no software changes")
	}

	// check package jobs if it is currently in a maintenance schedule
	if schedule.Now() {
		// TODO: perform package related jobs
	}

	// fmt.Println(util.Dumps(sched))
	return true
}

func loopRecoverable(cli *client.SweetToothClient) (successful bool) {
	defer recoverable(false)
	loop(cli)
	return true
}

func main() {
	// initialize the terminal logger for human-friendly output
	initLoggingTerm()
	schedule.SetSchedule(schedule.Schedule([]schedule.ScheduleEntry{
		{
			RRule:   "FREQ=DAILY;INTERVAL=1",
			TimeBeg: schedule.NewTime16(2, 00),
			TimeEnd: schedule.NewTime16(2, 29),
		},
	}))

	// display the obligatory banner
	banner()

	// initialize the configuration directory which stores the keys, schedule, and other information
	if err := config.Bootstrap(); err != nil {
		log.Panic().Err(err).Msg("Failed to bootstrap the local config directory")
	}
	log.Info().Msg("✅ Bootstrapped config directory")

	// load the configuration file (or create a default config)
	cfg, err := loadConf()
	if err != nil {
		log.Panic().Err(err).Send()
	}

	// set logfile output
	initLoggingFile(config.LogFile())

	// set the level if provided, use info by default
	setLogLevel(cfg.Logging.level)

	// if insecure is used, then ignore SSL errors with the URL (not recommended)
	if cfg.Server.Insecure {
		log.Warn().Msg("ignoring SSL errors with server")
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// create a new SweetTooth client
	cli := client.NewSweetToothClient(cfg.Server.Url)

	// just keep checking in forever every 15 seconds
	go func() {
		for {
			func() {
				defer recoverable(true)
				err := cli.Check()
				log.Debug().Err(err).Msg("background check in")
			}()
			time.Sleep(time.Second * 15)
		}
	}()

	for {
		if !loopRecoverable(cli) {
			log.Info().Msg("Recovered from an error. Re-starting application loop.")
		}
		time.Sleep(5 * time.Second)
	}
}
