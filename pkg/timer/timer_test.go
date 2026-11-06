package timer

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	if s := FormatDuration(3 * time.Minute); s != "03:00" {
		t.Errorf("expected 03:00, got %s", s)
	}
	if s := FormatDuration(65 * time.Minute); s != "01:05:00" {
		t.Errorf("expected 01:05:00, got %s", s)
	}
	if s := FormatDuration(-5 * time.Second); s != "00:00" {
		t.Errorf("expected 00:00 for negative, got %s", s)
	}
}

func TestTimerLifecycle(t *testing.T) {
	mgr := NewManager()
	if len(mgr.Timers()) < 3 {
		t.Fatalf("expected presets, got %d", len(mgr.Timers()))
	}

	timer := mgr.AddTimer("Work", 10*time.Minute)
	now := time.Now()

	timer.Start(now)
	if timer.State != StateRunning {
		t.Errorf("expected running, got %s", timer.State)
	}

	// 2 minutes later
	now2 := now.Add(2 * time.Minute)
	expired := timer.Update(now2)
	if expired {
		t.Errorf("should not be expired after 2 minutes")
	}
	if timer.Remaining.Round(time.Minute) != 8*time.Minute {
		t.Errorf("expected 8 minutes remaining, got %v", timer.Remaining)
	}

	// Pause
	timer.Pause(now2)
	if timer.State != StatePaused {
		t.Errorf("expected paused, got %s", timer.State)
	}

	// Resume 5 minutes later
	now3 := now2.Add(5 * time.Minute)
	timer.Resume(now3)
	if timer.State != StateRunning {
		t.Errorf("expected running after resume")
	}

	// Fast forward to expiry
	now4 := now3.Add(9 * time.Minute)
	expired = timer.Update(now4)
	if !expired || timer.State != StateExpired {
		t.Errorf("expected timer to expire, state=%s, expired=%v", timer.State, expired)
	}

	// Reset
	timer.Reset()
	if timer.State != StateIdle || timer.Remaining != 10*time.Minute {
		t.Errorf("expected idle after reset, got %s, rem=%v", timer.State, timer.Remaining)
	}

	// Delete
	deleted := mgr.DeleteTimer(timer.ID)
	if !deleted {
		t.Errorf("failed to delete timer")
	}
}

func TestStopwatch(t *testing.T) {
	sw := NewStopwatch()
	now := time.Now()

	sw.Start(now)
	if sw.State() != StateRunning {
		t.Errorf("expected running, got %s", sw.State())
	}

	now1 := now.Add(1500 * time.Millisecond)
	lap1, ok := sw.Lap(now1)
	if !ok || lap1.Index != 1 || lap1.TotalTime < 1500*time.Millisecond {
		t.Errorf("unexpected lap1: %+v", lap1)
	}

	now2 := now1.Add(2000 * time.Millisecond)
	lap2, ok := sw.Lap(now2)
	if !ok || lap2.Index != 2 || lap2.LapTime < 2000*time.Millisecond {
		t.Errorf("unexpected lap2: %+v", lap2)
	}

	sw.Pause(now2)
	if sw.State() != StatePaused {
		t.Errorf("expected paused, got %s", sw.State())
	}

	sw.Reset()
	if sw.State() != StateIdle || len(sw.Laps()) != 0 {
		t.Errorf("expected idle with 0 laps after reset")
	}
}

func TestFormatPrecision(t *testing.T) {
	d := 65430 * time.Millisecond // 1m 5s 430ms
	s := FormatPrecision(d)
	if s != "01:05.43" {
		t.Errorf("expected 01:05.43, got %s", s)
	}
}
