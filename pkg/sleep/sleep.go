// Package sleep 管理睡眠助眠模式的倒计时状态机、环境音效选择以及结束前的音量线性渐弱计算。
package sleep

import (
	"fmt"
	"sync"
	"time"
)

// State 描述睡眠模式会话的当前状态。
// State describes the sleep mode session.
type State string

const (
	StateIdle      State = "Idle"      // 空闲未启动
	StateActive    State = "Active"    // 正常播放倒计时中
	StateFading    State = "Fading"    // 末尾音量渐弱阶段
	StateCompleted State = "Completed" // 倒计时已结束
)

// AmbientSound 标识助眠背景环境音类型。
// AmbientSound identifies ambient background sound.
type AmbientSound string

const (
	SoundRain   AmbientSound = "Gentle Rain"   // 轻柔细雨
	SoundStream AmbientSound = "Forest Stream" // 森林溪流
	SoundWaves  AmbientSound = "Ocean Waves"   // 舒缓海浪
	SoundWhite  AmbientSound = "White Noise"   // 纯白噪音
	SoundMute   AmbientSound = "Mute (Silent)" // 静音模式
)

// AmbientSounds 返回所有可选的助眠环境音列表。
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

// Session 管理睡眠计时器、环境音种类以及最后 5 分钟的音量线性衰减。
// Session manages sleep timer, ambient sound, and volume fade-out.
type Session struct {
	mu           sync.RWMutex
	state        State
	duration     time.Duration // 总睡眠时长
	fadeDuration time.Duration // 结束前音量线性渐弱窗口时长（默认 5 分钟）
	sound        AmbientSound  // 当前环境音类型
	startedAt    time.Time     // 启动时刻
	deadline     time.Time     // 结束截止时刻
	volume       float32       // 当前音量比例 (0.0 ~ 1.0)
}

// NewSession 创建一个默认时长 30 分钟、最后 5 分钟渐弱的睡眠会话管理器。
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

// Start 启动睡眠助眠模式会话。
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

// Stop 立即停止睡眠模式并重置为空闲状态。
// Stop ends the sleep mode session.
func (s *Session) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = StateIdle
	s.volume = 0.0
	s.deadline = time.Time{}
}

// Update 根据当前时间更新剩余睡眠时间与音量衰减比例。
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

	// 若进入渐弱窗口（默认最后 5 分钟），按剩余比例线性降低音量
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

// Remaining 返回当前会话的剩余倒计时时长。
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

// Info 返回当前睡眠会话的状态、所选环境音与实时音量。
// Info returns current session state and sound.
func (s *Session) Info() (State, AmbientSound, float32) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state, s.sound, s.volume
}

// FormatRemaining 将剩余睡眠时长格式化为 MM:SS 字符串。
// FormatRemaining formats remaining duration as MM:SS.
func (s *Session) FormatRemaining(now time.Time) string {
	rem := s.Remaining(now)
	totalSec := int(rem.Round(time.Second).Seconds())
	return fmt.Sprintf("%02d:%02d", totalSec/60, totalSec%60)
}
