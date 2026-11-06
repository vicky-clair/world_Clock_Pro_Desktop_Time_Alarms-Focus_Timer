package sleep

import (
	"fmt"
	"sync"
	"time"
)

// State describes the sleep mode session.
type State string

const (
	StateIdle      State = "Idle"
	StateActive    State = "Active"
	StateFading    State = "Fading"
	StateCompleted State = "Completed"
)

// AmbientSound identifies ambient background sound.
type AmbientSound string

const (
	SoundRain   AmbientSound = "Gentle Rain"
	SoundStream AmbientSound = "Forest Stream"
	SoundWaves  AmbientSound = "Ocean Waves"
	SoundWhite  AmbientSound = "White Noise"
	SoundMute   AmbientSound = "Mute (Silent)"
)

// AmbientSounds returns list of available ambient sounds.
func AmbientSounds() []AmbientSound {
	return []AmbientSound{
		SoundRain,
		SoundStream,
		SoundWaves,
		SoundWhite,
		SoundMute,
	}
}

// Session manages sleep timer, ambient sound, and volume fade-out.
type Session struct {
	mu           sync.RWMutex
	state        State
	duration     time.Duration
	fadeDuration time.Duration
	sound        AmbientSound
	startedAt    time.Time
	deadline     time.Time
	volume       float32 // 0.0 to 1.0
}

// NewSession creates a sleep session manager.
func NewSession() *Session {
	return &Session{
		state:        StateIdle,
		duration:     30 * time.Minute,
		fadeDuration: 5 * time.Minute,
		sound:        SoundRain,
		volume:       0.8,
	}
}

// Start initiates the sleep mode.
func (s *Session) Start(now time.Time, duration time.Duration, sound AmbientSound) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if duration <= 0 {
		duration = 30 * time.Minute
	}
	s.duration = duration
	s.sound = sound
	s.state = StateActive
	s.startedAt = now
	s.deadline = now.Add(duration)
	s.volume = 0.8
}

// Stop ends the sleep mode session.
func (s *Session) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = StateIdle
	s.volume = 0.0
	s.deadline = time.Time{}
}

// Update calculates remaining sleep time and volume fade.
func (s *Session) Update(now time.Time) (State, float32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == StateIdle || s.state == StateCompleted {
		return s.state, s.volume
	}

	rem := s.deadline.Sub(now)
	if rem <= 0 {
		s.state = StateCompleted
		s.volume = 0.0
		return s.state, s.volume
	}

	// Calculate fade-out if within fade duration
	if rem <= s.fadeDuration && s.fadeDuration > 0 {
		s.state = StateFading
		fraction := float32(rem) / float32(s.fadeDuration)
		s.volume = 0.8 * fraction
	} else {
		s.state = StateActive
		s.volume = 0.8
	}

	return s.state, s.volume
}

// Remaining returns remaining duration.
func (s *Session) Remaining(now time.Time) time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.state == StateIdle || s.state == StateCompleted {
		return 0
	}
	rem := s.deadline.Sub(now)
	if rem < 0 {
		return 0
	}
	return rem
}

// State returns current session state and sound.
func (s *Session) Info() (State, AmbientSound, float32) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state, s.sound, s.volume
}

// FormatRemaining formats remaining duration as MM:SS.
func (s *Session) FormatRemaining(now time.Time) string {
	rem := s.Remaining(now)
	totalSec := int(rem.Round(time.Second).Seconds())
	return fmt.Sprintf("%02d:%02d", totalSec/60, totalSec%60)
}
