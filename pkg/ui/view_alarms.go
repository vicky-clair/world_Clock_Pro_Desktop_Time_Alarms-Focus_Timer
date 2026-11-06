package ui

import (
	"fmt"

	"worldclock/pkg/alarm"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

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
	cfg := state.Store.Get()
	alarms := state.Scheduler.Alarms()

	f.Column().Fill().Pad(24).Gap(16).Children(func() {
		// Action Bar
		f.Row().FillWidth().Items(flex.Center).Gap(12).Children(func() {
			f.Text(state.T("alarm.title")).FontSize(22).Bold().Color(theme.Foreground)
			f.Spacer()
			if f.Button(IDBtnNewAlarmModal, state.T("alarm.new_alarm")).Clicked() {
				state.ShowNewAlarmModal = true
			}
		})

		// Alarm Cards
		if state.ShowNewAlarmModal {
			DrawNewAlarmModal(f, state, theme)
		} else {
			scrollAlarms := f.ScrollArea(IDScrollAlarms).Fill()
			scrollAlarms.Children(func() {
				f.Column().FillWidth().Gap(12).Children(func() {
					for _, alm := range alarms {
						a := alm
						f.Row().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Items(flex.Center).Gap(16).Children(func() {
							// Left: Time & Repeat
							f.Column().Width(180).Gap(4).Children(func() {
								f.Text(a.FormatTime(cfg.TimeFormat24H)).FontSize(26).Bold().Mono().Color(theme.Foreground)
								f.Text(localizeRepeatSummary(state, a)).FontSize(12).Bold().Color(theme.Primary)
							})

							// Middle: Title & Sound
							f.Column().Width(240).Gap(4).Children(func() {
								f.Text(localizeAlarmTitle(state, a.Title)).FontSize(16).Bold().Color(theme.Foreground)
								f.Text(fmt.Sprintf(state.T("alarm.sound_info"), localizeAlarmSound(state, a.Sound), a.SnoozeMinutes)).FontSize(12).Muted()
							})

							f.Spacer()

							// Toggle Enable/Disable
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

							// Delete Button
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

		// Custom Hour & Minute Steppers
		f.Row().Items(flex.Center).Gap(8).Children(func() {
			f.Text(fmt.Sprintf("%02d:%02d", state.NewAlarmHour, state.NewAlarmMinute)).FontSize(22).Bold().Mono().Color(theme.Foreground)
			if f.Button(IDBtnAlarmHourDown, state.T("alarm.step_h_down")).Small().Secondary().Clicked() {
				state.NewAlarmHour = (state.NewAlarmHour + 23) % 24
			}
			if f.Button(IDBtnAlarmHourUp, state.T("alarm.step_h_up")).Small().Secondary().Clicked() {
				state.NewAlarmHour = (state.NewAlarmHour + 1) % 24
			}
			if f.Button(IDBtnAlarmMinDown, state.T("alarm.step_m_down")).Small().Secondary().Clicked() {
				state.NewAlarmMinute = (state.NewAlarmMinute + 55) % 60
			}
			if f.Button(IDBtnAlarmMinUp, state.T("alarm.step_m_up")).Small().Secondary().Clicked() {
				state.NewAlarmMinute = (state.NewAlarmMinute + 5) % 60
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
					Sound:         "Classic Chime",
				})
				state.Persist()
				state.ShowNewAlarmModal = false
			}
		})
	})
}
