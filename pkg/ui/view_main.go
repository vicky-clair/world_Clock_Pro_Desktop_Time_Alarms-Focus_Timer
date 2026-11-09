package ui

import (
	"fmt"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
	"worldclock/pkg/clock"
	"worldclock/pkg/iploc"
	"worldclock/pkg/storage"
	"worldclock/pkg/weather"
)

// DrawMainClock 渲染主时钟屏幕（自动识别并展示当前电脑所在国家/地区与系统本地时间，支持阳历与农历双历展示、超大字体数字时钟、实时天气快照、公网 IP 定位徽章、沉浸状态栏模式以及底部快捷控制栏）。
// DrawMainClock renders the primary clock screen with digital typography, dual solar & lunar calendar, and IP location weather.
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
		// 当顶部状态栏隐藏时，在主时钟右上角仅保留一个醒目纯净的“⚙”极简图标，避免文字破坏纯粹沉浸感
		if cfg.HideHeader {
			f.Row().FillWidth().Items(flex.Center).Children(func() {
				f.Spacer()
				if f.Button(IDBtnClockMiniSettings, "⚙").Small().Ghost().Clicked() {
					state.ActiveTab = TabSettings
					state.ResetScroll()
				}
			})
		}

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

		// Floating Modern Controls Dock（当开启底部状态栏隐藏模式时，不予渲染，尽享纯净大字时钟）
		if !cfg.HideFooter {
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
		}
	})
}
