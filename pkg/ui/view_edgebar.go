package ui

import (
	"worldclock/pkg/clock"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawEdgeBar 渲染软件靠近屏幕边缘后自动收缩成的精简状态条视图（仅简洁展示当前时间、公历与农历日期，并提供按钮一键恢复正常窗口）。
// DrawEdgeBar renders the compact edge-docked status strip showing time, date, and a restore button.
func DrawEdgeBar(f *flex.Frame, state *AppState, theme ui.Theme) {
	cfg := state.Store.GetPrefs()
	now := state.Now.Local()
	langStr := string(state.Lang())
	meta := state.DailyClockMeta(now, langStr)

	timeStr, period := clock.FormatTime(now, cfg.TimeFormat24H, cfg.ShowSeconds)
	solarStr, lunarStr := meta.SolarText, meta.LunarText

	f.Row().Fill().PadX(16).PadY(8).Background(theme.Card).Items(flex.Center).Gap(14).Children(func() {
		// 左侧：边缘状态条标识与当前时间
		f.Row().Items(flex.Baseline).Gap(6).Children(func() {
			f.Text(timeStr).FontSize(30).Bold().Mono().Color(state.ClockColor(theme))
			if period != "" {
				f.Text(period).FontSize(14).Bold().Color(theme.Primary)
			}
		})

		// 中间：简洁显示阳历与农历日期
		f.Column().Gap(2).Children(func() {
			f.Text(solarStr).FontSize(13).Bold().Color(theme.Foreground)
			f.Text(lunarStr).FontSize(12).Color(theme.MutedForeground)
		})

		f.Spacer()

		// 右侧：一键恢复正常窗口按钮
		if f.Button(IDBtnRestoreEdgeBar, state.T("edgebar.restore")).Small().Clicked() {
			state.ExitEdgeBar()
		}
	})
}
