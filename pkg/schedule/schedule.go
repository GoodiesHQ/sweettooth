package schedule

import (
	"reflect"
	"strings"
	"time"
)

type Schedule []ScheduleEntry

var schedule Schedule

// model HH:MM 24-hour time format
type ScheduleTime struct {
	H uint8 `json:"hour"`
	M uint8 `json:"minute"`
}

func (st ScheduleTime) Value() uint16 {
	return uint16(st.H)<<8 | uint16(st.M)
}

type ScheduleEntry struct {
	Weeks []string     `json:"weeks"`
	Days  []string     `json:"days"`
	Start ScheduleTime `json:"start"`
	End   ScheduleTime `json:"end"`
}

func (entry ScheduleEntry) Normalized() ScheduleEntry {
	weeksNormal := make([]string, len(entry.Weeks))
	for i, week := range entry.Weeks {
		weeksNormal[i] = strings.ToUpper(week)
	}

	daysNormal := make([]string, len(entry.Days))
	for i, day := range entry.Days {
		daysNormal[i] = strings.ToUpper(day)
	}

	return ScheduleEntry{
		Weeks: weeksNormal,
		Days:  daysNormal,
		Start: entry.Start,
		End:   entry.End,
	}
}

func (entry ScheduleEntry) Equals(other ScheduleEntry) bool {
	return reflect.DeepEqual(entry, other)
}

func (sched Schedule) Contains(entry ScheduleEntry) bool {
	for _, e := range sched {
		if e.Equals(entry) {
			return true
		}
	}
	return false
}

func Bootstrap() error {
	SetSchedule(CreateSchedule(nil))
	return nil
}

func SetSchedule(sched Schedule) {
	schedule = sched
}

func CreateSchedule(entries []ScheduleEntry) Schedule {
	return Schedule(entries)
}

func AddEntry(entry ScheduleEntry) {
	entry = entry.Normalized()
	for _, existing := range schedule {
		if entry.Equals(existing) {
			return
		}
	}
	schedule = append(schedule, entry)
}

func contains(haystack []string, needle string) bool {
	needle = strings.ToUpper(needle)
	for _, check := range haystack {
		if needle == strings.ToUpper(check) {
			return true
		}
	}
	return false
}

func Now() bool {
	now := time.Now()

	nowDay := strings.ToUpper(now.Weekday().String()[:3])
	nowTime := ScheduleTime{H: uint8(now.Hour()), M: uint8(now.Minute())}.Value()

	for _, entry := range schedule {
		// check if the current day
		if !contains(entry.Days, nowDay) {
			continue
		}

		if (nowTime < entry.Start.Value()) || (nowTime > entry.End.Value()) {
			continue
		}

		return true
	}

	return false
}
