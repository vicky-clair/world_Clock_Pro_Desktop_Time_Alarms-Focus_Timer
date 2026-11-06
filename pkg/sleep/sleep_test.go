package sleep

import (
	"testing"
	"time"
)

func TestSleepSessionLifecycle(t *testing.T) {
	s := NewSession()
	now := time.Now()

	s.Start(now, 30*time.Minute, SoundWaves)
	state, sound, vol := s.Info()
	if state != StateActive || sound != SoundWaves || vol <= 0 {
		t.Fatalf("unexpected state after start: state=%s, sound=%s, vol=%f", state, sound, vol)
	}

	// Mid-session (10m in) -> should be active and full volume
	now1 := now.Add(10 * time.Minute)
	state1, vol1 := s.Update(now1)
	if state1 != StateActive || vol1 < 0.7 {
		t.Fatalf("expected Active at 10m, got %s, vol=%f", state1, vol1)
	}

	// Entering fade window (27m in, 3m remaining out of 5m fade) -> should be Fading
	now2 := now.Add(27 * time.Minute)
	state2, vol2 := s.Update(now2)
	if state2 != StateFading || vol2 >= 0.8 {
		t.Fatalf("expected Fading, got %s, vol=%f", state2, vol2)
	}

	// Expired (31m in) -> should be Completed and 0 volume
	now3 := now.Add(31 * time.Minute)
	state3, vol3 := s.Update(now3)
	if state3 != StateCompleted || vol3 != 0.0 {
		t.Fatalf("expected Completed, got %s, vol=%f", state3, vol3)
	}

	// Stop resets back to Idle
	s.Stop()
	state4, _, _ := s.Info()
	if state4 != StateIdle {
		t.Fatalf("expected Idle after stop, got %s", state4)
	}
}
