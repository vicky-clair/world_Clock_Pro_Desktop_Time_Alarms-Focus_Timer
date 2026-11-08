package ui

import (
	"fmt"
	"strings"
	"time"

	"worldclock/pkg/timer"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// localizeTimerName 将内置倒计时预设名称翻译为当前界面语言；自定义名称原样返回。
// localizeTimerName translates built-in timer preset names into the active UI language.
func localizeTimerName(state *AppState, name string) string {
	switch name {
	case "Tea Timer":
		return state.T("timer.name_tea")
	case "Pomodoro Focus":
		return state.T("timer.name_pomodoro")
	case "Short Break":
		return state.T("timer.name_break")
	case "Quick 10m":
		return state.T("timer.name_quick10m")
	default:
		return name
	}
}

// localizeTimerState 将倒计时运行状态翻译为当前界面语言。
// localizeTimerState translates timer state into the active UI language.
func localizeTimerState(state *AppState, st timer.State) string {
	switch st {
	case timer.StateIdle:
		return state.T("timer.state_idle")
	case timer.StateRunning:
		return state.T("timer.state_running")
	case timer.StatePaused:
		return state.T("timer.state_paused")
	case timer.StateExpired:
		return state.T("timer.state_expired")
	default:
		return string(st)
	}
}

// DrawTimers 渲染多路倒计时器列表（支持新建自定义任务、编辑原有任务、持久化删除）与百分之一秒高精度计圈秒表视图。
// DrawTimers renders countdown timers and high-precision stopwatch.
func DrawTimers(f *flex.Frame, state *AppState, theme ui.Theme) {
	now := state.Now
	timers := state.TimerMgr.Timers()
	sw := state.Stopwatch

	scrollTimers := f.ScrollArea(IDScrollTimers).Fill()
	scrollTimers.Children(func() {
		f.Column().FillWidth().Pad(24).Gap(24).Children(func() {
			// SECTION 1: COUNTDOWN TIMERS HEADER
			f.Row().FillWidth().Items(flex.Center).Gap(8).Children(func() {
				f.Text(state.T("timer.title")).FontSize(20).Bold().Color(theme.Foreground)
				f.Spacer()
				if f.Button(IDBtnOpenTimerModal, state.T("timer.new_custom")).Small().Clicked() {
					state.ShowTimerModal = true
					state.EditingTimerID = ""
					state.TimerNameField.SetText("")
					state.TimerModalHours = 0
					state.TimerModalMinutes = 15
					state.TimerModalSeconds = 0
				}
				if f.Button(IDBtnAddTimerCustom, state.T("timer.add_10m")).Small().Outline().Clicked() {
					state.TimerMgr.AddTimer("Quick 10m", 10*time.Minute)
					state.Persist()
				}
			})

			// 新建 / 编辑倒计时任务面板
			if state.ShowTimerModal {
				DrawTimerModal(f, state, theme)
			}

			f.Column().FillWidth().Gap(12).Children(func() {
				for _, tm := range timers {
					t := tm

					f.Row().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Items(flex.Center).Gap(16).Children(func() {
						f.Column().Width(180).Gap(4).Children(func() {
							f.Text(localizeTimerName(state, t.Name)).FontSize(16).Bold().Color(theme.Foreground)
							stateStr := localizeTimerState(state, t.State)
							color := theme.MutedForeground
							if t.State == timer.StateRunning {
								color = theme.Primary
							} else if t.State == timer.StateExpired {
								color = theme.Danger
							}
							f.Text(fmt.Sprintf("[%s]", stateStr)).FontSize(12).Bold().Color(color)
						})

						// Remaining Display & Progress
						f.Column().Width(220).Gap(8).Children(func() {
							f.Text(timer.FormatDuration(t.Remaining)).FontSize(28).Bold().Mono().Color(theme.Foreground)
							f.Progress(t.ProgressFraction()).Height(6)
						})

						f.Spacer()

						// Controls
						switch t.State {
						case timer.StateIdle:
							if f.Button(TimerStartID(t.ID), state.T("timer.start")).Small().Clicked() {
								t.Start(now)
							}
						case timer.StateRunning:
							if f.Button(TimerPauseID(t.ID), state.T("timer.pause")).Small().Outline().Clicked() {
								t.Pause(now)
							}
						case timer.StatePaused:
							if f.Button(TimerStartID(t.ID), state.T("timer.resume")).Small().Clicked() {
								t.Resume(now)
							}
						case timer.StateExpired:
							f.Text(state.T("timer.times_up")).Bold().Color(theme.Danger)
						}

						if f.Button(TimerResetID(t.ID), state.T("timer.reset")).Small().Ghost().Clicked() {
							t.Reset()
						}

						// 编辑修改原有倒计时任务按钮
						if f.Button(TimerEditID(t.ID), state.T("timer.edit")).Small().Outline().Clicked() {
							state.ShowTimerModal = true
							state.EditingTimerID = t.ID
							state.TimerNameField.SetText(localizeTimerName(state, t.Name))
							totalSec := int(t.Duration.Round(time.Second).Seconds())
							state.TimerModalHours = totalSec / 3600
							state.TimerModalMinutes = (totalSec % 3600) / 60
							state.TimerModalSeconds = totalSec % 60
						}

						// 删除倒计时任务按钮（立即持久化落盘，重启应用不再恢复）
						if f.Button(TimerDeleteID(t.ID), "✕").Small().Ghost().Clicked() {
							state.TimerMgr.DeleteTimer(t.ID)
							if state.EditingTimerID == t.ID {
								state.ShowTimerModal = false
								state.EditingTimerID = ""
							}
							state.Persist()
						}
					})
				}
			})

			f.Divider()

			// SECTION 2: STOPWATCH
			f.Column().FillWidth().Pad(24).Background(theme.Card).Radius(16).Border(1, theme.Border).Items(flex.Center).Gap(16).Children(func() {
				f.Text(state.T("stopwatch.title")).FontSize(20).Bold().Color(theme.Foreground)

				// Huge Elapsed Time Readout
				elapsed := sw.Elapsed(now)
				f.Text(timer.FormatPrecision(elapsed)).FontSize(56).Bold().Mono().Color(theme.Primary)

				// Action buttons
				f.Row().Items(flex.Center).Gap(12).Children(func() {
					if sw.State() == timer.StateRunning {
						if f.Button(IDBtnSwPause, state.T("timer.pause")).Outline().Clicked() {
							sw.Pause(now)
						}
						if f.Button(IDBtnSwLap, state.T("stopwatch.lap")).Secondary().Clicked() {
							sw.Lap(now)
						}
					} else {
						startLabel := state.T("timer.start")
						if sw.State() == timer.StatePaused {
							startLabel = state.T("timer.resume")
						}
						if f.Button(IDBtnSwStart, startLabel).Clicked() {
							if sw.State() == timer.StatePaused {
								sw.Resume(now)
							} else {
								sw.Start(now)
							}
						}
					}

					if f.Button(IDBtnSwReset, state.T("timer.reset")).Ghost().Clicked() {
						sw.Reset()
					}
				})

				// Lap Split History
				laps := sw.Laps()
				if len(laps) > 0 {
					f.Column().FillWidth().PadY(12).Gap(8).Children(func() {
						f.Text(fmt.Sprintf(state.T("stopwatch.laps"), len(laps))).FontSize(14).Bold().Muted()
						f.Scroll(0).Height(140).FillWidth().Children(func() {
							f.Column().FillWidth().Gap(6).Children(func() {
								// show newest laps first
								for i := len(laps) - 1; i >= 0; i-- {
									lap := laps[i]
									f.Row().FillWidth().Pad(8).Background(theme.Muted).Radius(6).Items(flex.Center).Children(func() {
										f.Text(fmt.Sprintf(state.T("stopwatch.lap_num"), lap.Index)).Bold().Width(100)
										f.Text(fmt.Sprintf("+%s", lap.LapDisplay)).Mono().Color(theme.Accent).Width(160)
										f.Spacer()
										f.Text(fmt.Sprintf("%s: %s", state.T("stopwatch.total"), lap.TotalDisplay)).Mono().Bold()
									})
								}
							})
						})
					})
				}
			})
		})
	})
	DrawScrollbar(f.Canvas(), scrollTimers, state.View.ScrollState(IDScrollTimers), theme)
}

// DrawTimerModal 渲染“新建自定义倒计时任务 / 修改原有倒计时任务”交互面板。
func DrawTimerModal(f *flex.Frame, state *AppState, theme ui.Theme) {
	isEditing := state.EditingTimerID != ""
	modalTitle := state.T("timer.create_title")
	if isEditing {
		modalTitle = state.T("timer.edit_title")
	}

	f.Column().FillWidth().Pad(20).Background(theme.Card).Border(2, theme.Primary).Radius(12).Gap(14).Children(func() {
		// Header
		f.Row().FillWidth().Items(flex.Center).Gap(12).Children(func() {
			f.Text(modalTitle).FontSize(18).Bold().Color(theme.Primary)
			f.Spacer()
			if f.Button(IDBtnCloseTimerModal, state.T("timer.cancel")).Small().Outline().Clicked() {
				state.ShowTimerModal = false
				state.EditingTimerID = ""
			}
		})

		// Task Name Input Field
		f.Row().FillWidth().Items(flex.Center).Gap(8).Children(func() {
			nameField := f.Field(&state.TimerNameField, IDInputTimerName, state.T("timer.task_name"), state.T("timer.name_placeholder")).Grow(1)
			if nameField.Submitted() {
				saveTimerFromModal(state)
			}
		})

		// Quick Task Presets (populates both name and duration)
		f.Text(state.T("timer.quick_presets")).FontSize(13).Muted()
		f.Row().Items(flex.Center).Gap(6).Children(func() {
			if f.Button(IDBtnTimerPresetTea, state.T("timer.preset_tea")).Small().Outline().Clicked() {
				state.TimerNameField.SetText(state.T("timer.name_tea"))
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = 0, 3, 0
			}
			if f.Button(IDBtnTimerPresetNoodle, state.T("timer.preset_noodle")).Small().Outline().Clicked() {
				state.TimerNameField.SetText(cleanPresetLabel(state.T("timer.preset_noodle")))
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = 0, 5, 0
			}
			if f.Button(IDBtnTimerPresetNap, state.T("timer.preset_nap")).Small().Outline().Clicked() {
				state.TimerNameField.SetText(cleanPresetLabel(state.T("timer.preset_nap")))
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = 0, 20, 0
			}
			if f.Button(IDBtnTimerPresetPomo, state.T("timer.preset_pomo")).Small().Outline().Clicked() {
				state.TimerNameField.SetText(state.T("timer.name_pomodoro"))
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = 0, 25, 0
			}
			if f.Button(IDBtnTimerPresetRead, state.T("timer.preset_read")).Small().Outline().Clicked() {
				state.TimerNameField.SetText(cleanPresetLabel(state.T("timer.preset_read")))
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = 0, 45, 0
			}
			if f.Button(IDBtnTimerPresetWorkout, state.T("timer.preset_workout")).Small().Outline().Clicked() {
				state.TimerNameField.SetText(cleanPresetLabel(state.T("timer.preset_workout")))
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = 1, 0, 0
			}
		})

		// Custom Duration Steppers (Hours : Minutes : Seconds) with exact total-seconds carry/borrow arithmetic
		f.Text(state.T("timer.duration_label")).FontSize(13).Muted()
		durDisplay := fmt.Sprintf("%02d:%02d:%02d", state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds)
		f.Row().Items(flex.Center).Gap(6).Children(func() {
			f.Text(durDisplay).FontSize(24).Bold().Mono().Color(theme.Foreground)

			if f.Button(IDBtnTimerHrDown, state.T("timer.step_h_down")).Small().Secondary().Clicked() {
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = StepTimerHMS(state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds, -3600)
			}
			if f.Button(IDBtnTimerHrUp, state.T("timer.step_h_up")).Small().Secondary().Clicked() {
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = StepTimerHMS(state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds, 3600)
			}
			if f.Button(IDBtnTimerMin10Down, state.T("timer.step_m10_down")).Small().Secondary().Clicked() {
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = StepTimerHMS(state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds, -600)
			}
			if f.Button(IDBtnTimerMinDown, state.T("timer.step_m1_down")).Small().Secondary().Clicked() {
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = StepTimerHMS(state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds, -60)
			}
			if f.Button(IDBtnTimerMinUp, state.T("timer.step_m1_up")).Small().Secondary().Clicked() {
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = StepTimerHMS(state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds, 60)
			}
			if f.Button(IDBtnTimerMin10Up, state.T("timer.step_m10_up")).Small().Secondary().Clicked() {
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = StepTimerHMS(state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds, 600)
			}
			if f.Button(IDBtnTimerSecDown, state.T("timer.step_s15_down")).Small().Secondary().Clicked() {
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = StepTimerHMS(state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds, -15)
			}
			if f.Button(IDBtnTimerSec1Down, state.T("timer.step_s1_down")).Small().Secondary().Clicked() {
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = StepTimerHMS(state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds, -1)
			}
			if f.Button(IDBtnTimerSec1Up, state.T("timer.step_s1_up")).Small().Secondary().Clicked() {
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = StepTimerHMS(state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds, 1)
			}
			if f.Button(IDBtnTimerSecUp, state.T("timer.step_s15_up")).Small().Secondary().Clicked() {
				state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds = StepTimerHMS(state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds, 15)
			}
		})

		// Save Button
		f.Row().Justify(flex.End).Gap(8).Children(func() {
			saveKey := "timer.save_new"
			if isEditing {
				saveKey = "timer.save_edit"
			}
			saveLabel := fmt.Sprintf(state.T(saveKey), durDisplay)
			if f.Button(IDBtnSaveTimerModal, saveLabel).Clicked() {
				saveTimerFromModal(state)
			}
		})
	})
}

func saveTimerFromModal(state *AppState) {
	name := strings.TrimSpace(state.TimerNameField.Text())
	if name == "" {
		name = state.T("timer.default_custom")
	}
	dur := time.Duration(state.TimerModalHours)*time.Hour +
		time.Duration(state.TimerModalMinutes)*time.Minute +
		time.Duration(state.TimerModalSeconds)*time.Second
	if dur <= 0 {
		dur = time.Minute
	}

	if state.EditingTimerID != "" {
		state.TimerMgr.UpdateTimer(state.EditingTimerID, name, dur)
	} else {
		state.TimerMgr.AddTimer(name, dur)
	}
	state.Persist()
	state.ShowTimerModal = false
	state.EditingTimerID = ""
}

func cleanPresetLabel(label string) string {
	if idx := strings.Index(label, " ("); idx > 0 {
		return strings.TrimSpace(label[:idx])
	}
	return label
}

// StepTimerHMS 按总秒数精确计算倒计时长（时:分:秒）的进位与借位，并严格限制在 [00:00:00, 23:59:59] 范围内。
// StepTimerHMS accurately computes timer duration adjustments with full hour/minute/second carry and borrow.
func StepTimerHMS(hours, minutes, seconds, deltaSeconds int) (int, int, int) {
	const maxSec = 24*3600 - 1 // 23:59:59
	total := hours*3600 + minutes*60 + seconds + deltaSeconds
	if total < 0 {
		total = 0
	}
	if total > maxSec {
		total = maxSec
	}
	return total / 3600, (total % 3600) / 60, total % 60
}
