package ui

import (
	"fmt"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawHeader renders the top navigation bar and active alarm alert if triggered.
func DrawHeader(f *flex.Frame, state *AppState, theme ui.Theme) {
	// Check for actively ringing alarm
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

	// Navigation Bar
	f.Row().FillWidth().PadX(12).PadY(8).Background(theme.Card).Border(1, theme.Border).Items(flex.Center).Gap(4).Children(func() {
		// Modern Segmented Navigation Tabs
		btnClk := f.Button(TabClockID, state.T("nav.clock")).Small()
		if state.ActiveTab == TabClock {
			btnClk.Outline()
		} else {
			btnClk.Ghost()
		}
		if btnClk.Clicked() {
			state.ActiveTab = TabClock
			state.ResetScroll()
		}

		btnWld := f.Button(TabWorldID, state.T("nav.world")).Small()
		if state.ActiveTab == TabWorld {
			btnWld.Outline()
		} else {
			btnWld.Ghost()
		}
		if btnWld.Clicked() {
			state.ActiveTab = TabWorld
			state.ResetScroll()
		}

		btnAlm := f.Button(TabAlarmsID, state.T("nav.alarms")).Small()
		if state.ActiveTab == TabAlarms {
			btnAlm.Outline()
		} else {
			btnAlm.Ghost()
		}
		if btnAlm.Clicked() {
			state.ActiveTab = TabAlarms
			state.ResetScroll()
		}

		btnTmr := f.Button(TabTimersID, state.T("nav.timers")).Small()
		if state.ActiveTab == TabTimers {
			btnTmr.Outline()
		} else {
			btnTmr.Ghost()
		}
		if btnTmr.Clicked() {
			state.ActiveTab = TabTimers
			state.ResetScroll()
		}

		btnTrk := f.Button(TabTrackerID, state.T("nav.tracker")).Small()
		if state.ActiveTab == TabTracker {
			btnTrk.Outline()
		} else {
			btnTrk.Ghost()
		}
		if btnTrk.Clicked() {
			state.ActiveTab = TabTracker
			state.ResetScroll()
		}

		btnSlp := f.Button(TabSleepID, state.T("nav.sleep")).Small()
		if state.ActiveTab == TabSleep {
			btnSlp.Outline()
		} else {
			btnSlp.Ghost()
		}
		if btnSlp.Clicked() {
			state.ActiveTab = TabSleep
			state.ResetScroll()
		}

		btnNws := f.Button(TabNewsID, state.T("nav.news")).Small()
		if state.ActiveTab == TabNews {
			btnNws.Outline()
		} else {
			btnNws.Ghost()
		}
		if btnNws.Clicked() {
			state.ActiveTab = TabNews
			state.ResetScroll()
		}

		btnSet := f.Button(TabSettingsID, state.T("nav.settings")).Small()
		if state.ActiveTab == TabSettings {
			btnSet.Outline()
		} else {
			btnSet.Ghost()
		}
		if btnSet.Clicked() {
			state.ActiveTab = TabSettings
			state.ResetScroll()
		}

		// Popout button
		if f.Button(TabPopoutID, state.T("nav.popout")).Outline().Small().Clicked() {
			state.IsPopout = true
			state.ActiveTab = TabPopout
			state.Persist()
		}
	})
}
