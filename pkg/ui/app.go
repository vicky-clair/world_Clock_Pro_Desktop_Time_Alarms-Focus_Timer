package ui

import (
	"os"
	"time"

	"worldclock/pkg/clock"
	"worldclock/pkg/timer"
	"worldclock/pkg/wallpaper"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// NewAppConfig creates the Viva app configuration for the World Clock application.
func NewAppConfig(state *AppState) app.Config {
	var inv app.Invalidator
	state.Invalidator = &inv

	cfg := app.Config{
		Tree:        &Tree,
		Title:       "World Clock Pro: Desktop Time, Alarms & Focus Timer",
		Width:       840,
		Height:      580,
		Invalidator: &inv,
		TextInput:   flex.Fields(&state.CitySearchField),
	}

	cfg.Paint = func(canvas *ui.Canvas, size ui.Size) {
		// Update per-frame instant
		now := time.Now()
		state.Now = now
		state.UTC = now.UTC()

		// Advance wallpaper slideshow if deadline is due (O(1) zero-allocation check when not due)
		if state.Wallpaper.AdvanceIfDue(now) {
			state.Persist()
		}

		settings := state.Store.Get()
		theme := BuildTheme(settings.Theme, state.Wallpaper.Config().Enabled, state.Wallpaper.Config().CardOpacity, settings.UIFontColor)

		// Check for alarms to ring
		state.Scheduler.CheckTriggers(now)

		// Advance background timers, tracker, and sleep mode regardless of active tab
		state.Tracker.Update(now)
		state.Sleep.Update(now)
		for _, t := range state.TimerMgr.Timers() {
			t.Update(now)
		}

		// Check if any fast-ticking timer or stopwatch is running
		fastTick := false
		if state.Stopwatch.State() == timer.StateRunning {
			fastTick = true
		}
		for _, t := range state.TimerMgr.Timers() {
			if t.State == timer.StateRunning {
				fastTick = true
				break
			}
		}
		trStatus := state.Tracker.Status()
		if trStatus == "Working" || trStatus == "Short Break" || trStatus == "Big Break" {
			fastTick = true
		}
		// Strictly adhere to Viva idle rule:
		// No continuous polling! Schedule exact redraw only when needed:
		if fastTick {
			canvas.RequestFrameIn(40 * time.Millisecond) // ~25 fps when stopwatch/timer running
		} else {
			delay := clock.NextBoundaryDelay(now, settings.ShowSeconds)
			if wpDelay := state.Wallpaper.NextSlideDelay(now); wpDelay > 0 && wpDelay < delay {
				delay = wpDelay
			}
			canvas.RequestFrameIn(delay)
		}

		// Draw root background
		bounds := ui.Rect{Width: size.Width, Height: size.Height}
		canvas.Fill(bounds, theme.Background)
		if state.Wallpaper.Config().Enabled {
			state.Wallpaper.Draw(canvas, bounds, theme)
		}
		// Draw UI tree
		state.View.Draw(canvas, theme, bounds, func(f *flex.Frame) {
			if state.IsPopout {
				DrawPopoutClock(f, state, theme)
				return
			}

			f.Column().Fill().Children(func() {
				// Global Top Navigation and Ringing Alarm Alert
				DrawHeader(f, state, theme)

				// Active Tab Screen
				f.Column().Fill().Children(func() {
					switch state.ActiveTab {
					case TabClock:
						DrawMainClock(f, state, theme)
					case TabWorld:
						DrawWorldClock(f, state, theme)
					case TabAlarms:
						DrawAlarms(f, state, theme)
					case TabTimers:
						DrawTimers(f, state, theme)
					case TabTracker:
						DrawTracker(f, state, theme)
					case TabSleep:
						DrawSleepMode(f, state, theme)
					case TabNews:
						DrawNews(f, state, theme)
					case TabSettings:
						DrawSettings(f, state, theme)
					case TabPopout:
						DrawPopoutClock(f, state, theme)
					default:
						DrawMainClock(f, state, theme)
					}
				})
			})
		})
	}

	cfg.OnScroll = func(event ui.ScrollEvent) ui.EventResult {
		return state.View.Wheel(event)
	}

	cfg.OnFileDrop = func(event ui.FileDropEvent) ui.EventResult {
		if event.Phase == ui.FileDropDrop && len(event.Paths) > 0 {
			first := event.Paths[0]
			if info, err := os.Stat(first); err == nil && info.IsDir() {
				if count, err := state.Wallpaper.LoadFolder(first); err == nil && count > 0 {
					state.Persist()
					return ui.EventResult{Handled: true, Repaint: true}
				}
			}
			if len(event.Paths) > 1 {
				if state.Wallpaper.LoadImages(event.Paths) > 0 {
					state.Persist()
					return ui.EventResult{Handled: true, Repaint: true}
				}
			}
			if wallpaper.IsSupportedImageExt(first) {
				_ = state.Wallpaper.LoadImage(first)
				state.Persist()
				return ui.EventResult{Handled: true, Repaint: true}
			}
		}
		if event.Phase == ui.FileDropOver {
			return ui.EventResult{Handled: true}
		}
		return ui.EventResult{}
	}

	return cfg
}
