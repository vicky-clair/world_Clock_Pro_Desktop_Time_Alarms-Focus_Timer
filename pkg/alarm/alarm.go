// Package alarm 实现闹钟数据模型、周期性触发时间推算、贪睡（Snooze）与闹钟调度管理。
package alarm

import (
	"fmt"
	"time"
)

// ScheduleType 描述闹钟的重复周期类型。
// ScheduleType describes the recurrence pattern of an alarm.
type ScheduleType string

const (
	ScheduleOneTime ScheduleType = "one_time" // 仅响铃一次（解除后自动关闭）
	ScheduleDaily   ScheduleType = "daily"    // 每天重复
	ScheduleWeekly  ScheduleType = "weekly"   // 按指定的星期几重复
)

// Alarm 表示用户配置的单条闹钟记录。
// Alarm represents a user-configured alarm.
type Alarm struct {
	ID            string       `json:"id"`            // 闹钟唯一标识符
	Title         string       `json:"title"`         // 闹钟标题/标签
	Hour          int          `json:"hour"`          // 小时 (0-23)
	Minute        int          `json:"minute"`        // 分钟 (0-59)
	ScheduleType  ScheduleType `json:"scheduleType"`  // 重复类型：one_time, daily, weekly
	DaysOfWeek    [7]bool      `json:"daysOfWeek"`    // 星期掩码（下标 0 为周日，1 为周一 ... 6 为周六）
	IsEnabled     bool         `json:"isEnabled"`     // 是否启用
	SnoozeMinutes int          `json:"snoozeMinutes"` // 贪睡延后分钟数（默认 5 分钟）
	Sound         string       `json:"sound"`         // 铃声名称
}

// NextOccurrence 计算在给定参考时间 after 之后的下一次触发时刻。
// 若闹钟未启用或无有效后续触发时刻，返回 ok = false。
// NextOccurrence computes the next trigger time for an alarm after the given reference instant.
func (a *Alarm) NextOccurrence(after time.Time) (time.Time, bool) {
	if !a.IsEnabled {
		return time.Time{}, false
	}

	loc := after.Location()
	year, month, day := after.Date()

	// 构造当天的候选触发时刻
	candidate := time.Date(year, month, day, a.Hour, a.Minute, 0, 0, loc)

	switch a.ScheduleType {
	case ScheduleOneTime:
		if candidate.After(after) {
			return candidate, true
		}
		// 若当天时间已过，则顺延至次日同一时刻
		nextDay := candidate.AddDate(0, 0, 1)
		return nextDay, true

	case ScheduleDaily:
		if candidate.After(after) {
			return candidate, true
		}
		return candidate.AddDate(0, 0, 1), true

	case ScheduleWeekly:
		// 向后扫描未来 14 天，查找第一个匹配星期掩码的日期
		for d := range 14 {
			c := candidate.AddDate(0, 0, d)
			if c.After(after) {
				w := int(c.Weekday())
				if a.DaysOfWeek[w] {
					return c, true
				}
			}
		}
		return time.Time{}, false
	}

	return time.Time{}, false
}

// FormatTime 根据 24 小时制或 12 小时制格式化闹钟时间（例如 "07:30" 或 "07:30 AM"）。
// FormatTime formats the alarm time as HH:MM with optional AM/PM.
func (a *Alarm) FormatTime(is24H bool) string {
	h := a.Hour
	m := a.Minute
	if is24H {
		return fmt.Sprintf("%02d:%02d", h, m)
	}
	period := "AM"
	if h >= 12 {
		period = "PM"
	}
	h12 := h % 12
	if h12 == 0 {
		h12 = 12
	}
	return fmt.Sprintf("%02d:%02d %s", h12, m, period)
}

// RepeatSummary 返回可读的重复规则摘要字符串（如 "Once"、"Daily"、"Weekdays"、"Weekends"）。
// RepeatSummary returns a human-readable summary of the repeat rule.
func (a *Alarm) RepeatSummary() string {
	switch a.ScheduleType {
	case ScheduleOneTime:
		return "Once"
	case ScheduleDaily:
		return "Daily"
	case ScheduleWeekly:
		names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
		var active []string
		for i, on := range a.DaysOfWeek {
			if on {
				active = append(active, names[i])
			}
		}
		if len(active) == 7 {
			return "Daily"
		}
		if len(active) == 5 && !a.DaysOfWeek[0] && !a.DaysOfWeek[6] {
			return "Weekdays"
		}
		if len(active) == 2 && a.DaysOfWeek[0] && a.DaysOfWeek[6] {
			return "Weekends"
		}
		if len(active) == 0 {
			return "No days"
		}
		res := ""
		for i, n := range active {
			if i > 0 {
				res += ", "
			}
			res += n
		}
		return res
	}
	return "Custom"
}
