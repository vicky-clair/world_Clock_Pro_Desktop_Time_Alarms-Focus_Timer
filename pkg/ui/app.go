package ui

import (
	"os"
	"time"

	"worldclock/pkg/clock"
	"worldclock/pkg/storage"
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

	cfg.OnWake = func() {
		state.ProcessPendingWake()
	}

	cfg.OnCloseRequest = func(allow func()) {
		state.Persist()
		if state.ForceQuit || state.Store.Get().CloseBehavior == storage.CloseBehaviorExit {
			allow()
			if state.Application != nil {
				state.Application.Quit()
			}
			return
		}
		if state.Window != nil {
			state.Window.Hide()
			return
		}
		allow()
	}

	cfg.OnClosed = func() {
		state.Persist()
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

// BuildStatusIcon generates a crisp 32x32 straight-RGBA clock icon for the system tray.
func BuildStatusIcon() app.StatusIcon {
	const sz = 32
	rgba := make([]byte, sz*sz*4)
	cx, cy := 15.5, 15.5
	rOuter := 13.5
	rInner := 11.0
	for y := range sz {
		for x := range sz {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist2 := dx*dx + dy*dy
			off := (y*sz + x) * 4
			if dist2 <= rOuter*rOuter {
				if dist2 >= rInner*rInner {
					// Olive-gold rim (#8F9779)
					rgba[off+0] = 0x8F
					rgba[off+1] = 0x97
					rgba[off+2] = 0x79
					rgba[off+3] = 0xFF
				} else {
					// Dark clock face (#181F1C)
					rgba[off+0] = 0x18
					rgba[off+1] = 0x1F
					rgba[off+2] = 0x1C
					rgba[off+3] = 0xFF
				}
			}
		}
	}
	// Draw 12-o'clock vertical hand and 3-o'clock horizontal hand in bright white
	for y := 7; y <= 16; y++ {
		for x := 15; x <= 16; x++ {
			off := (y*sz + x) * 4
			rgba[off+0], rgba[off+1], rgba[off+2], rgba[off+3] = 0xFF, 0xFF, 0xFF, 0xFF
		}
	}
	for x := 15; x <= 22; x++ {
		for y := 15; y <= 16; y++ {
			off := (y*sz + x) * 4
			rgba[off+0], rgba[off+1], rgba[off+2], rgba[off+3] = 0xFF, 0xBE, 0x0B, 0xFF
		}
	}
	return app.StatusIcon{
		Size: sz,
		RGBA: rgba,
	}
}
