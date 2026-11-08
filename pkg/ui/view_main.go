package ui

import (
	"fmt"
	"time"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
	"worldclock/pkg/clock"
	"worldclock/pkg/iploc"
	"worldclock/pkg/storage"
	"worldclock/pkg/weather"
)

// DrawMainClock 渲染主时钟屏幕（自动识别并展示当前电脑所在国家/地区与系统本地时间，支持阳历与农历双历展示、超大字体数字时钟、矢量模拟表盘、实时天气快照以及底部快捷控制栏）。
// DrawMainClock renders the primary clock screen (digital or analog) with dual solar & lunar calendar.
func DrawMainClock(f *flex.Frame, state *AppState, theme ui.Theme) {
	cfg := state.Store.GetPrefs()
	now := state.Now.Local()
	langStr := string(state.Lang())
	meta := state.DailyClockMeta(now, langStr)
	sysLoc := meta.SysLoc
	locationBadge := meta.LocationText

	timeStr, period := clock.FormatTime(now, cfg.TimeFormat24H, cfg.ShowSeconds)
	// 从自然日缓存中获取公历（阳历）与农历（阴历）本地化双历文本（避免每秒重复推算）
	solarStr, lunarStr := meta.SolarText, meta.LunarText

	// 获取今日法定节日、传统农历节日、二十四节气与网络特色节日
	var holidayBadge string
	if cfg.ShowHolidays {
		holidayBadge = meta.HolidayBadge
	}

	// 异步拉取公网 IP 所在国家、省份、城市与经纬度（严格每 6 小时自动更新一次）
	var ipBadge, ipLocStr string
	var ipInfo iploc.Info
	if state.IPLoc != nil {
		ipInfo = state.IPLoc.EnsureFresh(langStr, state.Invalidate)
		ipBadge = ipInfo.FormatBadge(langStr)
		ipLocStr = ipInfo.FormatLocation()
	}

	primaryCity := sysLoc.WeatherID
	if primaryCity == "" {
		primaryCity = "beijing"
	}

	// 实时天气获取：若已成功解析出公网 IP 的经纬度，优先获取 IP 所在位置的真实天气（每 6 小时自动更新）；离线时回退为系统时区代表城市
	var wSnap weather.Snapshot
	if ipInfo.IsSuccess && (ipInfo.Latitude != 0 || ipInfo.Longitude != 0) {
		wSnap = state.Weather.EnsureFreshCoords("iploc-current", ipInfo.Latitude, ipInfo.Longitude, state.Invalidate)
	} else {
		wSnap = state.Weather.EnsureFresh(primaryCity, state.Invalidate)
	}
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

			// 1. 表盘下方的数字时钟读数（居中大字号）
			f.Row().Items(flex.Baseline).Gap(10).Children(func() {
				state.ApplyDigitsFont(f.Text(timeStr).FontSize(52).Bold()).Color(state.ClockColor(theme))
				if period != "" {
					f.Text(period).FontSize(24).Bold().Color(theme.Primary)
				}
			})

			// 2. 紧随时间下方：IP 所在国家 与 所在国家/地区 徽章
			f.Row().PadX(16).PadY(5).Background(theme.Muted).Radius(12).Items(flex.Center).Gap(8).Children(func() {
				if ipBadge != "" {
					f.Text(ipBadge).FontSize(15).Bold().Color(theme.Accent)
					f.Text("•").FontSize(13).Muted()
				}
				f.Text(locationBadge).FontSize(15).Bold().Color(theme.Primary)
			})

			// 3. 阳历、农历与节假日信息（层级优雅、间距协调）
			f.Column().Items(flex.Center).Gap(6).Children(func() {
				f.Text(solarStr).FontSize(22).Bold().Color(theme.Foreground)
				f.Text(lunarStr).FontSize(16).Color(theme.MutedForeground)
				if holidayBadge != "" {
					f.Row().PadX(12).PadY(3).Background(theme.Muted).Radius(8).Items(flex.Center).Children(func() {
						f.Text(holidayBadge).FontSize(14).Bold().Color(theme.Primary)
					})
				}
			})

			// 4. 底部实时天气快照（天气温度前面附带 IP 所在国家省份和城市）
			f.Row().Items(flex.Center).PadX(18).PadY(6).Background(theme.Muted).Radius(14).Gap(10).Children(func() {
				if ipLocStr != "" {
					f.Text(fmt.Sprintf("📍 %s", ipLocStr)).FontSize(16).Bold().Color(theme.Foreground)
					f.Text("•").FontSize(14).Muted()
				}
				f.Text(fmt.Sprintf("%s %s (%s)", wSnap.Icon, tempStr, condStr)).FontSize(17).Bold().Color(theme.Primary)
			})
		} else {
			// Digital Clock Display with Frosted Glass Card
			// 主时钟视觉层级优化：超大时间居最顶端，归属地紧随其下，双历与天气和谐呼应
			mainTimeFontSize := state.EffectiveClockFontSize() * 1.8
			f.Column().Pad(32).Background(theme.Card).Radius(24).Border(1, theme.Border).Items(flex.Center).Gap(18).Children(func() {
				// 1. 超大数字时钟主体（居于最上方核心视觉焦点）
				f.Row().Items(flex.Baseline).Gap(12).Children(func() {
					state.ApplyDigitsFont(f.Text(timeStr).FontSize(mainTimeFontSize).Bold()).Color(state.ClockColor(theme))
					if period != "" {
						f.Text(period).FontSize(mainTimeFontSize * 0.36).Bold().Color(theme.Primary)
					}
				})

				// 2. 紧随时间下方：IP 所在国家 与 所在国家/地区 徽章
				f.Row().PadX(16).PadY(5).Background(theme.Muted).Radius(12).Items(flex.Center).Gap(8).Children(func() {
					if ipBadge != "" {
						f.Text(ipBadge).FontSize(15).Bold().Color(theme.Accent)
						f.Text("•").FontSize(13).Muted()
					}
					f.Text(locationBadge).FontSize(15).Bold().Color(theme.Primary)
				})

				// 3. 阳历、农历与节假日信息（排版舒展、字体比例协调）
				f.Column().Items(flex.Center).Gap(6).Children(func() {
					f.Text(solarStr).FontSize(22).Bold().Color(theme.Foreground)
					f.Text(lunarStr).FontSize(16).Color(theme.MutedForeground)
					if holidayBadge != "" {
						f.Row().PadX(12).PadY(3).Background(theme.Muted).Radius(8).Items(flex.Center).Children(func() {
							f.Text(holidayBadge).FontSize(14).Bold().Color(theme.Primary)
						})
					}
				})

				// 4. 底部实时天气快照（天气温度前面加上 IP 所在国家省份和城市！）
				f.Row().Items(flex.Center).PadX(18).PadY(6).Background(theme.Muted).Radius(14).Gap(10).Children(func() {
					if ipLocStr != "" {
						f.Text(fmt.Sprintf("📍 %s", ipLocStr)).FontSize(16).Bold().Color(theme.Foreground)
						f.Text("•").FontSize(14).Muted()
					}
					f.Text(fmt.Sprintf("%s %s (%s)", wSnap.Icon, tempStr, condStr)).FontSize(17).Bold().Color(theme.Primary)
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

			if cfg.EdgeDockEnabled {
				if f.Button(IDBtnQuickEdgeBar, state.T("clock.edge_bar_mode")).Outline().Clicked() {
					state.EnterEdgeBar("top")
				}
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
