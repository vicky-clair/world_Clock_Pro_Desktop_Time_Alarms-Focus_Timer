package ui

import (
	"fmt"

	"worldclock/pkg/license"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawHeader 渲染顶部全局分段导航栏以及正在触发的闹钟响铃提醒横幅。
// DrawHeader renders the top navigation bar and active alarm alert if triggered.
func DrawHeader(f *flex.Frame, state *AppState, theme ui.Theme) {
	// 检查是否有正在响铃的闹钟，若有则在顶部渲染醒目的红色横幅（支持贪睡 5 分钟与解除）
	ringing := state.Scheduler.Ringing()
	if ringing != nil {
		f.Row().FillWidth().Pad(12).Background(theme.Danger).Items(flex.Center).Gap(16).Children(func() {
			f.Text(state.T("alarm.alert")).Bold().Color(theme.PrimaryForeground)
			f.Text(fmt.Sprintf("%s (%s)", localizeAlarmTitle(state, ringing.Title), ringing.TriggerTime.Format("15:04"))).Color(theme.PrimaryForeground)
			f.Spacer()
			if f.Button(IDBtnSnooze, state.T("alarm.snooze")).Secondary().Clicked() {
				state.Scheduler.Snooze(5)
			}
			if f.Button(IDBtnDismiss, state.T("alarm.dismiss")).Clicked() {
				state.Scheduler.Dismiss()
				state.Persist()
			}
		})
	}

	// 顶部全局导航栏：按功能模块提供分段按钮路由，并集成 Pro 会员中心与悬浮小窗一键切换
	f.Row().FillWidth().PadX(12).PadY(8).Background(theme.Card).Border(1, theme.Border).Items(flex.Center).Gap(4).Children(func() {
		// 1. 主时钟标签页按钮
		btnClk := f.Button(TabClockID, state.T("nav.clock")).Small()
		if state.ActiveTab == TabClock && !state.ShowProModal {
			btnClk.Outline()
		} else {
			btnClk.Ghost()
		}
		if btnClk.Clicked() {
			state.ActiveTab = TabClock
			state.ShowProModal = false
			state.ResetScroll()
		}

		btnWld := f.Button(TabWorldID, state.T("nav.world")).Small()
		if state.ActiveTab == TabWorld && !state.ShowProModal {
			btnWld.Outline()
		} else {
			btnWld.Ghost()
		}
		if btnWld.Clicked() {
			state.ActiveTab = TabWorld
			state.ShowProModal = false
			state.ResetScroll()
		}

		btnAlm := f.Button(TabAlarmsID, state.T("nav.alarms")).Small()
		if state.ActiveTab == TabAlarms && !state.ShowProModal {
			btnAlm.Outline()
		} else {
			btnAlm.Ghost()
		}
		if btnAlm.Clicked() {
			state.ActiveTab = TabAlarms
			state.ShowProModal = false
			state.ResetScroll()
		}

		btnTmr := f.Button(TabTimersID, state.T("nav.timers")).Small()
		if state.ActiveTab == TabTimers && !state.ShowProModal {
			btnTmr.Outline()
		} else {
			btnTmr.Ghost()
		}
		if btnTmr.Clicked() {
			state.ActiveTab = TabTimers
			state.ShowProModal = false
			state.ResetScroll()
		}

		btnTrk := f.Button(TabTrackerID, state.T("nav.tracker")).Small()
		if state.ActiveTab == TabTracker && !state.ShowProModal {
			btnTrk.Outline()
		} else {
			btnTrk.Ghost()
		}
		if btnTrk.Clicked() {
			state.ActiveTab = TabTracker
			state.ShowProModal = false
			state.ResetScroll()
		}

		btnSlp := f.Button(TabSleepID, state.T("nav.sleep")).Small()
		if state.ActiveTab == TabSleep && !state.ShowProModal {
			btnSlp.Outline()
		} else {
			btnSlp.Ghost()
		}
		if btnSlp.Clicked() {
			state.ActiveTab = TabSleep
			state.ShowProModal = false
			state.ResetScroll()
		}

		btnNws := f.Button(TabNewsID, state.T("nav.news")).Small()
		if state.ActiveTab == TabNews && !state.ShowProModal {
			btnNws.Outline()
		} else {
			btnNws.Ghost()
		}
		if btnNws.Clicked() {
			state.ActiveTab = TabNews
			state.ShowProModal = false
			state.ResetScroll()
		}

		btnSet := f.Button(TabSettingsID, state.T("nav.settings")).Small()
		if state.ActiveTab == TabSettings && !state.ShowProModal {
			btnSet.Outline()
		} else {
			btnSet.Ghost()
		}
		if btnSet.Clicked() {
			state.ActiveTab = TabSettings
			state.ShowProModal = false
			state.ResetScroll()
		}

		// Pro 会员订阅与 7 天免费试用醒目入口按钮（符合微软应用商店显著位置展示规范）
		proLabel := state.T("nav.pro_active")
		if state.License.Plan() == license.PlanTrial {
			proLabel = fmt.Sprintf(state.T("nav.pro_trial"), state.License.TrialDaysRemaining(state.Now))
		}
		btnPro := f.Button(TabProID, proLabel).Small()
		if state.ShowProModal {
			btnPro.Outline()
		} else {
			btnPro.Secondary()
		}
		if btnPro.Clicked() {
			state.ShowProModal = !state.ShowProModal
			state.License.ClearNotice()
			state.ResetScroll()
		}

		// Popout button
		if f.Button(TabPopoutID, state.T("nav.popout")).Outline().Small().Clicked() {
			state.ShowProModal = false
			state.IsPopout = true
			state.ActiveTab = TabPopout
			state.Persist()
		}
	})
}
