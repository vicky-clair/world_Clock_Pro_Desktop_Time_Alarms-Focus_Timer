package ui

import (
	"fmt"

	"worldclock/pkg/tracker"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawTracker 渲染生产力专注计时器（番茄工作法）界面与累计统计数据卡片。
// DrawTracker renders the productivity time tracker and Pomodoro break manager.
func DrawTracker(f *flex.Frame, state *AppState, theme ui.Theme) {
	now := state.Now
	tr := state.Tracker

	status := tr.Status()
	cycles, breaks, totalWork, _ := tr.Stats()
	cfg := tr.Config()

	scrollTracker := f.ScrollArea(IDScrollTracker).Fill()
	scrollTracker.Children(func() {
		f.Column().FillWidth().Pad(24).Gap(20).Items(flex.Center).Children(func() {
			// Title Header
			f.Row().FillWidth().Items(flex.Center).Children(func() {
				f.Text(state.T("tracker.title")).FontSize(22).Bold().Color(theme.Foreground)
				f.Spacer()

				dndLabel := state.T("tracker.dnd_off")
				if cfg.DoNotDisturb {
					dndLabel = state.T("tracker.dnd_on")
				}
				btnDnd := f.Button(IDBtnTrackerDND, dndLabel).Small()
				if cfg.DoNotDisturb {
					btnDnd.Outline()
				} else {
					btnDnd.Secondary()
				}
				if btnDnd.Clicked() {
					tr.SetDND(!cfg.DoNotDisturb)
				}
			})

			// Main Status Card
			statusColor := theme.Primary
			if status == tracker.StatusShortBreak || status == tracker.StatusBigBreak {
				statusColor = theme.Accent
			} else if status == tracker.StatusPaused {
				statusColor = theme.MutedForeground
			}

			f.Column().FillWidth().Pad(28).Background(theme.Card).Radius(16).Border(1, theme.Border).Items(flex.Center).Gap(14).Children(func() {
				statusStr := string(status)
				switch status {
				case tracker.StatusWorking:
					statusStr = state.T("tracker.working")
				case tracker.StatusShortBreak:
					statusStr = state.T("tracker.short_break")
				case tracker.StatusBigBreak:
					statusStr = state.T("tracker.big_break")
				case tracker.StatusPaused:
					statusStr = state.T("tracker.paused")
				case tracker.StatusIdle:
					statusStr = state.T("tracker.idle")
				}
				f.Text(fmt.Sprintf("● %s", statusStr)).FontSize(16).Bold().Color(statusColor)

				// Giant Remaining Countdown
				f.Text(tr.FormatRemaining()).FontSize(64).Bold().Mono().Color(theme.Foreground)

				// Progress bar
				f.Progress(tr.ProgressFraction()).Height(8).FillWidth()

				// Interval Info
				cycleNum := (cycles % cfg.CyclesBeforeBig) + 1
				f.Text(fmt.Sprintf(state.T("tracker.cycle_info"), cycleNum, cfg.CyclesBeforeBig, state.T("tracker.big_break"), cfg.CyclesBeforeBig)).FontSize(13).Muted()

				// Controls
				f.Row().Items(flex.Center).Gap(12).Children(func() {
					switch status {
					case tracker.StatusIdle:
						if f.Button(IDBtnTrackerStart, state.T("tracker.start")).Clicked() {
							tr.Start(now)
						}

					case tracker.StatusWorking:
						if f.Button(IDBtnTrackerPause, state.T("timer.pause")).Outline().Clicked() {
							tr.Pause(now)
						}
						if f.Button(IDBtnTrackerReset, state.T("timer.reset")).Ghost().Clicked() {
							tr.Reset()
						}

					case tracker.StatusPaused:
						if f.Button(IDBtnTrackerResume, state.T("timer.resume")).Clicked() {
							tr.Resume(now)
						}
						if f.Button(IDBtnTrackerReset, state.T("timer.reset")).Ghost().Clicked() {
							tr.Reset()
						}

					case tracker.StatusShortBreak, tracker.StatusBigBreak:
						if f.Button(IDBtnTrackerSkip, state.T("tracker.skip")).Clicked() {
							tr.SkipBreak(now)
						}
						if f.Button(IDBtnTrackerPostpone, state.T("tracker.postpone")).Secondary().Clicked() {
							tr.PostponeBreak(now, 5)
						}
					}
				})
			})

			// Cumulative Stats Section
			f.Row().FillWidth().Gap(16).Children(func() {
				// Card 1: Completed Cycles
				f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Items(flex.Center).Gap(4).Children(func() {
					f.Text(state.T("tracker.stat_cycles")).FontSize(12).Muted()
					f.Text(fmt.Sprintf("%d", cycles)).FontSize(24).Bold().Color(theme.Primary)
				})

				// Card 2: Breaks Taken
				f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Items(flex.Center).Gap(4).Children(func() {
					f.Text(state.T("tracker.stat_breaks")).FontSize(12).Muted()
					f.Text(fmt.Sprintf("%d", breaks)).FontSize(24).Bold().Color(theme.Accent)
				})

				// Card 3: Total Work Time
				f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Items(flex.Center).Gap(4).Children(func() {
					f.Text(state.T("tracker.stat_focus")).FontSize(12).Muted()
					hrs := int(totalWork.Hours())
					mins := int(totalWork.Minutes()) % 60
					f.Text(fmt.Sprintf(state.T("tracker.focus_fmt"), hrs, mins)).FontSize(24).Bold().Color(theme.Foreground)
				})
			})
		})
	})
	DrawScrollbar(f.Canvas(), scrollTracker, state.View.ScrollState(IDScrollTracker), theme)
}
