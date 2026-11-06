package tracker

import (
	"testing"
	"time"
)

func TestTrackerTransitions(t *testing.T) {
	cfg := Config{
		WorkDuration:       20 * time.Minute,
		ShortBreakDuration: 4 * time.Minute,
		BigBreakDuration:   10 * time.Minute,
		CyclesBeforeBig:    2,
		DoNotDisturb:       false,
	}
	tr := NewTracker(&cfg)
	now := time.Now()

	// 1. Start Work
	tr.Start(now)
	if tr.Status() != StatusWorking {
		t.Fatalf("expected Working, got %s", tr.Status())
	}

	// 2. Fast forward 21m -> Should transition to Short Break
	now1 := now.Add(21 * time.Minute)
	trans := tr.Update(now1)
	if !trans || tr.Status() != StatusShortBreak {
		t.Fatalf("expected transition to Short Break, got %s, trans=%v", tr.Status(), trans)
	}

	// 3. Fast forward break end -> Back to Working
	now2 := now1.Add(5 * time.Minute)
	trans = tr.Update(now2)
	if !trans || tr.Status() != StatusWorking {
		t.Fatalf("expected transition back to Working, got %s", tr.Status())
	}

	// 4. Finish 2nd cycle -> Should transition to Big Break (CyclesBeforeBig=2)
	now3 := now2.Add(21 * time.Minute)
	trans = tr.Update(now3)
	if !trans || tr.Status() != StatusBigBreak {
		t.Fatalf("expected Big Break on cycle 2, got %s", tr.Status())
	}

	// 5. Test Skip Break
	tr.SkipBreak(now3)
	if tr.Status() != StatusWorking {
		t.Fatalf("expected Working after SkipBreak, got %s", tr.Status())
	}
}

func TestTrackerPauseResume(t *testing.T) {
	tr := NewTracker(nil)
	now := time.Now()

	tr.Start(now)
	now1 := now.Add(5 * time.Minute)
	tr.Pause(now1)
	if tr.Status() != StatusPaused {
		t.Fatalf("expected Paused, got %s", tr.Status())
	}

	now2 := now1.Add(10 * time.Minute)
	tr.Resume(now2)
	if tr.Status() != StatusWorking {
		t.Fatalf("expected Working after resume, got %s", tr.Status())
	}

	tr.Reset()
	if tr.Status() != StatusIdle {
		t.Fatalf("expected Idle after reset, got %s", tr.Status())
	}

	// Verify wall-clock work duration accuracy across high-frequency updates
	_, _, totalWork, _ := tr.Stats()
	if totalWork != 5*time.Minute {
		t.Fatalf("expected 5m accumulated work time, got %v", totalWork)
	}
}
