package ui

import (
	"fmt"
	"time"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
	"worldclock/pkg/clock"
	"worldclock/pkg/i18n"
	"worldclock/pkg/storage"
	"worldclock/pkg/weather"
)

// DrawMainClock renders the primary clock screen (digital or analog).
func DrawMainClock(f *flex.Frame, state *AppState, theme ui.Theme) {
	cfg := state.Store.Get()
	now := state.Now
	timeStr, period := clock.FormatTime(now, cfg.TimeFormat24H, cfg.ShowSeconds)
	dateStr := i18n.FormatDateLocalized(state.Lang(), int(now.Month()), now.Day(), now.Year(), int(now.Weekday()), cfg.ShowFullDate)

	f.Column().Fill().Pad(24).Gap(20).Items(flex.Center).Children(func() {
		// Clock Display Area
		if cfg.ClockStyle == string(clock.StyleAnalog) {
			drawAnalogFace(f, now, theme)
			// Text readout below analog face
			f.Row().Items(flex.Baseline).Gap(8).Children(func() {
				f.Text(timeStr).FontSize(32).Bold().Mono()
				if period != "" {
					f.Text(period).FontSize(18).Bold().Color(theme.Primary)
				}
			})
		} else {
			// Digital Clock Display with Frosted Glass Card
			f.Column().Pad(36).Background(theme.Card).Radius(16).Border(1, theme.Border).Items(flex.Center).Gap(14).Children(func() {
				f.Row().Items(flex.Center).Gap(16).Children(func() {
					f.Text(timeStr).FontSize(state.EffectiveClockFontSize()).Bold().Mono().Color(state.ClockColor(theme))
					if period != "" {
						f.Row().Items(flex.Center).PadX(10).PadY(4).Background(theme.Muted).Radius(8).Children(func() {
							f.Text(period).FontSize(18).Bold().Color(theme.Primary)
						})
					}
				})
				f.Text(dateStr).FontSize(18).Color(theme.MutedForeground)

				_, offsetSec := now.Zone()
				offsetHours := offsetSec / 3600
				primaryCity := "beijing"
				if cities := state.World.Cities(); len(cities) > 0 && cities[0].CityID != "" {
					primaryCity = cities[0].CityID
				}
				wSnap := state.Weather.GetWeather(primaryCity)
				tempStr := weather.FormatTemperature(wSnap.TempCelsius, cfg.TempFahrenheit)
				condStr := string(wSnap.Condition)
				switch wSnap.Condition {
				case weather.ConditionSunny:
					condStr = state.T("weather.sunny")
				case weather.ConditionPartlyCloudy:
					condStr = state.T("weather.partly_cloudy")
				case weather.ConditionCloudy:
					condStr = state.T("weather.cloudy")
				case weather.ConditionRainy:
					condStr = state.T("weather.rainy")
				case weather.ConditionSnowy:
					condStr = state.T("weather.snowy")
				case weather.ConditionThunder:
					condStr = state.T("weather.thunder")
				}
				f.Row().Items(flex.Center).Gap(8).Children(func() {
					f.Text(fmt.Sprintf("%s (UTC%+d)", state.T("clock.local_timezone"), offsetHours)).FontSize(13).Muted()
					f.Text("•").FontSize(12).Muted()
					f.Text(fmt.Sprintf("%s %s (%s)", wSnap.Icon, tempStr, condStr)).FontSize(13).Bold().Color(theme.Primary)
				})
			})
		}

		// Floating Modern Controls Dock
		f.Row().PadX(18).PadY(10).Background(theme.Card).Radius(24).Border(1, theme.Border).Items(flex.Center).Gap(10).Children(func() {
			formatLabel := state.T("clock.12h")
			if cfg.TimeFormat24H {
				formatLabel = state.T("clock.24h")
			}
			if f.Button(IDBtnToggle24H, formatLabel).Outline().Clicked() {
				_ = state.Store.Update(func(s *storage.Settings) {
					s.TimeFormat24H = !s.TimeFormat24H
				})
			}

			secLabel := state.T("clock.seconds_off")
			if cfg.ShowSeconds {
				secLabel = state.T("clock.seconds_on")
			}
			if f.Button(IDBtnToggleSeconds, secLabel).Outline().Clicked() {
				_ = state.Store.Update(func(s *storage.Settings) {
					s.ShowSeconds = !s.ShowSeconds
				})
			}

			styleLabel := state.T("clock.style_digital")
			if cfg.ClockStyle == string(clock.StyleAnalog) {
				styleLabel = state.T("clock.style_analog")
			}
			if f.Button(IDBtnToggleStyle, styleLabel).Outline().Clicked() {
				_ = state.Store.Update(func(s *storage.Settings) {
					if s.ClockStyle == string(clock.StyleAnalog) {
						s.ClockStyle = string(clock.StyleDigital)
					} else {
						s.ClockStyle = string(clock.StyleAnalog)
					}
				})
			}

			if f.Button(IDBtnQuickPopout, state.T("clock.popout_mode")).Secondary().Clicked() {
				state.IsPopout = true
				state.ActiveTab = TabPopout
				state.Persist()
			}
		})
	})
}

func drawAnalogFace(f *flex.Frame, now time.Time, theme ui.Theme) {
	slot := f.Slot().Size(240, 240)
	rect := slot.Rect()
	canvas := f.Canvas()

	cx := rect.X + rect.Width/2
	cy := rect.Y + rect.Height/2
	radius := rect.Width / 2 * 0.92
	if radius <= 0 {
		radius = 100
	}

	// Draw background circle / face
	canvas.RoundedBorder(ui.Rect{
		X:      cx - radius,
		Y:      cy - radius,
		Width:  radius * 2,
		Height: radius * 2,
	}, radius, 3, theme.Border, theme.Card)

	geom := clock.ComputeAnalogGeometry(cx, cy, radius, now)

	// Draw hour ticks
	for i := range 12 {
		pb := canvas.PathBuilder()
		pb.Reset()
		pb.MoveTo(geom.HourTicks[i].X, geom.HourTicks[i].Y)
		pb.LineTo(geom.HourTickOut[i].X, geom.HourTickOut[i].Y)
		tickPath := pb.Path(rect.Width, rect.Height)
		canvas.DrawPath(tickPath, rect, ui.PathStyle{
			Stroke:      theme.MutedForeground,
			StrokeWidth: 2,
		})
	}

	// Hour hand
	pbHr := canvas.PathBuilder()
	pbHr.Reset()
	pbHr.MoveTo(cx, cy)
	pbHr.LineTo(geom.HourHand.X, geom.HourHand.Y)
	canvas.DrawPath(pbHr.Path(rect.Width, rect.Height), rect, ui.PathStyle{
		Stroke:      theme.Foreground,
		StrokeWidth: 5,
	})

	// Minute hand
	pbMin := canvas.PathBuilder()
	pbMin.Reset()
	pbMin.MoveTo(cx, cy)
	pbMin.LineTo(geom.MinuteHand.X, geom.MinuteHand.Y)
	canvas.DrawPath(pbMin.Path(rect.Width, rect.Height), rect, ui.PathStyle{
		Stroke:      theme.Primary,
		StrokeWidth: 3.5,
	})

	// Second hand
	pbSec := canvas.PathBuilder()
	pbSec.Reset()
	pbSec.MoveTo(cx, cy)
	pbSec.LineTo(geom.SecondHand.X, geom.SecondHand.Y)
	canvas.DrawPath(pbSec.Path(rect.Width, rect.Height), rect, ui.PathStyle{
		Stroke:      theme.Danger,
		StrokeWidth: 1.5,
	})

	// Center cap
	canvas.Rounded(ui.Rect{
		X:      cx - 5,
		Y:      cy - 5,
		Width:  10,
		Height: 10,
	}, 5, theme.Danger)
}
