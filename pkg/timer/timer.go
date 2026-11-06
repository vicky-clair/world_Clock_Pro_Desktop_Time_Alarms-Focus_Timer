// Package timer 实现多路并行倒计时器与百分之一秒精度单调秒表功能。
package timer

import (
	"fmt"
	"sync"
	"time"
)

// State 表示倒计时器或秒表的当前运行状态。
// State represents the current status of a timer.
type State string

const (
	StateIdle    State = "idle"    // 空闲未启动
	StateRunning State = "running" // 正在计时中
	StatePaused  State = "paused"  // 已暂停
	StateExpired State = "expired" // 倒计时已归零到期
)

// Timer 表示单个倒计时器实例。
// Timer represents an individual countdown timer.
type Timer struct {
	ID        string        `json:"id"`        // 倒计时唯一 ID
	Name      string        `json:"name"`      // 倒计时名称
	Duration  time.Duration `json:"duration"`  // 初始设定总时长
	Remaining time.Duration `json:"remaining"` // 当前剩余时长
	State     State         `json:"state"`     // 当前运行状态

	deadline time.Time // 运行状态下的到期截止时间点
	pausedAt time.Time // 暂停时刻记录
}

// FormatDuration 将时间段格式化为 HH:MM:SS 或 MM:SS 字符串。
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

// Manager 并发安全地管理多个倒计时器实例。
// Manager manages multiple countdown timers.
type Manager struct {
	mu     sync.RWMutex
	timers []*Timer
}

// NewManager 创建包含三组常用预设（泡茶 3m、番茄专注 25m、短休 5m）的倒计时管理器。
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

// Timers 返回所有倒计时器指针切片的副本。
// Timers returns a slice copy of timers.
func (m *Manager) Timers() []*Timer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*Timer, len(m.timers))
	copy(res, m.timers)
	return res
}

// AddTimer 创建并添加一个新的倒计时器。
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

// DeleteTimer 根据 ID 删除对应的倒计时器。
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

// Start 从当前剩余时长启动倒计时。
// Start starts a timer from its current remaining duration.
func (t *Timer) Start(now time.Time) {
	if t.Remaining <= 0 {
		t.Remaining = t.Duration
	}
	t.deadline = now.Add(t.Remaining)
	t.State = StateRunning
}

// Pause 暂停正在运行的倒计时并冻结剩余时长。
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

// Resume 继续已暂停的倒计时。
// Resume resumes a paused timer.
func (t *Timer) Resume(now time.Time) {
	if t.State != StatePaused {
		return
	}
	t.Start(now)
}

// Reset 将倒计时器重置回初始设定总时长与空闲状态。
// Reset resets a timer back to idle with full duration.
func (t *Timer) Reset() {
	t.Remaining = t.Duration
	t.State = StateIdle
	t.deadline = time.Time{}
}

// Update 根据当前时间刷新剩余时长；若本次刚好倒计时归零到期则返回 true。
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

// ProgressFraction 返回当前倒计时已消耗的进度比例（0.0 ~ 1.0），用于渲染进度条。
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
