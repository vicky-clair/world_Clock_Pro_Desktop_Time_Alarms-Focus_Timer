package ui

import (
	"fmt"
	"time"

	"worldclock/pkg/timer"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

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

// DrawTimers renders countdown timers and high-precision stopwatch.
func DrawTimers(f *flex.Frame, state *AppState, theme ui.Theme) {
	now := state.Now
	timers := state.TimerMgr.Timers()
	sw := state.Stopwatch

	scrollTimers := f.ScrollArea(IDScrollTimers).Fill()
	scrollTimers.Children(func() {
		f.Column().FillWidth().Pad(24).Gap(24).Children(func() {
			// SECTION 1: COUNTDOWN TIMERS
			f.Row().FillWidth().Items(flex.Center).Children(func() {
				f.Text(state.T("timer.title")).FontSize(20).Bold().Color(theme.Foreground)
				f.Spacer()
				if f.Button(IDBtnAddTimerCustom, state.T("timer.add_10m")).Small().Outline().Clicked() {
					state.TimerMgr.AddTimer("Quick 10m", 10*time.Minute)
				}
			})

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
						f.Column().Width(240).Gap(8).Children(func() {
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

						if f.Button(TimerDeleteID(t.ID), "✕").Small().Ghost().Clicked() {
							state.TimerMgr.DeleteTimer(t.ID)
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
