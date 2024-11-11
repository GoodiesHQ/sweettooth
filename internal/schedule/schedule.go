package schedule

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/teambition/rrule-go"
)

// Schedule represents a collection of schedule entries.
type Schedule []ScheduleEntry

// ScheduleGroup maps schedule group types to their respective schedules.
type ScheduleGroup map[ScheduleGroupType]Schedule

// ISO format used in RRULEs
const (
	ISO8601 = "20060102T150405Z"
)

type ScheduleGroupType string

var (
	// Unless otherwise specified RRULE entries should be processed starting Jan 1 1970 according to the local time zone
	SCHED_START_TIME        = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.Local).UTC()
	SCHED_START_TIME_SUFFIX = fmt.Sprintf(";DTSTART=%s", SCHED_START_TIME.Format(ISO8601))
)

// Schedule Groups can be applied to a single node, a group, or an organization. All schedules are inherited.
const (
	SchedGroupNode         ScheduleGroupType = "node"
	SchedGroupGroup        ScheduleGroupType = "group"
	SchedGroupOrganization ScheduleGroupType = "organization"
)

// Object to store the local time (HH:MM) in 16 bits and acquire the current Hour and Minutes
type Time16 uint16

func (t Time16) H() int {
	return int((t >> 8) & 0xff)
}

func (t Time16) M() int {
	return int(t & 0xff)
}

// Create a new time
func NewTime16(h int, m int) Time16 {
	return Time16((h&0xff)<<8 | (m & 0xff))
}

// Struct which takes a timestamp, T, and includes the beginning and ending of that given day
type DayInterval struct {
	T   time.Time // a given arbitrary timestamp
	Beg time.Time // the beginning of the day T is in
	End time.Time // the end of the day T is in (Beg + 24hrs)
}

// An entry of the schedule including an RRule (day) and Beg/End time
type ScheduleEntry struct {
	RRule   string `json:"rrule"`
	TimeBeg Time16 `json:"time_beg"`
	TimeEnd Time16 `json:"time_end"`
}

// Check if a given schedule entry resides within a particular day and time range according to the RRules
func (se ScheduleEntry) Matches(di *DayInterval) bool {
	if se.RRule == "" {
		// empty rules are always false
		return false
	}

	// convert the rule string to a rule object
	rr, err := rrule.StrToRRule(se.RRule)

	if err != nil {
		// invalid rules are always false
		log.Warn().Err(err).Str("rrule", se.RRule+SCHED_START_TIME_SUFFIX).Msg("invalid rrule")
		return false
	}

	// check if no DTSTART was provided within the rrule
	if (rr.OrigOptions.Dtstart == time.Time{}) {
		// no start time was provided in the rule, use default time and re-run
		return ScheduleEntry{
			RRule:   se.RRule + SCHED_START_TIME_SUFFIX,
			TimeBeg: se.TimeBeg,
			TimeEnd: se.TimeEnd,
		}.Matches(di)
	}

	// check rule instances between the start and end of today
	if len(rr.Between(di.Beg, di.End, true)) == 0 {
		return false
	}

	// the day matches, now check the time range
	n := NewTime16(di.T.Hour(), di.T.Minute())
	return n >= se.TimeBeg && n <= se.TimeEnd
}

func NewDayInterval(t time.Time) DayInterval {
	// get the beginning and ending of the day at time t
	beg := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	end := beg.Add(24 * time.Hour)
	return DayInterval{T: t, Beg: beg, End: end}
}

func Now() DayInterval {
	return NewDayInterval(time.Now())
}

func (s *Schedule) Matches(di *DayInterval) bool {
	// iterate over the schedule entries
	for _, entry := range *s {
		// check if the entry includes the current timestamp
		if entry.Matches(di) {
			return true
		}
	}
	return false
}
