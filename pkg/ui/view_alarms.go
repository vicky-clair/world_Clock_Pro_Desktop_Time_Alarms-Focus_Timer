package ui

import (
	"fmt"

	"worldclock/pkg/alarm"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// localizeAlarmTitle 将内置预设闹钟标题翻译为当前界面语言，用户自定义标题则原样保留。
// localizeAlarmTitle translates built-in default/preset alarm titles into the active UI language.
func localizeAlarmTitle(state *AppState, title string) string {
	switch title {
	case "Morning Wake Up":
		return state.T("alarm.name_morning_wakeup")
	case "Daily Standup":
		return state.T("alarm.name_daily_standup")
	case "Morning Alarm":
		return state.T("alarm.name_morning")
	case "Lunch Standup":
		return state.T("alarm.name_lunch")
	case "Evening Reminder":
		return state.T("alarm.name_evening")
	case "New Alarm":
		return state.T("alarm.name_new")
	default:
		return title
	}
}

// localizeAlarmSound 将内置闹钟铃声名称（经典钟声 / 柔和风铃 / 数字脉冲）翻译为当前界面语言。
// localizeAlarmSound translates built-in alarm sound names into the active UI language.
func localizeAlarmSound(state *AppState, sound string) string {
	switch sound {
	case "Gentle Chime":
		return state.T("alarm.snd_gentle")
	case "Digital Bell":
		return state.T("alarm.snd_digital")
	case "Classic Chime":
		return state.T("alarm.snd_classic")
	default:
		return sound
	}
}

// localizeRepeatSummary 将闹钟重复周期摘要（仅一次 / 每天 / 工作日 / 周末）翻译为当前界面语言。
// localizeRepeatSummary translates alarm repeat summary into the active UI language.
func localizeRepeatSummary(state *AppState, a alarm.Alarm) string {
	sum := a.RepeatSummary()
	switch sum {
	case "Once":
		return state.T("alarm.once")
	case "Daily":
		return state.T("alarm.daily")
	case "Weekdays":
		return state.T("alarm.weekdays")
	case "Weekends":
		return state.T("alarm.weekends")
	default:
		return sum
	}
}

// DrawAlarms 渲染闹钟管理列表与新建闹钟面板。
// DrawAlarms renders the alarm manager and new alarm dialog.
func DrawAlarms(f *flex.Frame, state *AppState, theme ui.Theme) {
	cfg := state.Store.GetPrefs()
	alarms := state.Scheduler.Alarms()

	f.Column().Fill().Pad(24).Gap(16).Children(func() {
		// 顶部标题栏与“+ 新建闹钟”按钮
		f.Row().FillWidth().Items(flex.Center).Gap(12).Children(func() {
			f.Text(state.T("alarm.title")).FontSize(22).Bold().Color(theme.Foreground)
			f.Spacer()
			if f.Button(IDBtnNewAlarmModal, state.T("alarm.new_alarm")).Clicked() {
				state.ShowNewAlarmModal = true
			}
		})

		// 闹钟列表或新建闹钟面板
		if state.ShowNewAlarmModal {
			DrawNewAlarmModal(f, state, theme)
		} else {
			scrollAlarms := f.ScrollArea(IDScrollAlarms).Fill()
			scrollAlarms.Children(func() {
				f.Column().FillWidth().Gap(12).Children(func() {
					for _, alm := range alarms {
						a := alm
						f.Row().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Items(flex.Center).Gap(16).Children(func() {
							// 左侧：闹钟触发时间与重复周期
							f.Column().Width(180).Gap(4).Children(func() {
								state.ApplyDigitsFont(f.Text(a.FormatTime(cfg.TimeFormat24H)).FontSize(26).Bold()).Color(theme.Foreground)
								f.Text(localizeRepeatSummary(state, a)).FontSize(12).Bold().Color(theme.Primary)
							})

							// 中间：闹钟标题、铃声名称与贪睡时长
							f.Column().Width(240).Gap(4).Children(func() {
								f.Text(localizeAlarmTitle(state, a.Title)).FontSize(16).Bold().Color(theme.Foreground)
								f.Text(fmt.Sprintf(state.T("alarm.sound_info"), localizeAlarmSound(state, a.Sound), a.SnoozeMinutes)).FontSize(12).Muted()
							})

							f.Spacer()

							// 试听该闹钟的合成铃声按钮
							if f.Button(AlarmPreviewID(a.ID), state.T("alarm.preview")).Ghost().Small().Clicked() {
								if state.Audio != nil {
									state.Audio.PreviewAlarm(a.Sound)
								}
							}

							// 启用/停用闹钟开关按钮
							toggleLabel := state.T("common.off")
							if a.IsEnabled {
								toggleLabel = state.T("common.on")
							}
							btn := f.Button(AlarmToggleID(a.ID), toggleLabel)
							if a.IsEnabled {
								btn.Outline()
							} else {
								btn.Secondary()
							}
							if btn.Clicked() {
								state.Scheduler.ToggleAlarm(a.ID)
								state.Persist()
							}

							// 删除闹钟按钮
							if f.Button(AlarmDeleteID(a.ID), state.T("alarm.delete")).DestructiveOutline().Small().Clicked() {
								state.Scheduler.DeleteAlarm(a.ID)
								state.Persist()
							}
						})
					}
				})
			})
			DrawScrollbar(f.Canvas(), scrollAlarms, state.View.ScrollState(IDScrollAlarms), theme)
		}
	})
}

// DrawNewAlarmModal 渲染创建新闹钟的交互面板（支持早晨/午休/晚间快捷预设、±1小时/±5分钟/±1分钟精确微调、重复规则与铃声试听）。
// DrawNewAlarmModal renders the dialog to create an alarm.
func DrawNewAlarmModal(f *flex.Frame, state *AppState, theme ui.Theme) {
	f.Column().FillWidth().Pad(20).Background(theme.Card).Border(2, theme.Primary).Radius(10).Gap(14).Children(func() {
		f.Row().Items(flex.Center).Gap(12).Children(func() {
			f.Text(state.T("alarm.create_title")).FontSize(18).Bold().Color(theme.Primary)
			f.Spacer()
			if f.Button(IDBtnCloseNewAlarm, state.T("alarm.cancel")).Small().Outline().Clicked() {
				state.ShowNewAlarmModal = false
			}
		})

		// Time Presets
		f.Text(state.T("alarm.presets")).FontSize(13).Muted()
		f.Row().Items(flex.Center).Gap(8).Children(func() {
			if f.Button(IDBtnPresetMorning, state.T("alarm.morning")).Small().Outline().Clicked() {
				state.NewAlarmHour = 7
				state.NewAlarmMinute = 30
				state.NewAlarmTitle = "Morning Alarm"
			}
			if f.Button(IDBtnPresetNoon, state.T("alarm.noon")).Small().Outline().Clicked() {
				state.NewAlarmHour = 12
				state.NewAlarmMinute = 0
				state.NewAlarmTitle = "Lunch Standup"
			}
			if f.Button(IDBtnPresetEvening, state.T("alarm.evening")).Small().Outline().Clicked() {
				state.NewAlarmHour = 18
				state.NewAlarmMinute = 30
				state.NewAlarmTitle = "Evening Reminder"
			}
		})

		// Custom Hour & Minute Steppers (with exact 24h carry/borrow arithmetic)
		f.Row().Items(flex.Center).Gap(8).Children(func() {
			state.ApplyDigitsFont(f.Text(fmt.Sprintf("%02d:%02d", state.NewAlarmHour, state.NewAlarmMinute)).FontSize(22).Bold()).Color(theme.Foreground)
			if f.Button(IDBtnAlarmHourDown, state.T("alarm.step_h_down")).Small().Secondary().Clicked() {
				state.NewAlarmHour, state.NewAlarmMinute = StepAlarmTime(state.NewAlarmHour, state.NewAlarmMinute, -1, 0)
			}
			if f.Button(IDBtnAlarmHourUp, state.T("alarm.step_h_up")).Small().Secondary().Clicked() {
				state.NewAlarmHour, state.NewAlarmMinute = StepAlarmTime(state.NewAlarmHour, state.NewAlarmMinute, 1, 0)
			}
			if f.Button(IDBtnAlarmMinDown, state.T("alarm.step_m_down")).Small().Secondary().Clicked() {
				state.NewAlarmHour, state.NewAlarmMinute = StepAlarmTime(state.NewAlarmHour, state.NewAlarmMinute, 0, -5)
			}
			if f.Button(IDBtnAlarmMin1Down, state.T("alarm.step_m1_down")).Small().Secondary().Clicked() {
				state.NewAlarmHour, state.NewAlarmMinute = StepAlarmTime(state.NewAlarmHour, state.NewAlarmMinute, 0, -1)
			}
			if f.Button(IDBtnAlarmMin1Up, state.T("alarm.step_m1_up")).Small().Secondary().Clicked() {
				state.NewAlarmHour, state.NewAlarmMinute = StepAlarmTime(state.NewAlarmHour, state.NewAlarmMinute, 0, 1)
			}
			if f.Button(IDBtnAlarmMinUp, state.T("alarm.step_m_up")).Small().Secondary().Clicked() {
				state.NewAlarmHour, state.NewAlarmMinute = StepAlarmTime(state.NewAlarmHour, state.NewAlarmMinute, 0, 5)
			}
		})

		// Recurrence Selector
		f.Text(state.T("alarm.recurrence")).FontSize(13).Muted()
		f.Row().Gap(8).Children(func() {
			onceBtn := f.Button(IDBtnRecurOnce, state.T("alarm.once"))
			if state.NewAlarmSchedule == alarm.ScheduleOneTime {
				onceBtn.Outline()
			} else {
				onceBtn.Secondary()
			}
			if onceBtn.Clicked() {
				state.NewAlarmSchedule = alarm.ScheduleOneTime
			}

			dailyBtn := f.Button(IDBtnRecurDaily, state.T("alarm.daily"))
			if state.NewAlarmSchedule == alarm.ScheduleDaily {
				dailyBtn.Outline()
			} else {
				dailyBtn.Secondary()
			}
			if dailyBtn.Clicked() {
				state.NewAlarmSchedule = alarm.ScheduleDaily
			}

			weekBtn := f.Button(IDBtnRecurWeekdays, state.T("alarm.weekdays"))
			if state.NewAlarmSchedule == alarm.ScheduleWeekly {
				weekBtn.Outline()
			} else {
				weekBtn.Secondary()
			}
			if weekBtn.Clicked() {
				state.NewAlarmSchedule = alarm.ScheduleWeekly
			}
		})

		// Sound Selector & Preview
		if state.NewAlarmSound == "" {
			state.NewAlarmSound = "Classic Chime"
		}
		f.Text(state.T("alarm.select_sound")).FontSize(13).Muted()
		f.Row().Items(flex.Center).Gap(8).Children(func() {
			btnClassic := f.Button(IDBtnAlarmSoundClassic, state.T("alarm.snd_classic")).Small()
			if state.NewAlarmSound == "Classic Chime" {
				btnClassic.Outline()
			} else {
				btnClassic.Secondary()
			}
			if btnClassic.Clicked() {
				state.NewAlarmSound = "Classic Chime"
				if state.Audio != nil {
					state.Audio.PreviewAlarm(state.NewAlarmSound)
				}
			}

			btnGentle := f.Button(IDBtnAlarmSoundGentle, state.T("alarm.snd_gentle")).Small()
			if state.NewAlarmSound == "Gentle Chime" {
				btnGentle.Outline()
			} else {
				btnGentle.Secondary()
			}
			if btnGentle.Clicked() {
				state.NewAlarmSound = "Gentle Chime"
				if state.Audio != nil {
					state.Audio.PreviewAlarm(state.NewAlarmSound)
				}
			}

			btnDigital := f.Button(IDBtnAlarmSoundDigital, state.T("alarm.snd_digital")).Small()
			if state.NewAlarmSound == "Digital Bell" {
				btnDigital.Outline()
			} else {
				btnDigital.Secondary()
			}
			if btnDigital.Clicked() {
				state.NewAlarmSound = "Digital Bell"
				if state.Audio != nil {
					state.Audio.PreviewAlarm(state.NewAlarmSound)
				}
			}

			if f.Button(IDBtnAlarmPreviewSound, state.T("alarm.preview")).Small().Ghost().Clicked() {
				if state.Audio != nil {
					state.Audio.PreviewAlarm(state.NewAlarmSound)
				}
			}
		})

		// Save Button
		timeDisplay := fmt.Sprintf("%02d:%02d", state.NewAlarmHour, state.NewAlarmMinute)
		f.Row().Justify(flex.End).Gap(8).Children(func() {
			saveText := fmt.Sprintf(state.T("alarm.save"), timeDisplay)
			if f.Button(IDBtnSaveNewAlarm, saveText).Clicked() {
				var weekdays [7]bool
				if state.NewAlarmSchedule == alarm.ScheduleWeekly {
					weekdays[1] = true
					weekdays[2] = true
					weekdays[3] = true
					weekdays[4] = true
					weekdays[5] = true
				}
				state.Scheduler.AddAlarm(alarm.Alarm{
					Title:         state.NewAlarmTitle,
					Hour:          state.NewAlarmHour,
					Minute:        state.NewAlarmMinute,
					ScheduleType:  state.NewAlarmSchedule,
					DaysOfWeek:    weekdays,
					IsEnabled:     true,
					SnoozeMinutes: 5,
					Sound:         state.NewAlarmSound,
				})
				state.Persist()
				state.ShowNewAlarmModal = false
			}
		})
	})
}

// StepAlarmTime 按 24 小时制（00:00 ~ 23:59，共 1440 分钟）精确计算闹钟时间增减与跨小时/跨午夜进位借位。
// StepAlarmTime accurately computes 24-hour alarm time adjustments with full hour/midnight carry and borrow.
func StepAlarmTime(hour, minute, deltaHours, deltaMinutes int) (int, int) {
	const dayMins = 24 * 60
	total := (hour*60 + minute + deltaHours*60 + deltaMinutes) % dayMins
	if total < 0 {
		total += dayMins
	}
	return total / 60, total % 60
}
