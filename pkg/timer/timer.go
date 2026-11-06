package timer

import (
	"fmt"
	"sync"
	"time"
)

// State represents the current status of a timer.
type State string

const (
	StateIdle    State = "idle"
	StateRunning State = "running"
	StatePaused  State = "paused"
	StateExpired State = "expired"
)

// Timer represents an individual countdown timer.
type Timer struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Duration  time.Duration `json:"duration"`
	Remaining time.Duration `json:"remaining"`
	State     State         `json:"state"`

	deadline time.Time
	pausedAt time.Time
}

// FormatDuration formats duration as HH:MM:SS or MM:SS.
func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalSec := int(d.Round(time.Second).Seconds())
	hours := totalSec / 3600
	mins := (totalSec % 3600) / 60
	secs := totalSec % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, mins, secs)
	}
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

// Manager manages multiple countdown timers.
type Manager struct {
	mu     sync.RWMutex
	timers []*Timer
}

// NewManager creates a timer manager with standard presets.
func NewManager() *Manager {
	return &Manager{
		timers: []*Timer{
			{
				ID:        "timer-tea",
				Name:      "Tea Timer",
				Duration:  3 * time.Minute,
				Remaining: 3 * time.Minute,
				State:     StateIdle,
			},
			{
				ID:        "timer-pomodoro",
				Name:      "Pomodoro Focus",
				Duration:  25 * time.Minute,
				Remaining: 25 * time.Minute,
				State:     StateIdle,
			},
			{
				ID:        "timer-break",
				Name:      "Short Break",
				Duration:  5 * time.Minute,
				Remaining: 5 * time.Minute,
				State:     StateIdle,
			},
		},
	}
}

// Timers returns a slice copy of timers.
func (m *Manager) Timers() []*Timer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*Timer, len(m.timers))
	copy(res, m.timers)
	return res
}

// AddTimer creates a new timer with the given name and duration.
func (m *Manager) AddTimer(name string, d time.Duration) *Timer {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := &Timer{
		ID:        fmt.Sprintf("timer-%d", time.Now().UnixNano()),
		Name:      name,
		Duration:  d,
		Remaining: d,
		State:     StateIdle,
	}
	m.timers = append(m.timers, t)
	return t
}

// DeleteTimer deletes a timer by ID.
func (m *Manager) DeleteTimer(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, t := range m.timers {
		if t.ID == id {
			m.timers = append(m.timers[:i], m.timers[i+1:]...)
			return true
		}
	}
	return false
}

// Start starts a timer from its current remaining duration.
func (t *Timer) Start(now time.Time) {
	if t.Remaining <= 0 {
		t.Remaining = t.Duration
	}
	t.deadline = now.Add(t.Remaining)
	t.State = StateRunning
}

// Pause pauses a running timer.
func (t *Timer) Pause(now time.Time) {
	if t.State != StateRunning {
		return
	}
	t.Remaining = t.deadline.Sub(now)
	if t.Remaining < 0 {
		t.Remaining = 0
	}
	t.State = StatePaused
}

// Resume resumes a paused timer.
func (t *Timer) Resume(now time.Time) {
	if t.State != StatePaused {
		return
	}
	t.Start(now)
}

// Reset resets a timer back to idle with full duration.
func (t *Timer) Reset() {
	t.Remaining = t.Duration
	t.State = StateIdle
	t.deadline = time.Time{}
}

// Update updates the remaining time based on current time. Returns true if it just expired.
func (t *Timer) Update(now time.Time) bool {
	if t.State != StateRunning {
		return false
	}
	rem := t.deadline.Sub(now)
	if rem <= 0 {
		t.Remaining = 0
		t.State = StateExpired
		return true
	}
	t.Remaining = rem
	return false
}

// ProgressFraction returns the elapsed fraction between 0.0 and 1.0.
func (t *Timer) ProgressFraction() float32 {
	if t.Duration <= 0 {
		return 0
	}
	elapsed := t.Duration - t.Remaining
	if elapsed < 0 {
		elapsed = 0
	}
	f := float32(elapsed) / float32(t.Duration)
	if f > 1 {
		f = 1
	}
	return f
}
