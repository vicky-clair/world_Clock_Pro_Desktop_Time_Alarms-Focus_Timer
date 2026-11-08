package ui

import (
	"fmt"

	"worldclock/pkg/clock"
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

	primaryCity := sysLoc.WeatherID
	if primaryCity == "" {
		primaryCity = "beijing"
	}
	wSnap := state.Weather.EnsureFresh(primaryCity, state.Invalidate)
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
		// Header controls: 置顶锁定、显示位置控制（顶、中间、底部）与一键还原
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

		// 垂直位置调度：居中或靠底时在上部插入弹性拉伸 Spacer
		if pos == storage.PopoutPosCenter || pos == storage.PopoutPosBottom {
			f.Spacer()
		}

		// Clock representation: 整体超大字号呈现
		if cfg.ClockStyle == string(clock.StyleAnalog) {
			drawAnalogFace(f, now, theme, 260)
			f.Row().Items(flex.Baseline).Gap(10).Children(func() {
				state.ApplyDigitsFont(f.Text(timeStr).FontSize(48).Bold())
				if period != "" {
					f.Text(period).FontSize(26).Bold().Color(theme.Primary)
				}
			})
			f.Text(solarStr).FontSize(22).Bold().Color(theme.Foreground)
			f.Text(lunarStr).FontSize(18).Color(theme.MutedForeground)
			if holidayBadge != "" {
				f.Text(holidayBadge).FontSize(16).Bold().Color(theme.Primary)
			}
			f.Row().Items(flex.Center).Gap(8).Children(func() {
				f.Text(locationBadge).FontSize(18).Bold().Color(theme.Primary)
				f.Text("•").FontSize(16).Muted()
				f.Text(weatherText).FontSize(18).Bold().Color(theme.Foreground)
			})
		} else {
			// Digital Popout Clock: 超大字号数字时钟卡片 (88px 大时间 + 24px 双历 + 20px 徽章)
			f.Column().Pad(32).Background(theme.Card).Radius(20).Border(1, theme.Border).Items(flex.Center).Gap(14).Children(func() {
				// 所在国家/地区与本地时区徽章
				f.Row().PadX(16).PadY(6).Background(theme.Muted).Radius(12).Items(flex.Center).Gap(8).Children(func() {
					f.Text(locationBadge).FontSize(20).Bold().Color(theme.Primary)
				})

				// 超大数字时钟
				f.Row().Items(flex.Baseline).Gap(12).Children(func() {
					state.ApplyDigitsFont(f.Text(timeStr).FontSize(88).Bold()).Color(state.ClockColor(theme))
					if period != "" {
						f.Text(period).FontSize(28).Bold().Color(theme.Primary)
					}
				})

				// 阳历与农历双历
				f.Text(solarStr).FontSize(24).Bold().Color(theme.Foreground)
				f.Text(lunarStr).FontSize(20).Color(theme.MutedForeground)

				// 今日节日与节气
				if holidayBadge != "" {
					f.Row().PadX(14).PadY(5).Background(theme.Muted).Radius(10).Children(func() {
						f.Text(holidayBadge).FontSize(17).Bold().Color(theme.Primary)
					})
				}

				// 天气快照
				f.Row().PadX(14).PadY(5).Background(theme.Muted).Radius(10).Children(func() {
					f.Text(weatherText).FontSize(18).Bold().Color(theme.Primary)
				})
			})
		}

		// 垂直位置调度：居中或靠顶时在下部插入弹性拉伸 Spacer
		if pos == storage.PopoutPosCenter || pos == storage.PopoutPosTop {
			f.Spacer()
		}

		// Footer controls
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
		})
	})
}
