package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/goodieshq/sweettooth/pkg/choco"
	"github.com/goodieshq/sweettooth/pkg/config"
	"github.com/goodieshq/sweettooth/pkg/crypto"
	"github.com/goodieshq/sweettooth/pkg/crypto/dpapi"
	"github.com/goodieshq/sweettooth/pkg/schedule"
	"github.com/goodieshq/sweettooth/pkg/system"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const DRIFT_TOLERANCE = 60 * time.Second

func main() {
	// Initialize the logger for human-friendly output
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	var info *system.OSInfo

	data, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	fmt.Println(string(data))

	return

	// Bootstrap the scheduler (reset to an empty schedule and let the server provide the schedules)
	if err := schedule.Bootstrap(); err != nil {
		log.Fatal().Err(err).Msg("Failed to bootstrap the scheduler.")
	}

	log.Info().Msg("Bootstrapped scheduler")

	// create an example schedule
	/*
		schedule.AddEntry(schedule.ScheduleEntry{
			Weeks: []string{"1", "2", "3", "4", "5"},
			Days:  []string{"SUN", "SAT", "MON", "TUE"},
			Start: schedule.ScheduleTime{H: 00, M: 00},
			End:   schedule.ScheduleTime{H: 23, M: 59},
		})
	*/

	// initialize the configuration directory which stores the keys, schedule, and other information
	if err := config.Bootstrap(); err != nil {
		log.Fatal().Err(err).Msg("Failed to bootstrap the local config directory")
	}
	log.Info().Msg("Bootstrapped config directory")

	// initialize the crypto module by generating new keys or loading existing keys from disk
	if err := crypto.Bootstrap(dpapi.CipherDPAPI{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to bootstrap the cryptosystem")
	}
	log.Info().Msg("Bootstrapped cryptosystem")

	if err := choco.Bootstrap(); err != nil {
		log.Fatal().Err(err).Msg("Failed to bootstrap chocolatey")
	}
	log.Info().Msg("Bootstrapped chocolatey")

	return

	const pkg = "procexp"

	result := choco.Package(&choco.PackageParams{
		// Action: choco.ACTION_UNINSTALL,
		Action:         choco.PKG_ACTION_UPGRADE,
		Name:           pkg,
		Version:        "17.4.0",
		IgnoreChecksum: true,
	})

	if result.Err != nil {
		log.Error().Err(result.Err).Send()
	}

	log.Info().Msg("Done!")
}
