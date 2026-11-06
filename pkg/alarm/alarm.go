package alarm

import (
	"fmt"
	"time"
)

// ScheduleType describes the recurrence pattern of an alarm.
type ScheduleType string

const (
	ScheduleOneTime ScheduleType = "one_time"
	ScheduleDaily   ScheduleType = "daily"
	ScheduleWeekly  ScheduleType = "weekly"
)

// Alarm represents a user-configured alarm.
type Alarm struct {
	ID            string       `json:"id"`
	Title         string       `json:"title"`
	Hour          int          `json:"hour"`         // 0-23
	Minute        int          `json:"minute"`       // 0-59
	ScheduleType  ScheduleType `json:"scheduleType"` // one_time, daily, weekly
	DaysOfWeek    [7]bool      `json:"daysOfWeek"`   // index 0: Sunday, 1: Monday ... 6: Saturday
	IsEnabled     bool         `json:"isEnabled"`
	SnoozeMinutes int          `json:"snoozeMinutes"` // default 5 or 10
	Sound         string       `json:"sound"`
}

// NextOccurrence computes the next trigger time for an alarm after the given reference instant.
// If the alarm is disabled or has no valid future occurrence, ok is false.
func (a *Alarm) NextOccurrence(after time.Time) (time.Time, bool) {
	if !a.IsEnabled {
		return time.Time{}, false
	}

	loc := after.Location()
	year, month, day := after.Date()

	// Candidate on the same day
	candidate := time.Date(year, month, day, a.Hour, a.Minute, 0, 0, loc)

	switch a.ScheduleType {
	case ScheduleOneTime:
		if candidate.After(after) {
			return candidate, true
		}
		// If already past today, candidate tomorrow
		nextDay := candidate.AddDate(0, 0, 1)
		return nextDay, true

	case ScheduleDaily:
		if candidate.After(after) {
			return candidate, true
		}
		return candidate.AddDate(0, 0, 1), true

	case ScheduleWeekly:
		// Check next 14 days
		for d := range 14 {
			c := candidate.AddDate(0, 0, d)
			if c.After(after) {
				w := int(c.Weekday())
				if a.DaysOfWeek[w] {
					return c, true
				}
			}
		}
		return time.Time{}, false
	}

	return time.Time{}, false
}

// FormatTime formats the alarm time as HH:MM with optional AM/PM.
func (a *Alarm) FormatTime(is24H bool) string {
	h := a.Hour
	m := a.Minute
	if is24H {
		return fmt.Sprintf("%02d:%02d", h, m)
	}
	period := "AM"
	if h >= 12 {
		period = "PM"
	}
	h12 := h % 12
	if h12 == 0 {
		h12 = 12
	}
	return fmt.Sprintf("%02d:%02d %s", h12, m, period)
}

// RepeatSummary returns a human-readable summary of the repeat rule.
func (a *Alarm) RepeatSummary() string {
	switch a.ScheduleType {
	case ScheduleOneTime:
		return "Once"
	case ScheduleDaily:
		return "Daily"
	case ScheduleWeekly:
		names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
		var active []string
		for i, on := range a.DaysOfWeek {
			if on {
				active = append(active, names[i])
			}
		}
		if len(active) == 7 {
			return "Daily"
		}
		if len(active) == 5 && !a.DaysOfWeek[0] && !a.DaysOfWeek[6] {
			return "Weekdays"
		}
		if len(active) == 2 && a.DaysOfWeek[0] && a.DaysOfWeek[6] {
			return "Weekends"
		}
		if len(active) == 0 {
			return "No days"
		}
		res := ""
		for i, n := range active {
			if i > 0 {
				res += ", "
			}
			res += n
		}
		return res
	}
	return "Custom"
}
