package ui

import (
	"fmt"

	"worldclock/pkg/clock"
	"worldclock/pkg/i18n"
	"worldclock/pkg/storage"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawPopoutClock renders the compact popout clock widget.
func DrawPopoutClock(f *flex.Frame, state *AppState, theme ui.Theme) {
	cfg := state.Store.Get()
	now := state.Now
	timeStr, period := clock.FormatTime(now, cfg.TimeFormat24H, cfg.ShowSeconds)
	dateStr := i18n.FormatDateLocalized(state.Lang(), int(now.Month()), now.Day(), now.Year(), int(now.Weekday()), cfg.ShowFullDate)

	f.Column().Fill().Pad(16).Background(theme.Background).Items(flex.Center).Gap(12).Children(func() {
		// Header controls
		f.Row().FillWidth().Items(flex.Center).Children(func() {
			pinLabel := state.T("popout.pin_off")
			if cfg.Topmost {
				pinLabel = state.T("popout.pin_on")
			}
			if f.Button(IDBtnPinPopout, pinLabel).Small().Ghost().Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.Topmost = !s.Topmost
				})
			}

			f.Spacer()

			if f.Button(IDBtnExitPopout, state.T("nav.restore")).Small().Outline().Clicked() {
				state.IsPopout = false
				state.ActiveTab = TabClock
				state.Persist()
			}
		})

		f.Spacer()

		// Clock representation
		if cfg.ClockStyle == string(clock.StyleAnalog) {
			drawAnalogFace(f, now, theme)
			f.Row().Items(flex.Baseline).Gap(6).Children(func() {
				f.Text(timeStr).FontSize(24).Bold().Mono()
				if period != "" {
					f.Text(period).FontSize(14).Bold().Color(theme.Primary)
				}
			})
		} else {
			f.Column().Pad(20).Background(theme.Card).Radius(10).Border(1, theme.Border).Items(flex.Center).Gap(8).Children(func() {
				f.Row().Items(flex.Baseline).Gap(8).Children(func() {
					f.Text(timeStr).FontSize(44).Bold().Mono().Color(state.ClockColor(theme))
					if period != "" {
						f.Text(period).FontSize(18).Bold().Color(theme.Primary)
					}
				})
				f.Text(dateStr).FontSize(14).Color(theme.MutedForeground)
			})
		}

		f.Spacer()

		// Mini footer
		f.Row().Items(flex.Center).Gap(8).Children(func() {
			formatLabel := state.T("clock.12h")
			if cfg.TimeFormat24H {
				formatLabel = state.T("clock.24h")
			}
			if f.Button(IDBtnToggle24H, formatLabel).Small().Ghost().Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.TimeFormat24H = !s.TimeFormat24H
				})
			}

			secLabel := state.T("clock.seconds_off")
			if cfg.ShowSeconds {
				secLabel = state.T("clock.seconds_on")
			}
			if f.Button(IDBtnToggleSeconds, secLabel).Small().Ghost().Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.ShowSeconds = !s.ShowSeconds
				})
			}

			styleLabel := state.T("clock.style_digital")
			if cfg.ClockStyle == string(clock.StyleAnalog) {
				styleLabel = state.T("clock.style_analog")
			}
			if f.Button(IDBtnToggleStyle, styleLabel).Small().Ghost().Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					if s.ClockStyle == string(clock.StyleAnalog) {
						s.ClockStyle = string(clock.StyleDigital)
					} else {
						s.ClockStyle = string(clock.StyleAnalog)
					}
				})
			}
			profLabel := cfg.ActiveProfile
			switch cfg.ActiveProfile {
			case "Default", "":
				profLabel = state.T("settings.profile_default")
			case "Work":
				profLabel = state.T("settings.profile_work")
			}
			f.Text(fmt.Sprintf(state.T("popout.profile_fmt"), profLabel)).FontSize(11).Muted()
		})
	})
}
