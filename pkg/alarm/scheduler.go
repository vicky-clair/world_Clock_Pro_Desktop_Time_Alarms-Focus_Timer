package alarm

import (
	"fmt"
	"sync"
	"time"
)

// RingingAlarm 表示当前正在响铃提醒的闹钟状态。
// RingingAlarm represents an actively triggering alarm alert.
type RingingAlarm struct {
	AlarmID     string    // 触发的闹钟 ID
	Title       string    // 闹钟标题
	Sound       string    // 闹钟绑定的铃声名称
	TriggerTime time.Time // 触发时间戳
	SnoozeCount int       // 已贪睡次数
}

// Scheduler 负责闹钟列表的并发安全存储、周期匹配、幂等触发检查与贪睡管理。
// Scheduler handles alarm storage, recurrence scheduling, and firing.
type Scheduler struct {
	mu          sync.RWMutex
	alarms      []Alarm          // 闹钟列表
	lastFired   map[string]int64 // 记录每个闹钟上次触发的 Unix 分钟数，防止同一分钟内重复触发
	ringing     *RingingAlarm    // 当前正在响铃的闹钟（无则为 nil）
	snoozeUntil time.Time        // 贪睡静音截止时间
}

// NewScheduler 创建闹钟调度器；若传入的 initial 为 nil，则初始化两条默认示例闹钟。
// NewScheduler creates a new alarm scheduler with initial alarms or sensible defaults when initial is nil.
func NewScheduler(initial []Alarm) *Scheduler {
	s := &Scheduler{
		alarms:    initial,
		lastFired: make(map[string]int64),
	}
	if s.alarms == nil {
		var weekdays [7]bool
		weekdays[1] = true // 周一 Mon
		weekdays[2] = true // 周二 Tue
		weekdays[3] = true // 周三 Wed
		weekdays[4] = true // 周四 Thu
		weekdays[5] = true // 周五 Fri

		s.alarms = []Alarm{
			{
				ID:            "alarm-morning",
				Title:         "Morning Wake Up",
				Hour:          7,
				Minute:        30,
				ScheduleType:  ScheduleWeekly,
				DaysOfWeek:    weekdays,
				IsEnabled:     true,
				SnoozeMinutes: 5,
				Sound:         "Gentle Chime",
			},
			{
				ID:            "alarm-meeting",
				Title:         "Daily Standup",
				Hour:          9,
				Minute:        45,
				ScheduleType:  ScheduleDaily,
				IsEnabled:     false,
				SnoozeMinutes: 5,
				Sound:         "Digital Bell",
			},
		}
	}
	return s
}

// Alarms 返回所有闹钟的深拷贝切片。
// Alarms returns a copy of all alarms.
func (s *Scheduler) Alarms() []Alarm {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]Alarm, len(s.alarms))
	copy(res, s.alarms)
	return res
}

// AddAlarm 新增一条闹钟并返回其生成的唯一 ID。
// AddAlarm adds an alarm.
func (s *Scheduler) AddAlarm(a Alarm) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a.ID == "" {
		a.ID = fmt.Sprintf("alarm-%d", time.Now().UnixNano())
	}
	if a.SnoozeMinutes <= 0 {
		a.SnoozeMinutes = 5
	}
	s.alarms = append(s.alarms, a)
	return a.ID
}

// UpdateAlarm 根据 ID 更新已有闹钟配置。
// UpdateAlarm updates an existing alarm.
func (s *Scheduler) UpdateAlarm(a Alarm) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.alarms {
		if s.alarms[i].ID == a.ID {
			s.alarms[i] = a
			return true
		}
	}
	return false
}

// DeleteAlarm 根据 ID 删除指定闹钟。
// DeleteAlarm removes an alarm by ID.
func (s *Scheduler) DeleteAlarm(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := -1
	for i, a := range s.alarms {
		if a.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false
	}
	s.alarms = append(s.alarms[:idx], s.alarms[idx+1:]...)
	delete(s.lastFired, id)
	return true
}

// ToggleAlarm 切换指定闹钟的启用/停用状态，并返回切换后的新状态。
// ToggleAlarm toggles an alarm's enabled state.
func (s *Scheduler) ToggleAlarm(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.alarms {
		if s.alarms[i].ID == id {
			s.alarms[i].IsEnabled = !s.alarms[i].IsEnabled
			return s.alarms[i].IsEnabled
		}
	}
	return false
}

// Ringing 返回当前正在响铃的闹钟对象（若处于正常空闲状态则返回 nil）。
// Ringing returns the currently ringing alarm, if any.
func (s *Scheduler) Ringing() *RingingAlarm {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ringing
}

// Dismiss 解除当前正在响铃的闹钟；若该闹钟为一次性闹钟（ScheduleOneTime），则自动将其设为停用。
// Dismiss dismisses the currently ringing alarm.
func (s *Scheduler) Dismiss() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ringing != nil {
		// 一次性闹钟在解除后自动关闭
		for i := range s.alarms {
			if s.alarms[i].ID == s.ringing.AlarmID && s.alarms[i].ScheduleType == ScheduleOneTime {
				s.alarms[i].IsEnabled = false
			}
		}
	}
	s.ringing = nil
	s.snoozeUntil = time.Time{}
}

// SnoozeAt 将当前响铃闹钟延后 minutes 分钟（基于传入的参考时间 now）。
// SnoozeAt delays the currently ringing alarm until now + minutes.
func (s *Scheduler) SnoozeAt(now time.Time, minutes int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ringing == nil {
		return
	}
	if minutes <= 0 {
		minutes = 5
	}
	s.snoozeUntil = now.Add(time.Duration(minutes) * time.Minute)
	s.ringing.SnoozeCount++
}

// Snooze 将当前响铃闹钟从当前系统时间起延后 minutes 分钟。
// Snooze delays the currently ringing alarm by snoozeMinutes from now.
func (s *Scheduler) Snooze(minutes int) {
	s.SnoozeAt(time.Now(), minutes)
}

// CheckTriggers 检查给定时间点 now 是否有闹钟应当触发。
// 若处于贪睡静音期内则返回 nil；贪睡到期或新闹钟命中时返回对应的 *RingingAlarm。
// CheckTriggers checks if any alarm should fire at the given time.
func (s *Scheduler) CheckTriggers(now time.Time) *RingingAlarm {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 若当前已有响铃中的闹钟，检查其是否处于贪睡静音期
	if s.ringing != nil {
		if !s.snoozeUntil.IsZero() {
			if now.Before(s.snoozeUntil) {
				return nil // 仍在贪睡静音期内
			}
			// 贪睡期结束，重新恢复响铃提醒
			s.snoozeUntil = time.Time{}
			return s.ringing
		}
		return s.ringing
	}

	currentMinute := now.Unix() / 60
	h, m, _ := now.Clock()
	weekday := int(now.Weekday())

	for i := range s.alarms {
		alm := &s.alarms[i]
		if !alm.IsEnabled {
			continue
		}
		if alm.Hour != h || alm.Minute != m {
			continue
		}

		// 检查重复周期是否匹配当天
		matches := false
		switch alm.ScheduleType {
		case ScheduleOneTime, ScheduleDaily:
			matches = true
		case ScheduleWeekly:
			matches = alm.DaysOfWeek[weekday]
		}

		if !matches {
			continue
		}

		// 幂等防重：同一分钟内已触发过的闹钟不再重复触发
		if s.lastFired[alm.ID] == currentMinute {
			continue
		}

		s.lastFired[alm.ID] = currentMinute
		s.ringing = &RingingAlarm{
			AlarmID:     alm.ID,
			Title:       alm.Title,
			Sound:       alm.Sound,
			TriggerTime: now,
		}
		return s.ringing
	}

	return nil
}
