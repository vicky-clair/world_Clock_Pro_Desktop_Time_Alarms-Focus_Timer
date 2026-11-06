package tracker

import (
	"fmt"
	"sync"
	"time"
)

// Status represents the current mode of the time tracker.
type Status string

const (
	StatusIdle       Status = "Idle"
	StatusWorking    Status = "Working"
	StatusShortBreak Status = "Short Break"
	StatusBigBreak   Status = "Big Break"
	StatusPaused     Status = "Paused"
)

// Config defines the intervals for work and breaks.
type Config struct {
	WorkDuration       time.Duration `json:"workDuration"`       // default 25m
	ShortBreakDuration time.Duration `json:"shortBreakDuration"` // default 5m
	BigBreakDuration   time.Duration `json:"bigBreakDuration"`   // default 15m
	CyclesBeforeBig    int           `json:"cyclesBeforeBig"`    // default 4
	DoNotDisturb       bool          `json:"doNotDisturb"`
}

// DefaultConfig returns recommended initial settings.
func DefaultConfig() Config {
	return Config{
		WorkDuration:       25 * time.Minute,
		ShortBreakDuration: 5 * time.Minute,
		BigBreakDuration:   15 * time.Minute,
		CyclesBeforeBig:    4,
		DoNotDisturb:       false,
	}
}

// Tracker coordinates work sessions, break intervals, and productivity metrics.
type Tracker struct {
	mu          sync.RWMutex
	cfg         Config
	status      Status
	prevStatus  Status // before paused
	startedAt   time.Time
	deadline    time.Time
	pausedAt    time.Time
	lastTick    time.Time
	remaining   time.Duration
	completed   int // completed work intervals
	breaksTaken int
	totalWork   time.Duration
}

// NewTracker creates a new time tracker with optional custom config.
func NewTracker(cfg *Config) *Tracker {
	c := DefaultConfig()
	if cfg != nil {
		c = *cfg
	}
	return &Tracker{
		cfg:       c,
		status:    StatusIdle,
		remaining: c.WorkDuration,
	}
}

// Status returns current tracker status.
func (t *Tracker) Status() Status {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// Config returns tracker configuration.
func (t *Tracker) Config() Config {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cfg
}

// SetDND toggles Do Not Disturb.
func (t *Tracker) SetDND(dnd bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cfg.DoNotDisturb = dnd
}

// Start begins a work session.
func (t *Tracker) Start(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = StatusWorking
	t.remaining = t.cfg.WorkDuration
	t.startedAt = now
	t.lastTick = now
	t.deadline = now.Add(t.remaining)
}

// Pause pauses the current session.
func (t *Tracker) Pause(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.status == StatusIdle || t.status == StatusPaused {
		return
	}
	if t.status == StatusWorking && !t.lastTick.IsZero() && now.After(t.lastTick) {
		end := now
		if !t.deadline.IsZero() && end.After(t.deadline) {
			end = t.deadline
		}
		if end.After(t.lastTick) {
			t.totalWork += end.Sub(t.lastTick)
		}
	}
	t.lastTick = now
	t.remaining = t.deadline.Sub(now)
	if t.remaining < 0 {
		t.remaining = 0
	}
	t.prevStatus = t.status
	t.status = StatusPaused
}

// Resume resumes a paused session.
func (t *Tracker) Resume(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.status != StatusPaused {
		return
	}
	t.status = t.prevStatus
	t.lastTick = now
	t.deadline = now.Add(t.remaining)
}

// Reset resets tracker back to idle.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = StatusIdle
	t.remaining = t.cfg.WorkDuration
	t.startedAt = time.Time{}
	t.lastTick = time.Time{}
	t.deadline = time.Time{}
}

// Update evaluates time progress and performs state transitions.
// Returns true if a transition just occurred (e.g. break started or ended).
func (t *Tracker) Update(now time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.status == StatusIdle || t.status == StatusPaused {
		return false
	}

	rem := t.deadline.Sub(now)
	if rem > 0 {
		if t.status == StatusWorking && !t.lastTick.IsZero() && now.After(t.lastTick) {
			t.totalWork += now.Sub(t.lastTick)
		}
		t.lastTick = now
		t.remaining = rem
		return false
	}

	// Session expired -> transition
	if t.status == StatusWorking && !t.lastTick.IsZero() && t.deadline.After(t.lastTick) {
		t.totalWork += t.deadline.Sub(t.lastTick)
	}
	t.lastTick = now
	t.remaining = 0

	switch t.status {
	case StatusWorking:
		t.completed++
		if t.cfg.DoNotDisturb {
			// In DND, continue working
			t.remaining = t.cfg.WorkDuration
			t.deadline = now.Add(t.remaining)
			return true
		}

		if t.completed%t.cfg.CyclesBeforeBig == 0 {
			t.status = StatusBigBreak
			t.remaining = t.cfg.BigBreakDuration
		} else {
			t.status = StatusShortBreak
			t.remaining = t.cfg.ShortBreakDuration
		}
		t.deadline = now.Add(t.remaining)
		return true

	case StatusShortBreak, StatusBigBreak:
		t.breaksTaken++
		t.status = StatusWorking
		t.remaining = t.cfg.WorkDuration
		t.deadline = now.Add(t.remaining)
		return true
	}

	return false
}

// SkipBreak ends current break and starts next work session.
func (t *Tracker) SkipBreak(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.status != StatusShortBreak && t.status != StatusBigBreak {
		return
	}
	t.status = StatusWorking
	t.remaining = t.cfg.WorkDuration
	t.lastTick = now
	t.deadline = now.Add(t.remaining)
}

// PostponeBreak extends current break or postpones it by minutes.
func (t *Tracker) PostponeBreak(now time.Time, minutes int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	extra := time.Duration(minutes) * time.Minute
	t.remaining += extra
	t.deadline = now.Add(t.remaining)
}

// Stats returns session metrics.
func (t *Tracker) Stats() (completedCycles, breaks int, totalWork time.Duration, remaining time.Duration) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.completed, t.breaksTaken, t.totalWork.Round(time.Second), t.remaining
}

// ProgressFraction returns fraction of current session elapsed (0.0 to 1.0).
func (t *Tracker) ProgressFraction() float32 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var total time.Duration
	switch t.status {
	case StatusWorking:
		total = t.cfg.WorkDuration
	case StatusShortBreak:
		total = t.cfg.ShortBreakDuration
	case StatusBigBreak:
		total = t.cfg.BigBreakDuration
	default:
		return 0
	}

	if total <= 0 {
		return 0
	}
	elapsed := total - t.remaining
	if elapsed < 0 {
		elapsed = 0
	}
	f := float32(elapsed) / float32(total)
	if f > 1 {
		f = 1
	}
	return f
}

// FormatRemaining formats remaining session time as MM:SS.
func (t *Tracker) FormatRemaining() string {
	t.mu.RLock()
	rem := t.remaining
	t.mu.RUnlock()

	if rem < 0 {
		rem = 0
	}
	totalSec := int(rem.Round(time.Second).Seconds())
	return fmt.Sprintf("%02d:%02d", totalSec/60, totalSec%60)
}
