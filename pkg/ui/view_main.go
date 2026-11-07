package ui

import (
	"fmt"
	"time"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
	"worldclock/pkg/clock"
	"worldclock/pkg/storage"
	"worldclock/pkg/weather"
)

// DrawMainClock 渲染主时钟屏幕（自动识别并展示当前电脑所在国家/地区与系统本地时间，支持阳历与农历双历展示、超大字体数字时钟、矢量模拟表盘、实时天气快照以及底部快捷控制栏）。
// DrawMainClock renders the primary clock screen (digital or analog) with dual solar & lunar calendar.
func DrawMainClock(f *flex.Frame, state *AppState, theme ui.Theme) {
	cfg := state.Store.Get()
	now := state.Now.Local()
	sysLoc := clock.DetectSystemLocation(now)
	langStr := string(state.Lang())
	locationBadge := sysLoc.FormatBadge(langStr)

	timeStr, period := clock.FormatTime(now, cfg.TimeFormat24H, cfg.ShowSeconds)
	// 获取公历（阳历）与农历（阴历）本地化双历文本
	solarStr, lunarStr := clock.FormatDualCalendar(now, langStr)

	// 获取今日法定节日、传统农历节日、二十四节气与网络特色节日（如1024程序员节、黑五打折、双11等）
	var holidayBadge string
	if cfg.ShowHolidays {
		holidays := clock.GetHolidays(now, sysLoc.CountryCode)
		holidayBadge = clock.FormatHolidayBadge(holidays, langStr)
	}

	primaryCity := sysLoc.WeatherID
	if primaryCity == "" {
		primaryCity = "beijing"
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

	f.Column().Fill().Pad(24).Gap(18).Items(flex.Center).Children(func() {
		// Clock Display Area
		if cfg.ClockStyle == string(clock.StyleAnalog) {
			drawAnalogFace(f, now, theme, 280)
			// Text readout & country/region below analog face (放大至近两倍)
			f.Row().Items(flex.Baseline).Gap(10).Children(func() {
				f.Text(timeStr).FontSize(56).Bold().Mono()
				if period != "" {
					f.Text(period).FontSize(32).Bold().Color(theme.Primary)
				}
			})
			f.Text(solarStr).FontSize(26).Bold().Color(theme.Foreground)
			f.Text(lunarStr).FontSize(22).Bold().Color(theme.MutedForeground)
			if holidayBadge != "" {
				f.Row().PadX(14).PadY(4).Background(theme.Muted).Radius(10).Items(flex.Center).Children(func() {
					f.Text(holidayBadge).FontSize(18).Bold().Color(theme.Primary)
				})
			}
			f.Row().Items(flex.Center).Gap(10).Children(func() {
				f.Text(locationBadge).FontSize(22).Bold().Color(theme.Foreground)
				f.Text("•").FontSize(20).Muted()
				f.Text(fmt.Sprintf("%s %s (%s)", wSnap.Icon, tempStr, condStr)).FontSize(22).Bold().Color(theme.Primary)
			})
		} else {
			// Digital Clock Display with Frosted Glass Card
			// 主时钟字号整体放大近两倍，文字清晰醒目无重复内容
			mainTimeFontSize := state.EffectiveClockFontSize() * 1.8
			f.Column().Pad(36).Background(theme.Card).Radius(20).Border(1, theme.Border).Items(flex.Center).Gap(14).Children(func() {
				// 顶部显示当前电脑所在国家/地区与本地时区徽章（唯一呈现位置，杜绝重复）
				f.Row().PadX(18).PadY(8).Background(theme.Muted).Radius(14).Items(flex.Center).Gap(8).Children(func() {
					f.Text(locationBadge).FontSize(22).Bold().Color(theme.Primary)
				})

				// 超大数字时钟主体
				f.Row().Items(flex.Center).Gap(18).Children(func() {
					f.Text(timeStr).FontSize(mainTimeFontSize).Bold().Mono().Color(state.ClockColor(theme))
					if period != "" {
						f.Row().Items(flex.Center).PadX(14).PadY(6).Background(theme.Muted).Radius(10).Children(func() {
							f.Text(period).FontSize(32).Bold().Color(theme.Primary)
						})
					}
				})

				// 阳历与农历双历展示（清晰放大）
				f.Text(solarStr).FontSize(28).Bold().Color(theme.Foreground)
				f.Text(lunarStr).FontSize(24).Bold().Color(theme.MutedForeground)

				// 当日传统节日、二十四节气或网络特色节日（如1024程序员节、黑五打折、双11等）
				if holidayBadge != "" {
					f.Row().PadX(16).PadY(6).Background(theme.Muted).Radius(12).Items(flex.Center).Children(func() {
						f.Text(holidayBadge).FontSize(20).Bold().Color(theme.Primary)
					})
				}

				// 底部实时天气快照（独立呈现，无国家/城市重复文字）
				f.Row().Items(flex.Center).PadX(16).PadY(6).Background(theme.Muted).Radius(12).Gap(8).Children(func() {
					f.Text(fmt.Sprintf("%s %s (%s)", wSnap.Icon, tempStr, condStr)).FontSize(22).Bold().Color(theme.Primary)
				})
			})
		}

		// Floating Modern Controls Dock
		f.Row().PadX(20).PadY(12).Background(theme.Card).Radius(24).Border(1, theme.Border).Items(flex.Center).Gap(12).Children(func() {
			formatLabel := state.T("clock.12h")
			if cfg.TimeFormat24H {
				formatLabel = state.T("clock.24h")
			}
			if f.Button(IDBtnToggle24H, formatLabel).Outline().Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.TimeFormat24H = !s.TimeFormat24H
				})
			}

			secLabel := state.T("clock.seconds_off")
			if cfg.ShowSeconds {
				secLabel = state.T("clock.seconds_on")
			}
			if f.Button(IDBtnToggleSeconds, secLabel).Outline().Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
					s.ShowSeconds = !s.ShowSeconds
				})
			}

			styleLabel := state.T("clock.style_digital")
			if cfg.ClockStyle == string(clock.StyleAnalog) {
				styleLabel = state.T("clock.style_analog")
			}
			if f.Button(IDBtnToggleStyle, styleLabel).Outline().Clicked() {
				state.UpdateSettings(func(s *storage.Settings) {
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

// drawAnalogFace 使用 Viva 原生 Canvas Path 绘制高精度矢量模拟时钟表盘（包含外圈圆盘、12 小时刻度、平滑时针、分针、秒针与中心圆盖）。
func drawAnalogFace(f *flex.Frame, now time.Time, theme ui.Theme, faceSize float32) {
	if faceSize <= 0 {
		faceSize = 280
	}
	slot := f.Slot().Size(faceSize, faceSize)
	rect := slot.Rect()
	canvas := f.Canvas()

	cx := rect.X + rect.Width/2
	cy := rect.Y + rect.Height/2
	radius := rect.Width / 2 * 0.92
	if radius <= 0 {
		radius = faceSize / 2 * 0.92
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
			StrokeWidth: 2.5,
		})
	}

	// Hour hand
	pbHr := canvas.PathBuilder()
	pbHr.Reset()
	pbHr.MoveTo(cx, cy)
	pbHr.LineTo(geom.HourHand.X, geom.HourHand.Y)
	canvas.DrawPath(pbHr.Path(rect.Width, rect.Height), rect, ui.PathStyle{
		Stroke:      theme.Foreground,
		StrokeWidth: 6,
	})

	// Minute hand
	pbMin := canvas.PathBuilder()
	pbMin.Reset()
	pbMin.MoveTo(cx, cy)
	pbMin.LineTo(geom.MinuteHand.X, geom.MinuteHand.Y)
	canvas.DrawPath(pbMin.Path(rect.Width, rect.Height), rect, ui.PathStyle{
		Stroke:      theme.Primary,
		StrokeWidth: 4,
	})

	// Second hand
	pbSec := canvas.PathBuilder()
	pbSec.Reset()
	pbSec.MoveTo(cx, cy)
	pbSec.LineTo(geom.SecondHand.X, geom.SecondHand.Y)
	canvas.DrawPath(pbSec.Path(rect.Width, rect.Height), rect, ui.PathStyle{
		Stroke:      theme.Danger,
		StrokeWidth: 2,
	})

	// Center cap
	canvas.Rounded(ui.Rect{
		X:      cx - 6,
		Y:      cy - 6,
		Width:  12,
		Height: 12,
	}, 6, theme.Danger)
}
