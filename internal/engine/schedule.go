package engine

import (
	"github.com/goodieshq/sweettooth/internal/schedule"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/rs/zerolog/log"
)

// client routine which acquires the node's assigned maintenance windows
func (engine *SweetToothEngine) Schedule() {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	log.Trace().Str("routine", "Schedule").Msg("called")
	defer log.Trace().Str("routine", "Schedule").Msg("finished")

	engine.mustRun()

	// acquire the most up-to-date schedule from the server's database
	sch, err := engine.client.GetSchedule()
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
