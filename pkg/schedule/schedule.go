package schedule

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/teambition/rrule-go"
)

type Schedule []ScheduleEntry
type SchedGroup string
type ScheduleGroup map[SchedGroup]Schedule

const (
	ISO8601 = "20060102T150405Z"
)

var (
	// RRULE entries should be processed starting Jan 1 1970 according to the local time zone

	SCHED_START_TIME        = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.Local).UTC()
	SCHED_START_TIME_SUFFIX = fmt.Sprintf(";DTSTART=%s", SCHED_START_TIME.Format(ISO8601))
)

const (
	SchedGroupNode         SchedGroup = "node"
	SchedGroupGroup        SchedGroup = "group"
	SchedGroupOrganization SchedGroup = "organization"
)

type Time16 uint16

func (t Time16) H() int {
	return int(t >> 8 & 0xff)
}

func (t Time16) M() int {
	return int(t >> 8 & 0xff)
}

func NewTime16(hr int, min int) Time16 {
	h := uint16(hr & 0xff)
	m := uint16(min & 0xff)
	return Time16(h<<8 | m)
}

var schedule Schedule

type AllDay struct {
	T   time.Time // a given arbitrary timestamp
	Beg time.Time // the beginning of the day T is in
	End time.Time // the end of the day T is in (Beg + 24hrs)
}

type ScheduleEntry struct {
	RRule   string `json:"rrule"`
	TimeBeg Time16 `json:"time_beg"`
	TimeEnd Time16 `json:"time_end"`
}

func Bootstrap() error {
	// by default, the schedule should never return true without explicitly receiving it from the server
	SetSchedule(Schedule(nil))
	return nil
}

func SetSchedule(sched Schedule) {
	schedule = sched
	log.Debug().Msg("Updated the local system schedule")
}

func (entry ScheduleEntry) matches(allday *AllDay) bool {
	if entry.RRule == "" {
		// empty rules are always false
		return false
	}

	rr, err := rrule.StrToRRule(entry.RRule)

	if err != nil {
		// invalid rules are always false
		log.Warn().Err(err).Str("rrule", entry.RRule+SCHED_START_TIME_SUFFIX).Msg("invalid rrule")
		return false
	}

	// check if no DTSTART was provided within the rrule
	if (rr.OrigOptions.Dtstart == time.Time{}) {
		// no start time was provided in the rule, use default time and re-run
		return ScheduleEntry{
			RRule:   entry.RRule + SCHED_START_TIME_SUFFIX,
			TimeBeg: entry.TimeBeg,
			TimeEnd: entry.TimeEnd,
		}.matches(allday)
	}

	// check rule instances between the start and end of today
	if len(rr.Between(allday.Beg, allday.End, true)) == 0 {
		return false
	}

	// the day matches, now check the time range
	n := NewTime16(allday.T.Hour(), allday.T.Minute())
	return n >= entry.TimeBeg && n <= entry.TimeEnd
}

func NewAllDay(t time.Time) AllDay {
	// get the beginning and ending of the day at time t
	beg := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	end := beg.Add(24 * time.Hour)
	return AllDay{T: t, Beg: beg, End: end}
}

func Today() AllDay {
	return NewAllDay(time.Now())
}

// System-wide, check if the current schedule includes right now
func Now() bool {
	// get the current time, day beginning, and day end timestamps
	today := Today()

	// iterate over the schedule entries
	for _, entry := range schedule {
		// check if the entry includes the current timestamp
		if entry.matches(&today) {
			return true
		}
	}

	return false
}
