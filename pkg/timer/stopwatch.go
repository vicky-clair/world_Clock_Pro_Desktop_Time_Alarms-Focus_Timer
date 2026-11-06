package timer

import (
	"fmt"
	"sync"
	"time"
)

// LapRecord 表示秒表记录的单圈分段成绩。
// LapRecord represents a recorded lap.
type LapRecord struct {
	Index        int           `json:"index"`        // 圈数序号（从 1 开始）
	LapTime      time.Duration `json:"lapTime"`      // 本圈用时
	TotalTime    time.Duration `json:"totalTime"`    // 累计总用时
	LapDisplay   string        `json:"lapDisplay"`   // 本圈用时格式化字符串
	TotalDisplay string        `json:"totalDisplay"` // 累计总用时格式化字符串
}

// Stopwatch 提供高精度单调秒表计时与计圈功能。
// Stopwatch provides high precision monotonic stopwatch timing.
type Stopwatch struct {
	mu           sync.RWMutex
	state        State         // 当前秒表状态
	startTime    time.Time     // 本段启动时刻
	accumulated  time.Duration // 之前暂停段累计的时长
	lastLapTotal time.Duration // 上一圈记录时的累计总时长
	laps         []LapRecord   // 历史计圈列表
}

// NewStopwatch 创建一个处于空闲状态的秒表实例。
// NewStopwatch creates a new stopwatch.
func NewStopwatch() *Stopwatch {
	return &Stopwatch{
		state: StateIdle,
	}
}

// State 返回秒表当前状态。
// State returns current stopwatch state.
func (s *Stopwatch) State() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Start 启动秒表正向计时。
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

// Pause 暂停秒表并累加已走过的时间。
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

// Resume 恢复已暂停的秒表。
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

// Reset 清空秒表累计时间与所有计圈记录，恢复至归零状态。
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

// Elapsed 返回截至给定时刻 now 的秒表总计经过时长。
// Elapsed returns current elapsed duration.
func (s *Stopwatch) Elapsed(now time.Time) time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.state == StateRunning {
		return s.accumulated + now.Sub(s.startTime)
	}
	return s.accumulated
}

// Lap 在当前累计时间点记录一圈分段成绩（Lap Split）。
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

// Laps 返回已记录的所有计圈历史切片副本。
// Laps returns a copy of recorded laps.
func (s *Stopwatch) Laps() []LapRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]LapRecord, len(s.laps))
	copy(res, s.laps)
	return res
}

// FormatPrecision 将时间段格式化为带百分之一秒（1/100s）精度的 MM:SS.cc 或 HH:MM:SS.cc 字符串。
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
