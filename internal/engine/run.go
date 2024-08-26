package engine

import (
	"time"

	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/rs/zerolog/log"
)

const (
	DEFAULT_PERIOD_CHECKIN = time.Second * 60 // checkin period
	DEFAULT_PERIOD_LOOP    = time.Second * 10 // loop every 30 seconds
	DEFAULT_PERIOD_RECOVER = time.Second * 10 // recover after 30 seconds

	BYPASS_SCHEDULE = true // development only
)

func (engine *SweetToothEngine) run() {
	// just keep checking in forever every checkin period regardless of what the result is, updates last_seen in the database
	log.Info().Msg("RUN CALLED")
	go backgroundCheckin(engine.client)

	// enter the logic loop
	engine.loop()
}

func backgroundCheckin(cli *client.SweetToothClient) {
	for {
		func() {
			defer util.Recoverable(true) // let this function re-run if it panics
			err := cli.Check()
			log.Trace().Err(err).Msg("background check in") // only output on Trace level or it will fill up log files
		}()
		time.Sleep(DEFAULT_PERIOD_CHECKIN)
	}
}