// Package tracker 实现生产力专注计时（番茄工作法）状态机，协调专注阶段、短休息、长休息与勿扰模式。
package tracker

import (
	"fmt"
	"sync"
	"time"
)

// Status 表示专注计时器的当前工作阶段状态。
// Status represents the current mode of the time tracker.
type Status string

const (
	StatusIdle       Status = "Idle"        // 空闲未启动
	StatusWorking    Status = "Working"     // 正在专注工作中
	StatusShortBreak Status = "Short Break" // 短暂休息阶段（默认 5 分钟）
	StatusBigBreak   Status = "Big Break"   // 长休息阶段（每 4 个番茄周期后默认 15 分钟）
	StatusPaused     Status = "Paused"      // 已暂停
)

// Config 定义专注工作时长、短休时长、长休时长、长休间隔周期数及勿扰模式开关。
// Config defines the intervals for work and breaks.
type Config struct {
	WorkDuration       time.Duration `json:"workDuration"`       // 单次专注时长（默认 25 分钟）
	ShortBreakDuration time.Duration `json:"shortBreakDuration"` // 短休息时长（默认 5 分钟）
	BigBreakDuration   time.Duration `json:"bigBreakDuration"`   // 长休息时长（默认 15 分钟）
	CyclesBeforeBig    int           `json:"cyclesBeforeBig"`    // 每完成多少个专注周期进入一次长休息（默认 4）
	DoNotDisturb       bool          `json:"doNotDisturb"`       // 勿扰模式（开启后自动跳过休息提醒连续计入专注）
}

// DefaultConfig 返回推荐的标准番茄工作法配置（25m 专注 / 5m 短休 / 15m 长休 / 4 轮周期）。
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

// Tracker 协调工作会话、休息间隔流转以及累计生产力统计指标。
// Tracker coordinates work sessions, break intervals, and productivity metrics.
type Tracker struct {
	mu          sync.RWMutex
	cfg         Config
	status      Status
	prevStatus  Status        // 暂停前的阶段状态
	startedAt   time.Time     // 启动时刻
	deadline    time.Time     // 当前阶段结束截止时刻
	pausedAt    time.Time     // 暂停时刻
	lastTick    time.Time     // 上次累计专注时间的采样时刻
	remaining   time.Duration // 当前阶段剩余时长
	completed   int           // 已完成的专注周期数
	breaksTaken int           // 已执行的休息次数
	totalWork   time.Duration // 累计净专注总时长
}

// NewTracker 创建一个专注计时器实例（可选传入自定义配置）。
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

// Status 返回当前专注计时器所处阶段状态。
// Status returns current tracker status.
func (t *Tracker) Status() Status {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// Config 返回专注计时器配置副本。
// Config returns tracker configuration.
func (t *Tracker) Config() Config {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cfg
}

// SetDND 切换勿扰模式开关。
// SetDND toggles Do Not Disturb.
func (t *Tracker) SetDND(dnd bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cfg.DoNotDisturb = dnd
}

// AdjustWorkDuration 按指定分钟或时长增量精确调整单次专注工作时长（限制在 1 分钟 ~ 180 分钟之间）。
// 若当前处于空闲待机状态（StatusIdle），同步刷新待机剩余时长显示。
func (t *Tracker) AdjustWorkDuration(delta time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	next := t.cfg.WorkDuration + delta
	if next < time.Minute {
		next = time.Minute
	}
	if next > 180*time.Minute {
		next = 180 * time.Minute
	}
	t.cfg.WorkDuration = next
	if t.status == StatusIdle {
		t.remaining = next
	}
}

// Start 启动一个新的专注工作周期。
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

// Pause 暂停当前会话并精确结算已完成的专注工作时长。
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

// Resume 恢复已暂停的会话并重算截止时间。
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

// Reset 将专注计时器重置回空闲初始状态（保留历史统计计数）。
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

// Update 评估当前进度并在阶段到期时执行自动流转（工作 -> 短休/长休 -> 工作）。
// 若刚发生阶段切换则返回 true。
// Update evaluates time progress and performs state transitions.
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

	// 当前阶段到期，结算工作时长并切换至下一阶段
	if t.status == StatusWorking && !t.lastTick.IsZero() && t.deadline.After(t.lastTick) {
		t.totalWork += t.deadline.Sub(t.lastTick)
	}
	t.lastTick = now
	t.remaining = 0

	switch t.status {
	case StatusWorking:
		t.completed++
		if t.cfg.DoNotDisturb {
			// 勿扰模式下不进入休息，直接开启下一轮专注计时
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

// SkipBreak 跳过当前休息阶段，立即进入下一轮专注工作。
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

// PostponeBreak 将当前休息或阶段延长指定的分钟数。
// PostponeBreak extends current break or postpones it by minutes.
func (t *Tracker) PostponeBreak(now time.Time, minutes int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	extra := time.Duration(minutes) * time.Minute
	t.remaining += extra
	t.deadline = now.Add(t.remaining)
}

// Stats 返回累计完成的专注周期数、已休息次数、净专注总时长以及当前阶段剩余时长。
// Stats returns session metrics.
func (t *Tracker) Stats() (completedCycles, breaks int, totalWork time.Duration, remaining time.Duration) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.completed, t.breaksTaken, t.totalWork.Round(time.Second), t.remaining
}

// ProgressFraction 返回当前阶段已进行的进度比例（0.0 ~ 1.0）。
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

// FormatRemaining 将当前阶段剩余时间格式化为 MM:SS 字符串。
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
