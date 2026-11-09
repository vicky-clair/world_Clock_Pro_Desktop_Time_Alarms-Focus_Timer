package ui

import (
	"fmt"

	"worldclock/pkg/clock"
	"worldclock/pkg/iploc"
	"worldclock/pkg/storage"
	"worldclock/pkg/weather"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawPopoutClock 渲染紧凑型桌面悬浮小窗时钟视图（显示电脑所在国家/地区与本地时间，支持阳历与农历双历、顶/中/底三段垂直位置自由切换与一键还原）。
// DrawPopoutClock renders the optimized popout clock widget with enlarged typography, dual calendar, and top/center/bottom positioning.
func DrawPopoutClock(f *flex.Frame, state *AppState, theme ui.Theme) {
	cfg := state.Store.GetPrefs()
	now := state.Now.Local()
	langStr := string(state.Lang())
	meta := state.DailyClockMeta(now, langStr)
	sysLoc := meta.SysLoc
	locationBadge := meta.LocationText
	timeStr, period := clock.FormatTime(now, cfg.TimeFormat24H, cfg.ShowSeconds)
	solarStr, lunarStr := meta.SolarText, meta.LunarText

	var holidayBadge string
	if cfg.ShowHolidays {
		holidayBadge = meta.HolidayBadge
	}

	pos := cfg.PopoutPosition
	if pos == "" {
		pos = storage.PopoutPosCenter
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
	weatherText := fmt.Sprintf("%s %s (%s)", wSnap.Icon, tempStr, condStr)

	f.Column().Fill().Pad(20).Background(theme.Background).Items(flex.Center).Gap(16).Children(func() {
		// Header controls: 置顶锁定、显示位置控制（顶、中间、底部）与一键还原（开启顶部状态栏隐藏时自动隐藏）
		if !cfg.HideHeader {
			f.Row().FillWidth().Items(flex.Center).Children(func() {
				pinLabel := state.T("popout.pin_off")
				if cfg.Topmost {
					pinLabel = state.T("popout.pin_on")
				}
				if f.Button(IDBtnPinPopout, pinLabel).Outline().Clicked() {
					state.SetTopmost(!cfg.Topmost)
				}

				f.Spacer()

				// 位置切换按钮组：支持在顶、中间、底部显示
				f.Row().Items(flex.Center).Gap(6).Children(func() {
					f.Text(state.T("popout.pos_label") + ":").FontSize(14).Muted()

					btnTop := f.Button(IDBtnPopoutPosTop, state.T("popout.pos_top")).Small()
					if pos != storage.PopoutPosTop {
						btnTop.Outline()
					}
					if btnTop.Clicked() {
						state.UpdateSettings(func(s *storage.Settings) {
							s.PopoutPosition = storage.PopoutPosTop
						})
					}

					btnCenter := f.Button(IDBtnPopoutPosCenter, state.T("popout.pos_center")).Small()
					if pos != storage.PopoutPosCenter {
						btnCenter.Outline()
					}
					if btnCenter.Clicked() {
						state.UpdateSettings(func(s *storage.Settings) {
							s.PopoutPosition = storage.PopoutPosCenter
						})
					}

					btnBottom := f.Button(IDBtnPopoutPosBottom, state.T("popout.pos_bottom")).Small()
					if pos != storage.PopoutPosBottom {
						btnBottom.Outline()
					}
					if btnBottom.Clicked() {
						state.UpdateSettings(func(s *storage.Settings) {
							s.PopoutPosition = storage.PopoutPosBottom
						})
					}
				})

				f.Spacer()

				if f.Button(IDBtnExitPopout, state.T("nav.restore")).Clicked() {
					state.IsPopout = false
					state.ActiveTab = TabClock
					state.Persist()
				}
			})
		} else {
			// 顶部状态栏隐藏时，在悬浮时钟右上角提供轻巧微型的“↙ 还原”快捷按钮，保持纯净并防止无法退出悬浮窗
			f.Row().FillWidth().Items(flex.Center).Children(func() {
				f.Spacer()
				if f.Button(IDBtnExitPopout, "↙ "+state.T("nav.restore")).Small().Ghost().Clicked() {
					state.IsPopout = false
					state.ActiveTab = TabClock
					state.Persist()
				}
			})
		}

		// 垂直位置调度：居中或靠底时在上部插入弹性拉伸 Spacer
		if pos == storage.PopoutPosCenter || pos == storage.PopoutPosBottom {
			f.Spacer()
		}

		// Clock representation: 超大字号数字时钟卡片
		f.Column().Pad(28).Background(theme.Card).Radius(20).Border(1, theme.Border).Items(flex.Center).Gap(14).Children(func() {
			// 1. 超大数字时钟主体（居于最上方核心视觉焦点）
			f.Row().Items(flex.Baseline).Gap(10).Children(func() {
				state.ApplyDigitsFont(f.Text(timeStr).FontSize(76).Bold()).Color(state.ClockColor(theme))
				if period != "" {
					f.Text(period).FontSize(24).Bold().Color(theme.Primary)
				}
			})

			// 2. 紧随时间下方：IP 所在国家 与 所在国家/地区 徽章
			f.Row().PadX(14).PadY(4).Background(theme.Muted).Radius(10).Items(flex.Center).Gap(6).Children(func() {
				if ipBadge != "" {
					f.Text(ipBadge).FontSize(13).Bold().Color(theme.Accent)
					f.Text("•").FontSize(11).Muted()
				}
				f.Text(locationBadge).FontSize(13).Bold().Color(theme.Primary)
			})

			// 3. 阳历、农历与节假日
			f.Column().Items(flex.Center).Gap(4).Children(func() {
				f.Text(solarStr).FontSize(18).Bold().Color(theme.Foreground)
				f.Text(lunarStr).FontSize(14).Color(theme.MutedForeground)
				if holidayBadge != "" {
					f.Row().PadX(10).PadY(2).Background(theme.Muted).Radius(6).Children(func() {
						f.Text(holidayBadge).FontSize(12).Bold().Color(theme.Primary)
					})
				}
			})

			// 4. 底部实时天气快照
			f.Row().PadX(14).PadY(4).Background(theme.Muted).Radius(10).Gap(6).Children(func() {
				if ipLocStr != "" {
					f.Text(fmt.Sprintf("📍 %s", ipLocStr)).FontSize(13).Bold().Color(theme.Foreground)
					f.Text("•").FontSize(11).Muted()
				}
				f.Text(weatherText).FontSize(14).Bold().Color(theme.Primary)
			})
		})

		// 垂直位置调度：居中或靠顶时在下部插入弹性拉伸 Spacer
		if pos == storage.PopoutPosCenter || pos == storage.PopoutPosTop {
			f.Spacer()
		}

		// Footer controls (当开启底部状态栏隐藏模式时，不予渲染)
		if !cfg.HideFooter {
			f.Row().Items(flex.Center).Gap(10).Children(func() {
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
			})
		}
	})
}
