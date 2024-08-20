package main

import (
	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/goodieshq/sweettooth/pkg/schedule"
	"github.com/goodieshq/sweettooth/pkg/util"
	"github.com/rs/zerolog/log"
)

func doSchedule(cli *client.SweetToothClient) {
	log.Trace().Str("routine", "doSchedule").Msg("called")
	defer log.Trace().Str("routine", "doSchedule").Msg("finished")

	// acquire the most up-to-date schedule from the server's database
	sch, err := cli.GetSchedule()
	if err != nil {
		// any error is unexpected, go ahead and panic
		log.Panic().Err(err).Msg("failed to get client schedule")
	}
	log.Debug().RawJSON("schedule", []byte(util.Dumps(sch))).Msg("received schedule from server")

	if sch != nil {
		// set the schedule on the system
		schedule.SetSchedule(sch)
	}
}
