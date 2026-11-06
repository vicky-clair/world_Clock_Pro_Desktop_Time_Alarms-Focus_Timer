package alarm

import (
	"testing"
	"time"
)

func TestAlarmNextOccurrence(t *testing.T) {
	ref := time.Date(2026, 11, 5, 8, 0, 0, 0, time.UTC) // Thursday

	// Daily alarm at 09:30
	daily := &Alarm{
		ID:           "test-daily",
		Title:        "Test Daily",
		Hour:         9,
		Minute:       30,
		ScheduleType: ScheduleDaily,
		IsEnabled:    true,
	}

	next, ok := daily.NextOccurrence(ref)
	if !ok || next.Hour() != 9 || next.Minute() != 30 || next.Day() != 5 {
		t.Errorf("expected daily at 09:30 on same day, got %v", next)
	}

	// Past daily alarm at 07:30 -> should fire next day
	dailyPast := &Alarm{
		ID:           "test-past",
		Title:        "Test Past",
		Hour:         7,
		Minute:       30,
		ScheduleType: ScheduleDaily,
		IsEnabled:    true,
	}
	nextPast, ok := dailyPast.NextOccurrence(ref)
	if !ok || nextPast.Day() != 6 || nextPast.Hour() != 7 {
		t.Errorf("expected daily at 07:30 tomorrow, got %v", nextPast)
	}

	// Weekly alarm for Monday only (ref is Thursday 2026-11-05)
	// Next Monday is 2026-11-09
	var weeklyDays [7]bool
	weeklyDays[1] = true // Monday
	weekly := &Alarm{
		ID:           "test-weekly",
		Title:        "Test Weekly",
		Hour:         10,
		Minute:       0,
		ScheduleType: ScheduleWeekly,
		DaysOfWeek:   weeklyDays,
		IsEnabled:    true,
	}
	nextWeekly, ok := weekly.NextOccurrence(ref)
	if !ok || nextWeekly.Weekday() != time.Monday || nextWeekly.Day() != 9 {
		t.Errorf("expected next Monday Nov 9, got %v", nextWeekly)
	}
}

func TestSchedulerLifecycle(t *testing.T) {
	sched := NewScheduler(nil)

	id := sched.AddAlarm(Alarm{
		Title:        "Test Run",
		Hour:         14,
		Minute:       30,
		ScheduleType: ScheduleOneTime,
		IsEnabled:    true,
	})

	// Check trigger before time
	before := time.Date(2026, 11, 5, 14, 29, 0, 0, time.UTC)
	if r := sched.CheckTriggers(before); r != nil {
		t.Errorf("should not trigger before time")
	}

	// Check trigger at time
	at := time.Date(2026, 11, 5, 14, 30, 0, 0, time.UTC)
	r := sched.CheckTriggers(at)
	if r == nil || r.AlarmID != id {
		t.Fatalf("expected alarm %s to trigger at 14:30", id)
	}

	// Test Snooze
	sched.SnoozeAt(at, 5)
	// 2 minutes later, still snoozed
	midSnooze := at.Add(2 * time.Minute)
	if sched.CheckTriggers(midSnooze) != nil {
		t.Errorf("should be quiet during snooze")
	}

	// After snooze expires
	afterSnooze := at.Add(6 * time.Minute)
	if sched.CheckTriggers(afterSnooze) == nil {
		t.Errorf("should re-trigger after snooze expires")
	}

	// Dismiss: will disable the OneTime alarm
	sched.Dismiss()
	if sched.Ringing() != nil {
		t.Errorf("expected ringing to be nil after dismiss")
	}

	// OneTime alarm should now be disabled
	for _, a := range sched.Alarms() {
		if a.ID == id && a.IsEnabled {
			t.Errorf("expected OneTime alarm to be disabled after dismiss")
		}
	}
	// Delete
	deleted := sched.DeleteAlarm(id)
	if !deleted {
		t.Errorf("expected alarm to be deleted")
	}
}
