package schedule

import (
	"github.com/goodieshq/sweettooth/internal/schedule"
	"github.com/rs/zerolog/log"
)

// Schedule of when updates can be applied
var clientSchedule schedule.Schedule

func Bootstrap() error {
	// by default, the schedule should never return true without explicitly receiving it from the server
	SetSchedule(schedule.Schedule(nil))
	return nil
}

// System-wide, check if the current schedule matches the time right now
func Matches() bool {
	// get the current time, day beginning, and day end timestamps
	today := schedule.Now()

	// return true if the current assigned schedule matches
	return clientSchedule.Matches(&today)
}

// Set the current program's schedule
func SetSchedule(sched schedule.Schedule) {
	var action string

	if sched == nil {
		action = "Cleared"
	} else {
		action = "Updated"
	}

	clientSchedule = sched

	log.Debug().Msg(action + " the local system schedule")
}
