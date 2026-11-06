package timer

import (
	"fmt"
	"sync"
	"time"
)

// LapRecord represents a recorded lap.
type LapRecord struct {
	Index        int           `json:"index"`
	LapTime      time.Duration `json:"lapTime"`
	TotalTime    time.Duration `json:"totalTime"`
	LapDisplay   string        `json:"lapDisplay"`
	TotalDisplay string        `json:"totalDisplay"`
}

// Stopwatch provides high precision monotonic stopwatch timing.
type Stopwatch struct {
	mu           sync.RWMutex
	state        State
	startTime    time.Time
	accumulated  time.Duration
	lastLapTotal time.Duration
	laps         []LapRecord
}

// NewStopwatch creates a new stopwatch.
func NewStopwatch() *Stopwatch {
	return &Stopwatch{
		state: StateIdle,
	}
}

// State returns current stopwatch state.
func (s *Stopwatch) State() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Start begins stopwatch counting.
func (s *Stopwatch) Start(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == StateRunning {
		return
	}
	s.startTime = now
	s.state = StateRunning
}

// Pause pauses the stopwatch.
func (s *Stopwatch) Pause(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != StateRunning {
		return
	}
	s.accumulated += now.Sub(s.startTime)
	s.state = StatePaused
}

// Resume resumes a paused stopwatch.
func (s *Stopwatch) Resume(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != StatePaused {
		return
	}
	s.startTime = now
	s.state = StateRunning
}

// Reset clears the stopwatch back to zero.
func (s *Stopwatch) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = StateIdle
	s.accumulated = 0
	s.lastLapTotal = 0
	s.laps = nil
	s.startTime = time.Time{}
}

// Elapsed returns current elapsed duration.
func (s *Stopwatch) Elapsed(now time.Time) time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.state == StateRunning {
		return s.accumulated + now.Sub(s.startTime)
	}
	return s.accumulated
}

// Lap records a lap split at current elapsed time.
func (s *Stopwatch) Lap(now time.Time) (LapRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != StateRunning {
		return LapRecord{}, false
	}
	total := s.accumulated + now.Sub(s.startTime)
	lapDelta := total - s.lastLapTotal
	if lapDelta < 0 {
		lapDelta = 0
	}
	s.lastLapTotal = total

	rec := LapRecord{
		Index:        len(s.laps) + 1,
		LapTime:      lapDelta,
		TotalTime:    total,
		LapDisplay:   FormatPrecision(lapDelta),
		TotalDisplay: FormatPrecision(total),
	}
	s.laps = append(s.laps, rec)
	return rec, true
}

// Laps returns a copy of recorded laps.
func (s *Stopwatch) Laps() []LapRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]LapRecord, len(s.laps))
	copy(res, s.laps)
	return res
}

// FormatPrecision formats a duration into MM:SS.cc (centiseconds, 1/100s).
func FormatPrecision(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalCentis := int(d.Milliseconds() / 10)
	centis := totalCentis % 100
	totalSec := totalCentis / 100
	hours := totalSec / 3600
	mins := (totalSec % 3600) / 60
	secs := totalSec % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d.%02d", hours, mins, secs, centis)
	}
	return fmt.Sprintf("%02d:%02d.%02d", mins, secs, centis)
}
