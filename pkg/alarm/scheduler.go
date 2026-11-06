package alarm

import (
	"fmt"
	"sync"
	"time"
)

// RingingAlarm represents an actively triggering alarm alert.
type RingingAlarm struct {
	AlarmID     string
	Title       string
	TriggerTime time.Time
	SnoozeCount int
}

// Scheduler handles alarm storage, recurrence scheduling, and firing.
type Scheduler struct {
	mu          sync.RWMutex
	alarms      []Alarm
	lastFired   map[string]int64 // alarmID -> unix minute to prevent re-firing in same minute
	ringing     *RingingAlarm
	snoozeUntil time.Time
}

// NewScheduler creates a new alarm scheduler with initial alarms or sensible defaults when initial is nil.
func NewScheduler(initial []Alarm) *Scheduler {
	s := &Scheduler{
		alarms:    initial,
		lastFired: make(map[string]int64),
	}
	if s.alarms == nil {
		var weekdays [7]bool
		weekdays[1] = true // Mon
		weekdays[2] = true // Tue
		weekdays[3] = true // Wed
		weekdays[4] = true // Thu
		weekdays[5] = true // Fri

		s.alarms = []Alarm{
			{
				ID:            "alarm-morning",
				Title:         "Morning Wake Up",
				Hour:          7,
				Minute:        30,
				ScheduleType:  ScheduleWeekly,
				DaysOfWeek:    weekdays,
				IsEnabled:     true,
				SnoozeMinutes: 5,
				Sound:         "Gentle Chime",
			},
			{
				ID:            "alarm-meeting",
				Title:         "Daily Standup",
				Hour:          9,
				Minute:        45,
				ScheduleType:  ScheduleDaily,
				IsEnabled:     false,
				SnoozeMinutes: 5,
				Sound:         "Digital Bell",
			},
		}
	}
	return s
}

// Alarms returns a copy of all alarms.
func (s *Scheduler) Alarms() []Alarm {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]Alarm, len(s.alarms))
	copy(res, s.alarms)
	return res
}

// AddAlarm adds an alarm.
func (s *Scheduler) AddAlarm(a Alarm) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a.ID == "" {
		a.ID = fmt.Sprintf("alarm-%d", time.Now().UnixNano())
	}
	if a.SnoozeMinutes <= 0 {
		a.SnoozeMinutes = 5
	}
	s.alarms = append(s.alarms, a)
	return a.ID
}

// UpdateAlarm updates an existing alarm.
func (s *Scheduler) UpdateAlarm(a Alarm) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.alarms {
		if s.alarms[i].ID == a.ID {
			s.alarms[i] = a
			return true
		}
	}
	return false
}

// DeleteAlarm removes an alarm by ID.
func (s *Scheduler) DeleteAlarm(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := -1
	for i, a := range s.alarms {
		if a.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false
	}
	s.alarms = append(s.alarms[:idx], s.alarms[idx+1:]...)
	delete(s.lastFired, id)
	return true
}

// ToggleAlarm toggles an alarm's enabled state.
func (s *Scheduler) ToggleAlarm(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.alarms {
		if s.alarms[i].ID == id {
			s.alarms[i].IsEnabled = !s.alarms[i].IsEnabled
			return s.alarms[i].IsEnabled
		}
	}
	return false
}

// Ringing returns the currently ringing alarm, if any.
func (s *Scheduler) Ringing() *RingingAlarm {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ringing
}

// Dismiss dismisses the currently ringing alarm.
func (s *Scheduler) Dismiss() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ringing != nil {
		// If OneTime alarm, disable it after dismiss
		for i := range s.alarms {
			if s.alarms[i].ID == s.ringing.AlarmID && s.alarms[i].ScheduleType == ScheduleOneTime {
				s.alarms[i].IsEnabled = false
			}
		}
	}
	s.ringing = nil
	s.snoozeUntil = time.Time{}
}

// Snooze delays the currently ringing alarm by snoozeMinutes.
// SnoozeAt delays the currently ringing alarm until now + minutes.
func (s *Scheduler) SnoozeAt(now time.Time, minutes int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ringing == nil {
		return
	}
	if minutes <= 0 {
		minutes = 5
	}
	s.snoozeUntil = now.Add(time.Duration(minutes) * time.Minute)
	s.ringing.SnoozeCount++
}

// Snooze delays the currently ringing alarm by snoozeMinutes from now.
func (s *Scheduler) Snooze(minutes int) {
	s.SnoozeAt(time.Now(), minutes)
}

// CheckTriggers checks if any alarm should fire at the given time.
// Returns the ringing alarm if newly triggered or still active.
func (s *Scheduler) CheckTriggers(now time.Time) *RingingAlarm {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If currently snoozed, check if snooze expired
	if s.ringing != nil {
		if !s.snoozeUntil.IsZero() {
			if now.Before(s.snoozeUntil) {
				return nil // still in snooze quiet period
			}
			// Snooze expired, re-trigger
			s.snoozeUntil = time.Time{}
			return s.ringing
		}
		return s.ringing
	}

	currentMinute := now.Unix() / 60
	h, m, _ := now.Clock()
	weekday := int(now.Weekday())

	for i := range s.alarms {
		alm := &s.alarms[i]
		if !alm.IsEnabled {
			continue
		}
		if alm.Hour != h || alm.Minute != m {
			continue
		}

		// Check repeat schedule match
		matches := false
		switch alm.ScheduleType {
		case ScheduleOneTime, ScheduleDaily:
			matches = true
		case ScheduleWeekly:
			matches = alm.DaysOfWeek[weekday]
		}

		if !matches {
			continue
		}

		// Idempotency: don't fire twice in the same minute
		if s.lastFired[alm.ID] == currentMinute {
			continue
		}

		s.lastFired[alm.ID] = currentMinute
		s.ringing = &RingingAlarm{
			AlarmID:     alm.ID,
			Title:       alm.Title,
			TriggerTime: now,
		}
		return s.ringing
	}

	return nil
}
