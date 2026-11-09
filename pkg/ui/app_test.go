package ui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"worldclock/pkg/alarm"
	"worldclock/pkg/clock"
	"worldclock/pkg/iploc"
	"worldclock/pkg/license"
	"worldclock/pkg/storage"

	"github.com/vivaui/viva/app"
	vui "github.com/vivaui/viva/ui"
)

func clickID(t *testing.T, driver *app.Driver, id vui.ID) bool {
	t.Helper()
	for _, hit := range driver.Snapshot().Hits {
		if hit.ID == id {
			b := hit.Bounds
			if err := driver.Click(b.X+b.Width/2, b.Y+b.Height/2); err != nil {
				t.Fatalf("click failed on %v: %v", id, err)
			}
			return true
		}
	}
	return false
}

func hasLabel(snapshot app.Snapshot, substr string) bool {
	for _, l := range snapshot.Labels {
		if strings.Contains(l.Value, substr) {
			return true
		}
	}
	return false
}

func setupTestApp(t *testing.T) (*AppState, *app.Driver) {
	t.Helper()
	tmpDir := t.TempDir()
	store := storage.NewStore(filepath.Join(tmpDir, "settings.json"))
	_ = store.Update(func(s *storage.Settings) {
		s.Language = "en"
	})
	state := NewAppState(store)

	var canvas vui.Canvas
	cfg := NewAppConfig(state)
	cfg.Width = 900
	cfg.Height = 1600
	driver, err := app.NewDriver(cfg, &canvas)
	if err != nil {
		t.Fatalf("failed to create app.Driver: %v", err)
	}
	return state, driver
}

func TestMainClockInteractions(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	if state.ActiveTab != TabClock {
		t.Fatalf("expected initial tab to be Clock, got %s", state.ActiveTab)
	}

	// 1. Toggle 24H -> 12H
	if !state.Store.Get().TimeFormat24H {
		t.Fatal("expected default to be 24H")
	}
	if !clickID(t, driver, IDBtnToggle24H) {
		t.Fatal("failed to click 24H toggle button")
	}
	if state.Store.Get().TimeFormat24H {
		t.Fatal("expected format to change to 12H")
	}

	// 2. Toggle Seconds ON -> OFF
	if !state.Store.Get().ShowSeconds {
		t.Fatal("expected default seconds ON")
	}
	if !clickID(t, driver, IDBtnToggleSeconds) {
		t.Fatal("failed to click seconds toggle button")
	}
	if state.Store.Get().ShowSeconds {
		t.Fatal("expected seconds to be OFF")
	}

	// 3. Clock Style is consistently digital
	if state.Store.Get().ClockStyle != string(clock.StyleDigital) {
		t.Fatal("expected default digital clock style")
	}
}

func TestTabNavigation(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Switch to World Clock
	if !clickID(t, driver, TabWorldID) {
		t.Fatal("failed to click World tab")
	}
	if state.ActiveTab != TabWorld {
		t.Fatalf("expected tab World, got %s", state.ActiveTab)
	}
	if !hasLabel(driver.Snapshot(), "World Clocks") {
		t.Fatal("expected 'World Clocks' title on screen")
	}

	// Switch to Alarms
	if !clickID(t, driver, TabAlarmsID) {
		t.Fatal("failed to click Alarms tab")
	}
	if state.ActiveTab != TabAlarms {
		t.Fatalf("expected tab Alarms, got %s", state.ActiveTab)
	}
	if !hasLabel(driver.Snapshot(), "Alarms") {
		t.Fatal("expected 'Alarms' title on screen")
	}

	// Switch to Timers
	if !clickID(t, driver, TabTimersID) {
		t.Fatal("failed to click Timers tab")
	}
	if state.ActiveTab != TabTimers {
		t.Fatalf("expected tab Timers, got %s", state.ActiveTab)
	}
	if !hasLabel(driver.Snapshot(), "Countdown Timers") {
		t.Fatal("expected 'Countdown Timers' on screen")
	}

	// Switch to Settings
	if !clickID(t, driver, TabSettingsID) {
		t.Fatal("failed to click Settings tab")
	}
	if state.ActiveTab != TabSettings {
		t.Fatalf("expected tab Settings, got %s", state.ActiveTab)
	}
	if !hasLabel(driver.Snapshot(), "Settings & Preferences") {
		t.Fatal("expected 'Settings & Preferences' on screen")
	}

	// Switch to Popout mode
	if !clickID(t, driver, TabPopoutID) {
		t.Fatal("failed to click Popout tab")
	}
	if !state.IsPopout || state.ActiveTab != TabPopout {
		t.Fatal("expected IsPopout to be true")
	}
	if !hasLabel(driver.Snapshot(), "Restore ↙") {
		t.Fatal("expected 'Restore ↙' button in Popout mode")
	}

	// Restore from Popout mode
	if !clickID(t, driver, IDBtnExitPopout) {
		t.Fatal("failed to click Restore button in Popout mode")
	}
	if state.IsPopout || state.ActiveTab != TabClock {
		t.Fatal("expected to exit Popout mode back to Clock")
	}
}

func TestWorldClockAddRemove(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Navigate to World Tab
	clickID(t, driver, TabWorldID)

	initialCount := len(state.World.Cities())

	// Open Add City modal
	if !clickID(t, driver, IDBtnAddCityModal) {
		t.Fatal("failed to click + Add City button")
	}
	if !state.ShowAddCityModal {
		t.Fatal("expected ShowAddCityModal to be true")
	}

	// Click Add Afghanistan (index 0 of 193 catalog)
	addAfgID := CityAddCatalogID("afghanistan")
	if !clickID(t, driver, addAfgID) {
		t.Fatal("failed to click Add Afghanistan button")
	}

	if len(state.World.Cities()) != initialCount+1 {
		t.Fatalf("expected city count %d, got %d", initialCount+1, len(state.World.Cities()))
	}
	if state.ShowAddCityModal {
		t.Fatal("modal should close after adding city")
	}

	// Remove Afghanistan
	rmAfgID := CityRemoveID("afghanistan")
	if !clickID(t, driver, rmAfgID) {
		t.Fatal("failed to click Remove Afghanistan button")
	}
	if len(state.World.Cities()) != initialCount {
		t.Fatalf("expected city count back to %d, got %d", initialCount, len(state.World.Cities()))
	}

	// Verify original default cities (beijing, tokyo, london, newyork) can be reordered, favorited, and deleted!
	if !clickID(t, driver, CityMoveDownID("beijing")) {
		t.Fatal("failed to click MoveDown on default city beijing")
	}
	if state.World.Cities()[1].CityID != "beijing" {
		t.Fatalf("expected beijing at index 1 after MoveDown, got %s", state.World.Cities()[1].CityID)
	}
	if !clickID(t, driver, CityMoveUpID("beijing")) {
		t.Fatal("failed to click MoveUp on default city beijing")
	}
	if state.World.Cities()[0].CityID != "beijing" {
		t.Fatalf("expected beijing back at index 0 after MoveUp, got %s", state.World.Cities()[0].CityID)
	}

	for _, origID := range []string{"beijing", "tokyo", "london", "newyork"} {
		if !clickID(t, driver, CityRemoveID(origID)) {
			t.Fatalf("failed to click Remove button for default city %s", origID)
		}
	}
	if len(state.World.Cities()) != 0 {
		t.Fatalf("expected 0 cities after deleting all 4 original cities, got %d", len(state.World.Cities()))
	}

	// Reload state from store and verify deleted default cities stay deleted
	reloadedState := NewAppState(state.Store)
	if len(reloadedState.World.Cities()) != 0 {
		t.Fatalf("expected reloaded state to keep 0 cities, got %d", len(reloadedState.World.Cities()))
	}
}

func TestAlarmsAddToggle(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Navigate to Alarms Tab
	clickID(t, driver, TabAlarmsID)

	initialCount := len(state.Scheduler.Alarms())

	// Open New Alarm modal
	if !clickID(t, driver, IDBtnNewAlarmModal) {
		t.Fatal("failed to click + New Alarm button")
	}
	if !state.ShowNewAlarmModal {
		t.Fatal("expected ShowNewAlarmModal to be true")
	}

	// Click Morning preset
	if !clickID(t, driver, IDBtnPresetMorning) {
		t.Fatal("failed to click Morning preset")
	}
	if state.NewAlarmHour != 7 || state.NewAlarmMinute != 30 {
		t.Fatalf("expected 07:30, got %02d:%02d", state.NewAlarmHour, state.NewAlarmMinute)
	}

	// Click Save
	if !clickID(t, driver, IDBtnSaveNewAlarm) {
		t.Fatal("failed to click Save New Alarm")
	}
	if len(state.Scheduler.Alarms()) != initialCount+1 {
		t.Fatalf("expected alarms count %d, got %d", initialCount+1, len(state.Scheduler.Alarms()))
	}

	// Toggle newly created alarm
	newAlarmID := state.Scheduler.Alarms()[len(state.Scheduler.Alarms())-1].ID
	togID := AlarmToggleID(newAlarmID)
	if !clickID(t, driver, togID) {
		t.Fatal("failed to click toggle alarm button")
	}

	// Verify it was toggled
	found := false
	for _, a := range state.Scheduler.Alarms() {
		if a.ID == newAlarmID {
			if a.IsEnabled {
				t.Fatal("expected alarm to be disabled after toggle")
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("could not find alarm after toggle")
	}
}

func TestTimersAndStopwatch(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Navigate to Timers Tab
	clickID(t, driver, TabTimersID)

	firstTimer := state.TimerMgr.Timers()[0]
	startID := TimerStartID(firstTimer.ID)
	pauseID := TimerPauseID(firstTimer.ID)
	resetID := TimerResetID(firstTimer.ID)

	// Start timer
	if !clickID(t, driver, startID) {
		t.Fatal("failed to click Start Timer")
	}
	if firstTimer.State != "running" {
		t.Fatalf("expected timer running, got %s", firstTimer.State)
	}

	// Pause timer
	if !clickID(t, driver, pauseID) {
		t.Fatal("failed to click Pause Timer")
	}
	if firstTimer.State != "paused" {
		t.Fatalf("expected timer paused, got %s", firstTimer.State)
	}

	// Reset timer
	if !clickID(t, driver, resetID) {
		t.Fatal("failed to click Reset Timer")
	}
	if firstTimer.State != "idle" {
		t.Fatalf("expected timer idle, got %s", firstTimer.State)
	}

	// Stopwatch operations
	sw := state.Stopwatch
	if !clickID(t, driver, IDBtnSwStart) {
		t.Fatal("failed to click Start Stopwatch")
	}
	if sw.State() != "running" {
		t.Fatalf("expected stopwatch running, got %s", sw.State())
	}

	// Record Lap
	if !clickID(t, driver, IDBtnSwLap) {
		t.Fatal("failed to click Lap Split")
	}
	if len(sw.Laps()) != 1 {
		t.Fatalf("expected 1 lap, got %d", len(sw.Laps()))
	}

	// Pause
	if !clickID(t, driver, IDBtnSwPause) {
		t.Fatal("failed to click Pause Stopwatch")
	}
	if sw.State() != "paused" {
		t.Fatalf("expected stopwatch paused, got %s", sw.State())
	}

	// Reset
	if !clickID(t, driver, IDBtnSwReset) {
		t.Fatal("failed to click Reset Stopwatch")
	}
	if sw.State() != "idle" || len(sw.Laps()) != 0 {
		t.Fatalf("expected stopwatch reset to idle with 0 laps")
	}
}

func TestRingingAlarmAlert(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Simulate an alarm firing
	state.Scheduler.AddAlarm(alarm.Alarm{
		ID:           "test-alert",
		Title:        "Important Meeting",
		Hour:         14,
		Minute:       0,
		ScheduleType: alarm.ScheduleOneTime,
		IsEnabled:    true,
	})

	at := time.Date(2026, 11, 5, 14, 0, 0, 0, time.UTC)
	state.Scheduler.CheckTriggers(at)

	// Trigger paint with alarm ringing
	if err := driver.Resize(900, 800); err != nil {
		t.Fatal(err)
	}
	snap := driver.Snapshot()
	if !hasLabel(snap, "ALARM ALERT") {
		t.Fatal("expected ALARM ALERT banner when alarm is ringing")
	}
	if !hasLabel(snap, "Important Meeting") {
		t.Fatal("expected alarm title in banner")
	}

	// Snooze
	if !clickID(t, driver, IDBtnSnooze) {
		t.Fatal("failed to click Snooze button")
	}

	// Dismiss
	if !clickID(t, driver, IDBtnDismiss) {
		t.Fatal("failed to click Dismiss button")
	}
	if state.Scheduler.Ringing() != nil {
		t.Fatal("expected ringing to be dismissed")
	}
}

func TestAlarmCatchUpInUI(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// 添加一条 3 分钟前设定的闹钟（模拟电脑睡眠期间错过精确分钟点）
	alarmTime := time.Now().Add(-3 * time.Minute)
	state.Scheduler.AddAlarm(alarm.Alarm{
		ID:           "test-sleep-catchup",
		Title:        "Sleep Catchup Test",
		Hour:         alarmTime.Hour(),
		Minute:       alarmTime.Minute(),
		ScheduleType: alarm.ScheduleDaily,
		IsEnabled:    true,
	})

	// 触发真实 Paint 重绘管线，验证 CheckTriggers 在渲染中自动追赶并触发补响
	if err := driver.Resize(920, 800); err != nil {
		t.Fatal(err)
	}
	snap := driver.Snapshot()
	if !hasLabel(snap, "ALARM ALERT") {
		t.Fatal("expected alarm alert banner to catch up and trigger even if checked late")
	}
	if !hasLabel(snap, "Sleep Catchup Test") {
		t.Fatal("expected alarm title in banner")
	}

	// 验证 Dismiss 解除
	if !clickID(t, driver, IDBtnDismiss) {
		t.Fatal("failed to click Dismiss button")
	}
	if state.Scheduler.Ringing() != nil {
		t.Fatal("expected alarm to be dismissed")
	}
}

func TestSettingsThemeSwitch(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Navigate to Settings -> Wallpaper/Theme subtab
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabWallpaper)
	// Switch to Dark
	if !clickID(t, driver, IDThemeDark) {
		t.Fatal("failed to click Dark theme button")
	}
	if state.Store.Get().Theme != storage.ThemeDark {
		t.Fatalf("expected theme dark, got %s", state.Store.Get().Theme)
	}

	// Switch to Nord
	if !clickID(t, driver, IDThemeNord) {
		t.Fatal("failed to click Nord theme button")
	}
	if state.Store.Get().Theme != storage.ThemeNord {
		t.Fatalf("expected theme nord, got %s", state.Store.Get().Theme)
	}

	// Switch to Olive
	if !clickID(t, driver, IDThemeOlive) {
		t.Fatal("failed to click Olive theme button")
	}
	if state.Store.Get().Theme != storage.ThemeOlive {
		t.Fatalf("expected theme olive, got %s", state.Store.Get().Theme)
	}
}

func TestTrackerViewInteractions(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Navigate to Tracker
	if !clickID(t, driver, TabTrackerID) {
		t.Fatal("failed to click Tracker tab")
	}
	if !hasLabel(driver.Snapshot(), "Productivity Time Tracker") {
		t.Fatal("expected Tracker screen title")
	}

	// DND toggle
	if !clickID(t, driver, IDBtnTrackerDND) {
		t.Fatal("failed to click DND button")
	}
	if !state.Tracker.Config().DoNotDisturb {
		t.Fatal("expected DND to be true")
	}

	// Start Work
	if !clickID(t, driver, IDBtnTrackerStart) {
		t.Fatal("failed to click Start Work Session")
	}
	if state.Tracker.Status() != "Working" {
		t.Fatalf("expected Working status, got %s", state.Tracker.Status())
	}

	// Pause
	if !clickID(t, driver, IDBtnTrackerPause) {
		t.Fatal("failed to click Pause button")
	}
	if state.Tracker.Status() != "Paused" {
		t.Fatalf("expected Paused status, got %s", state.Tracker.Status())
	}

	// Resume
	if !clickID(t, driver, IDBtnTrackerResume) {
		t.Fatal("failed to click Resume button")
	}
	if state.Tracker.Status() != "Working" {
		t.Fatalf("expected Working status after resume, got %s", state.Tracker.Status())
	}

	// Reset
	if !clickID(t, driver, IDBtnTrackerReset) {
		t.Fatal("failed to click Reset button")
	}
	if state.Tracker.Status() != "Idle" {
		t.Fatalf("expected Idle status after reset, got %s", state.Tracker.Status())
	}
}

func TestSleepViewInteractions(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Navigate to Sleep tab
	if !clickID(t, driver, TabSleepID) {
		t.Fatal("failed to click Sleep tab")
	}
	if !hasLabel(driver.Snapshot(), "SLEEP & AMBIENT MODE") {
		t.Fatal("expected Sleep screen title")
	}

	// Select 15m duration
	if !clickID(t, driver, IDBtnSleep15m) {
		t.Fatal("failed to click 15m duration")
	}
	if state.SleepDuration != 15*time.Minute {
		t.Fatalf("expected 15m duration, got %v", state.SleepDuration)
	}

	// Select Waves sound
	if !clickID(t, driver, IDBtnSleepWaves) {
		t.Fatal("failed to click Waves sound")
	}

	// Start Sleep Mode
	if !clickID(t, driver, IDBtnSleepStart) {
		t.Fatal("failed to click Start Sleep Mode")
	}
	st, snd, _ := state.Sleep.Info()
	if st != "Active" || snd != "Ocean Waves" {
		t.Fatalf("expected Active Ocean Waves, got state=%s, snd=%s", st, snd)
	}

	// Exit Sleep Mode
	if !clickID(t, driver, IDBtnSleepStop) {
		t.Fatal("failed to click Exit Sleep Mode")
	}
	stEnd, _, _ := state.Sleep.Info()
	if stEnd != "Idle" {
		t.Fatalf("expected Idle after stop, got %s", stEnd)
	}
}

func TestNewsViewInteractions(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Navigate to On This Day (History) tab
	if !clickID(t, driver, TabNewsID) {
		t.Fatal("failed to click On This Day tab")
	}
	if !hasLabel(driver.Snapshot(), "On This Day in History") {
		t.Fatal("expected On This Day in History screen title")
	}

	// Filter Tech
	if !clickID(t, driver, IDNewsCatTech) {
		t.Fatal("failed to click Tech category")
	}
	if state.NewsCategory != "Tech" {
		t.Fatalf("expected Tech category, got %s", state.NewsCategory)
	}

	// Filter Science
	if !clickID(t, driver, IDNewsCatScience) {
		t.Fatal("failed to click Science category")
	}
	if state.NewsCategory != "Science" {
		t.Fatalf("expected Science category, got %s", state.NewsCategory)
	}
}

func TestTemperatureUnitToggle(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Navigate to Settings -> Clock tab (where Temperature Unit is located)
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabClock)
	if state.Store.Get().TempFahrenheit {
		t.Fatal("expected default Celsius")
	}

	// Toggle to Fahrenheit
	if !clickID(t, driver, IDOptTempFahrenheit) {
		t.Fatal("failed to click Temperature Unit button")
	}
	if !state.Store.Get().TempFahrenheit {
		t.Fatal("expected Fahrenheit after toggle")
	}

	// Navigate back to Clock tab
	clickID(t, driver, TabClockID)
	snap := driver.Snapshot()
	if !hasLabel(snap, "°F") {
		t.Fatal("expected °F in weather readout")
	}
}

func TestLanguageSwitching(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Navigate to Settings -> Language tab
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	// 1. Switch to Chinese (ZH)
	if !clickID(t, driver, IDLangZH) {
		t.Fatal("failed to click ZH language button")
	}
	if state.Lang() != "zh" {
		t.Fatalf("expected zh language, got %s", state.Lang())
	}
	snapZH := driver.Snapshot()
	if !hasLabel(snapZH, "时钟") {
		t.Fatal("expected '时钟' label in Chinese UI")
	}
	if !hasLabel(snapZH, "世界时钟") {
		t.Fatal("expected '世界时钟' label in Chinese UI")
	}

	// 2. Switch to Japanese (JA)
	if !clickID(t, driver, IDLangJA) {
		t.Fatal("failed to click JA language button")
	}
	if state.Lang() != "ja" {
		t.Fatalf("expected ja language, got %s", state.Lang())
	}
	snapJA := driver.Snapshot()
	if !hasLabel(snapJA, "時計") {
		t.Fatal("expected '時計' label in Japanese UI")
	}
	if !hasLabel(snapJA, "世界時計") {
		t.Fatal("expected '世界時計' label in Japanese UI")
	}

	// 3. Switch back to English (EN)
	if !clickID(t, driver, IDLangEN) {
		t.Fatal("failed to click EN language button")
	}
	if state.Lang() != "en" {
		t.Fatalf("expected en language, got %s", state.Lang())
	}
	snapEN := driver.Snapshot()
	if !hasLabel(snapEN, "Clock") {
		t.Fatal("expected 'Clock' label in English UI")
	}
}

func TestWallpaperSettings(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Navigate to Settings
	// Navigate to Settings -> Wallpaper tab
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabWallpaper)
	if state.Wallpaper.Config().Enabled {
		t.Fatal("expected default wallpaper disabled")
	}

	// Toggle Wallpaper ON
	if !clickID(t, driver, IDWallpaperToggle) {
		t.Fatal("failed to click Wallpaper toggle")
	}
	if !state.Wallpaper.Config().Enabled {
		t.Fatal("expected wallpaper enabled")
	}

	// Change Fit Mode to Tile
	if !clickID(t, driver, IDWallpaperFitTile) {
		t.Fatal("failed to click Tile fit mode")
	}
	if state.Wallpaper.Config().FitMode != "tile" {
		t.Fatalf("expected tile fit mode, got %s", state.Wallpaper.Config().FitMode)
	}

	// Change Fit Mode to Stretch
	if !clickID(t, driver, IDWallpaperFitStretch) {
		t.Fatal("failed to click Stretch fit mode")
	}
	if state.Wallpaper.Config().FitMode != "stretch" {
		t.Fatalf("expected stretch fit mode, got %s", state.Wallpaper.Config().FitMode)
	}

	// Opacity down
	initialOp := state.Wallpaper.Config().Opacity
	if !clickID(t, driver, IDWallpaperOpacityDown) {
		t.Fatal("failed to click Opacity down")
	}
	if state.Wallpaper.Config().Opacity >= initialOp {
		t.Fatalf("expected opacity to decrease from %f, got %f", initialOp, state.Wallpaper.Config().Opacity)
	}

	// Opacity up
	if !clickID(t, driver, IDWallpaperOpacityUp) {
		t.Fatal("failed to click Opacity up")
	}

	// Test Presets
	if !clickID(t, driver, IDPresetGrid) {
		t.Fatal("failed to click Grid preset")
	}
	if state.Wallpaper.Config().Preset != "grid" {
		t.Fatalf("expected grid preset, got %s", state.Wallpaper.Config().Preset)
	}

	// Card glass opacity down/up
	initialGlass := state.Wallpaper.Config().CardOpacity
	if !clickID(t, driver, IDCardGlassDown) {
		t.Fatal("failed to click Card Glass down")
	}
	if state.Wallpaper.Config().CardOpacity >= initialGlass {
		t.Fatal("expected card glass to decrease")
	}
	// Test sample local image button
	if !clickID(t, driver, IDBtnSampleWallpaper) {
		t.Fatal("failed to click Sample Wallpaper button")
	}
	if state.Wallpaper.Config().ImagePath == "" {
		t.Fatal("expected custom ImagePath to be populated after sample load")
	}

	// Test mouse wheel OnScroll event on ScrollArea
	if err := driver.Scroll(vui.ScrollEvent{X: 400, Y: 300, DeltaY: -100}); err != nil {
		t.Fatalf("driver.Scroll failed: %v", err)
	}
	ss := state.View.ScrollState(IDScrollSettings)
	if ss != nil && ss.Offset < 0 {
		t.Fatal("expected non-negative scroll offset")
	}

	// Test file drop handler
	accepted, err := driver.Drop(vui.FileDropEvent{
		Phase: vui.FileDropDrop,
		Paths: []string{"C:\\fake\\custom_wallpaper.png"},
	})
	if err != nil || !accepted {
		t.Fatalf("driver.Drop failed: %v, accepted=%v", err, accepted)
	}
}

func TestCountry193AndContinents(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Navigate to World tab
	clickID(t, driver, TabWorldID)

	// Open Add City modal
	if !clickID(t, driver, IDBtnAddCityModal) {
		t.Fatal("failed to click Add City modal button")
	}
	if !state.ShowAddCityModal {
		t.Fatal("expected modal to be open")
	}

	// Filter by Asia
	if !clickID(t, driver, IDContAsia) {
		t.Fatal("failed to click Asia continent tab")
	}
	if state.SelectedContinent != "亚洲" {
		t.Fatalf("expected Asia selected, got %s", state.SelectedContinent)
	}

	// Filter by Europe
	if !clickID(t, driver, IDContEurope) {
		t.Fatal("failed to click Europe continent tab")
	}
	if state.SelectedContinent != "欧洲" {
		t.Fatalf("expected Europe selected, got %s", state.SelectedContinent)
	}
	// Add Albania from 193 catalog (first European country)
	addAlbaniaID := CityAddCatalogID("albania")
	if !clickID(t, driver, addAlbaniaID) {
		t.Fatal("failed to click Add Albania")
	}

	// Verify Albania is in user cities
	found := false
	for _, c := range state.World.Cities() {
		if c.CityID == "albania" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected Albania in user cities")
	}
}

func TestFontSizeAndColorCustomization(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Navigate to Settings -> Language sub-tab (merged Language & Font settings)
	clickID(t, driver, TabSettingsID)
	if !clickID(t, driver, IDSettingsTabLanguage) {
		t.Fatal("failed to click Language sub-tab")
	}

	// 1. Change Font Size to 76px
	if !clickID(t, driver, IDFontSize76) {
		t.Fatal("failed to click 76px font size")
	}
	if state.Store.Get().ClockFontSize != 76 {
		t.Fatalf("expected 76px font size, got %d", state.Store.Get().ClockFontSize)
	}
	if state.EffectiveClockFontSize() != 76 {
		t.Fatalf("expected 76 EffectiveClockFontSize, got %f", state.EffectiveClockFontSize())
	}

	// 2. Change Font Color to Gold
	if !clickID(t, driver, IDFontColorGold) {
		t.Fatal("failed to click Gold font color")
	}
	if state.Store.Get().ClockFontColor != "gold" {
		t.Fatalf("expected gold font color, got %s", state.Store.Get().ClockFontColor)
	}
	theme := BuildTheme(storage.ThemeDark, false, 0.5, "default")
	c := state.ClockColor(theme)
	if c.Red != 1.0 || c.Alpha != 1.0 { // Gold is #FFBE0B -> Red is 255/255 = 1.0
		t.Fatalf("unexpected clock color: %+v", c)
	}

	// 3. Change UI Font Color to Black (for light wallpapers)
	if !clickID(t, driver, IDUIColorBlack) {
		t.Fatal("failed to click Black UI color")
	}
	if state.Store.Get().UIFontColor != "black" {
		t.Fatalf("expected black UI font color, got %s", state.Store.Get().UIFontColor)
	}
	themeBlack := BuildTheme(storage.ThemeDark, true, 0.5, "black")
	if themeBlack.Foreground.Red > 0.20 { // Black foreground should be dark (0x111315 < 0.20)
		t.Fatalf("expected dark foreground for black text mode, got %+v", themeBlack.Foreground)
	}

	// 4. Change UI Font Color to White (for dark wallpapers)
	if !clickID(t, driver, IDUIColorWhite) {
		t.Fatal("failed to click White UI color")
	}
	if state.Store.Get().UIFontColor != "white" {
		t.Fatalf("expected white UI font color, got %s", state.Store.Get().UIFontColor)
	}

	// 5. Change Font Scale to 110%
	if !clickID(t, driver, IDFontScale110) {
		t.Fatal("failed to click 110% font scale")
	}
	if state.Store.Get().FontScale != 1.10 {
		t.Fatalf("expected 1.10 font scale, got %f", state.Store.Get().FontScale)
	}
	// Navigate back to Clock tab
	clickID(t, driver, TabClockID)
	snap := driver.Snapshot()
	if len(snap.Labels) == 0 {
		t.Fatal("expected labels in main clock")
	}
}

func TestWorldClockLanguageSync(t *testing.T) {
	_, driver := setupTestApp(t)
	defer driver.Close()

	// 1. Switch to Chinese
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangZH)

	// Check World Clock in Chinese
	clickID(t, driver, TabWorldID)
	snapZH := driver.Snapshot()
	if !hasLabel(snapZH, "北京") {
		t.Fatal("expected '北京' in Chinese World Clock")
	}
	if !hasLabel(snapZH, "中国") {
		t.Fatal("expected '中国' in Chinese World Clock")
	}
	if !hasLabel(snapZH, "东京") {
		t.Fatal("expected '东京' in Chinese World Clock")
	}

	// Check Add City Modal in Chinese
	clickID(t, driver, IDBtnAddCityModal)
	snapModalZH := driver.Snapshot()
	if !hasLabel(snapModalZH, "阿富汗") {
		t.Fatal("expected '阿富汗' in Chinese Add City modal")
	}
	if !hasLabel(snapModalZH, "喀布尔") {
		t.Fatal("expected '喀布尔' in Chinese Add City modal")
	}
	clickID(t, driver, IDBtnCloseAddCity)

	// 2. Switch to Japanese
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangJA)

	// Check World Clock in Japanese
	clickID(t, driver, TabWorldID)
	snapJA := driver.Snapshot()
	if !hasLabel(snapJA, "東京") {
		t.Fatal("expected '東京' in Japanese World Clock")
	}
	if !hasLabel(snapJA, "ロンドン") {
		t.Fatal("expected 'ロンドン' in Japanese World Clock")
	}

	// 3. Switch to English
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangEN)

	// Check World Clock in English
	clickID(t, driver, TabWorldID)
	snapEN := driver.Snapshot()
	if !hasLabel(snapEN, "Beijing") {
		t.Fatal("expected 'Beijing' in English World Clock")
	}
	if !hasLabel(snapEN, "China") {
		t.Fatal("expected 'China' in English World Clock")
	}
	if !hasLabel(snapEN, "Tokyo") {
		t.Fatal("expected 'Tokyo' in English World Clock")
	}
}

func TestAllScreensNoFormatOrUTF8Errors(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Record a stopwatch lap so stopwatch lap list is also rendered
	state.Stopwatch.Start(time.Now())
	state.Stopwatch.Lap(time.Now().Add(time.Second))

	tabs := []vui.ID{
		TabClockID,
		TabWorldID,
		TabAlarmsID,
		TabTimersID,
		TabTrackerID,
		TabSleepID,
		TabNewsID,
		TabSettingsID,
	}

	langs := []vui.ID{IDLangZH, IDLangEN, IDLangJA}
	for _, langBtn := range langs {
		clickID(t, driver, TabSettingsID)
		clickID(t, driver, IDSettingsTabLanguage)
		clickID(t, driver, langBtn)

		for _, tabID := range tabs {
			clickID(t, driver, tabID)
			snap := driver.Snapshot()
			for _, lbl := range snap.Labels {
				if strings.Contains(lbl.Value, "%!") {
					t.Fatalf("found fmt.Sprintf format error in label %q on tab %v", lbl.Value, state.ActiveTab)
				}
			}
		}
	}
}

func TestWorldClockSearchInput(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	clickID(t, driver, TabWorldID)
	if !clickID(t, driver, IDBtnAddCityModal) {
		t.Fatal("failed to open Add City modal")
	}

	// Focus the city search input field and type "Shanghai"
	if !clickID(t, driver, IDInputCitySearch) {
		t.Fatal("failed to click city search input field")
	}
	if err := driver.Type("Shanghai"); err != nil {
		t.Fatalf("driver.Type failed: %v", err)
	}
	if state.CitySearchQuery != "Shanghai" {
		t.Fatalf("expected CitySearchQuery 'Shanghai', got %q", state.CitySearchQuery)
	}

	// Add Shanghai from filtered results
	if !clickID(t, driver, CityAddCatalogID("shanghai")) {
		t.Fatal("expected Shanghai in filtered catalog results")
	}
	found := false
	for _, c := range state.World.Cities() {
		if c.CityID == "shanghai" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected shanghai to be added to world clock cities")
	}

	// Re-open modal and test Clear Search button
	clickID(t, driver, IDBtnAddCityModal)
	if !clickID(t, driver, IDBtnClearCitySearch) {
		t.Fatal("expected Clear Search button when query is non-empty")
	}
	if state.CitySearchQuery != "" || state.CitySearchField.Text() != "" {
		t.Fatalf("expected search query to be cleared, got %q", state.CitySearchQuery)
	}
}

func TestWallpaperSlideshowUI(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabWallpaper)

	// Enable wallpaper
	if !clickID(t, driver, IDWallpaperToggle) {
		t.Fatal("failed to toggle wallpaper ON")
	}

	// Load 4-photo sample carousel
	if !clickID(t, driver, IDBtnSampleSlideshow) {
		t.Fatal("failed to click Sample Carousel button")
	}
	if len(state.Wallpaper.Config().ImagePaths) != 4 || !state.Wallpaper.Config().SlideshowEnabled {
		t.Fatalf("expected 4 sample photos with slideshow enabled, got %+v", state.Wallpaper.Config())
	}

	// Test 1m / 5m / 10m / 15m interval buttons
	if !clickID(t, driver, IDWallpaperInt1m) {
		t.Fatal("failed to click 1m interval button")
	}
	if state.Wallpaper.Config().IntervalMinutes != 1 {
		t.Fatalf("expected 1m interval, got %d", state.Wallpaper.Config().IntervalMinutes)
	}
	if !clickID(t, driver, IDWallpaperInt10m) {
		t.Fatal("failed to click 10m interval button")
	}
	if state.Wallpaper.Config().IntervalMinutes != 10 {
		t.Fatalf("expected 10m interval, got %d", state.Wallpaper.Config().IntervalMinutes)
	}
	if !clickID(t, driver, IDWallpaperInt15m) {
		t.Fatal("failed to click 15m interval button")
	}
	if state.Wallpaper.Config().IntervalMinutes != 15 {
		t.Fatalf("expected 15m interval, got %d", state.Wallpaper.Config().IntervalMinutes)
	}
	if !clickID(t, driver, IDWallpaperInt5m) {
		t.Fatal("failed to click 5m interval button")
	}
	if state.Wallpaper.Config().IntervalMinutes != 5 {
		t.Fatalf("expected 5m interval, got %d", state.Wallpaper.Config().IntervalMinutes)
	}

	// Test Sequential / Shuffle order buttons
	if !clickID(t, driver, IDWallpaperOrderSeq) {
		t.Fatal("failed to click Sequential mode button")
	}
	if state.Wallpaper.Config().SlideshowOrder != "sequential" {
		t.Fatalf("expected sequential order, got %s", state.Wallpaper.Config().SlideshowOrder)
	}

	idx0 := state.Wallpaper.Config().CurrentIndex
	if !clickID(t, driver, IDWallpaperSlideNext) {
		t.Fatal("failed to click Next slide button")
	}
	if state.Wallpaper.Config().CurrentIndex != (idx0+1)%4 {
		t.Fatalf("expected index %d after Next, got %d", (idx0+1)%4, state.Wallpaper.Config().CurrentIndex)
	}
	if !clickID(t, driver, IDWallpaperSlidePrev) {
		t.Fatal("failed to click Prev slide button")
	}
	if state.Wallpaper.Config().CurrentIndex != idx0 {
		t.Fatalf("expected index %d after Prev, got %d", idx0, state.Wallpaper.Config().CurrentIndex)
	}

	if !clickID(t, driver, IDWallpaperOrderShuffle) {
		t.Fatal("failed to click Shuffle mode button")
	}
	if state.Wallpaper.Config().SlideshowOrder != "shuffle" {
		t.Fatalf("expected shuffle order, got %s", state.Wallpaper.Config().SlideshowOrder)
	}
}

func TestOnOffAndFullLocalizationZH_JA_EN(t *testing.T) {
	_, driver := setupTestApp(t)
	defer driver.Close()

	// 1. Switch to Chinese (ZH)
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangZH)

	// Main clock: "显示秒数: 是"
	clickID(t, driver, TabClockID)
	if !hasLabel(driver.Snapshot(), "显示秒数: 是") {
		t.Fatal("expected '显示秒数: 是' on Chinese Main Clock")
	}

	// Alarms: "是" / "否"
	clickID(t, driver, TabAlarmsID)
	snapAlarmsZH := driver.Snapshot()
	if !hasLabel(snapAlarmsZH, "是") || !hasLabel(snapAlarmsZH, "否") {
		t.Fatal("expected '是' and '否' on Chinese Alarms screen")
	}
	if !hasLabel(snapAlarmsZH, "早晨起床闹钟") {
		t.Fatal("expected localized default alarm title '早晨起床闹钟' on Chinese Alarms screen")
	}

	// Tracker: "勿扰: 否"
	clickID(t, driver, TabTrackerID)
	if !hasLabel(driver.Snapshot(), "勿扰: 否") {
		t.Fatal("expected '勿扰: 否' on Chinese Tracker screen")
	}

	// Settings Wallpaper: "否" -> click -> "是"
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabWallpaper)
	if !hasLabel(driver.Snapshot(), "否") {
		t.Fatal("expected '否' for disabled wallpaper toggle in Chinese")
	}
	clickID(t, driver, IDWallpaperToggle)
	if !hasLabel(driver.Snapshot(), "是") {
		t.Fatal("expected '是' for enabled wallpaper toggle in Chinese")
	}

	// 2. Switch to Japanese (JA)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangJA)

	clickID(t, driver, TabClockID)
	if !hasLabel(driver.Snapshot(), "秒表示: はい") {
		t.Fatal("expected '秒表示: はい' on Japanese Main Clock")
	}

	clickID(t, driver, TabAlarmsID)
	snapAlarmsJA := driver.Snapshot()
	if !hasLabel(snapAlarmsJA, "はい") || !hasLabel(snapAlarmsJA, "いいえ") {
		t.Fatal("expected 'はい' and 'いいえ' on Japanese Alarms screen")
	}

	// 3. Switch to English (EN)
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangEN)

	clickID(t, driver, TabClockID)
	if !hasLabel(driver.Snapshot(), "Seconds: ON") {
		t.Fatal("expected 'Seconds: ON' on English Main Clock")
	}

	clickID(t, driver, TabAlarmsID)
	snapAlarmsEN := driver.Snapshot()
	if !hasLabel(snapAlarmsEN, "ON") || !hasLabel(snapAlarmsEN, "OFF") {
		t.Fatal("expected 'ON' and 'OFF' on English Alarms screen")
	}
}

func TestWorldClockDuplicateAlertAndPersistence(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Switch to Chinese to test Chinese duplicate reminder
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangZH)

	clickID(t, driver, TabWorldID)
	initialCount := len(state.World.Cities())

	// Open Add City Modal
	if !clickID(t, driver, IDBtnAddCityModal) {
		t.Fatal("failed to open Add City modal")
	}

	// 1. Add Afghanistan first time -> should succeed and close modal
	if !clickID(t, driver, CityAddCatalogID("afghanistan")) {
		t.Fatal("failed to click Add Afghanistan")
	}
	if len(state.World.Cities()) != initialCount+1 {
		t.Fatalf("expected %d cities after adding Afghanistan, got %d", initialCount+1, len(state.World.Cities()))
	}
	if state.ShowAddCityModal {
		t.Fatal("expected Add City modal to close after adding new city")
	}

	// 2. Re-open modal and try adding Afghanistan a second time -> should trigger duplicate alert and NOT add
	if !clickID(t, driver, IDBtnAddCityModal) {
		t.Fatal("failed to re-open Add City modal")
	}
	if !clickID(t, driver, CityAddCatalogID("afghanistan")) {
		t.Fatal("failed to click already-added Afghanistan button")
	}
	if len(state.World.Cities()) != initialCount+1 {
		t.Fatalf("duplicate city should not be added: expected %d, got %d", initialCount+1, len(state.World.Cities()))
	}
	if !state.ShowAddCityModal {
		t.Fatal("expected Add City modal to remain open on duplicate city")
	}
	if state.DuplicateCityAlert == "" {
		t.Fatal("expected DuplicateCityAlert to be populated")
	}
	snapDup := driver.Snapshot()
	if !hasLabel(snapDup, "重复添加提醒") {
		t.Fatal("expected duplicate alert dialog title '重复添加提醒' on screen")
	}

	// Dismiss duplicate alert dialog
	if !clickID(t, driver, IDBtnDismissDuplicateCity) {
		t.Fatal("failed to click dismiss button on duplicate alert dialog")
	}
	if state.DuplicateCityAlert != "" {
		t.Fatalf("expected DuplicateCityAlert to be cleared, got %q", state.DuplicateCityAlert)
	}
	clickID(t, driver, IDBtnCloseAddCity)

	// 3. Mutate various settings across tabs and verify World Clock cities never disappear
	clickID(t, driver, TabClockID)
	clickID(t, driver, IDBtnToggle24H)
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabClock)
	clickID(t, driver, IDCloseBehaviorExit)
	clickID(t, driver, IDSettingsTabWallpaper)
	clickID(t, driver, IDThemeDark)

	// Reload from disk into a fresh Store + AppState
	reloadedStore := storage.NewStore(state.Store.FilePath())
	reloadedState := NewAppState(reloadedStore)
	if len(reloadedState.World.Cities()) != initialCount+1 {
		t.Fatalf("expected %d persisted cities after settings updates, got %d", initialCount+1, len(reloadedState.World.Cities()))
	}
	if !reloadedState.World.HasCity("afghanistan") {
		t.Fatal("expected custom city 'afghanistan' to remain persisted in settings.json")
	}
}

func TestCloseBehaviorSetting(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	if state.Store.Get().CloseBehavior != storage.CloseBehaviorMinimize {
		t.Fatalf("expected default CloseBehavior to be %q, got %q", storage.CloseBehaviorMinimize, state.Store.Get().CloseBehavior)
	}

	clickID(t, driver, TabSettingsID)
	if !clickID(t, driver, IDSettingsTabClock) {
		t.Fatal("failed to click IDSettingsTabClock button")
	}
	if !clickID(t, driver, IDCloseBehaviorExit) {
		t.Fatal("failed to click IDCloseBehaviorExit button")
	}
	if state.Store.Get().CloseBehavior != storage.CloseBehaviorExit {
		t.Fatalf("expected CloseBehavior %q, got %q", storage.CloseBehaviorExit, state.Store.Get().CloseBehavior)
	}

	if !clickID(t, driver, IDCloseBehaviorMinimize) {
		t.Fatal("failed to click IDCloseBehaviorMinimize button")
	}
	if state.Store.Get().CloseBehavior != storage.CloseBehaviorMinimize {
		t.Fatalf("expected CloseBehavior %q, got %q", storage.CloseBehaviorMinimize, state.Store.Get().CloseBehavior)
	}
}

func TestCountryDisplayAudioAndHistoryToday(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// 1. Main Clock should display detected Country/Region badge
	snapMain := driver.Snapshot()
	if !hasLabel(snapMain, "Country/Region:") {
		t.Fatal("expected Country/Region badge on main clock screen")
	}

	// Switch to Chinese and verify "所在国家/地区：" and "历史上的今天"
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangZH)
	clickID(t, driver, TabClockID)
	if !hasLabel(driver.Snapshot(), "所在国家/地区：") {
		t.Fatal("expected '所在国家/地区：' badge on Chinese main clock screen")
	}

	// 2. Test Alarm sound selector & preview in New Alarm Modal
	clickID(t, driver, TabAlarmsID)
	clickID(t, driver, IDBtnNewAlarmModal)
	if !clickID(t, driver, IDBtnAlarmSoundGentle) {
		t.Fatal("failed to click Gentle Chime sound option")
	}
	if state.NewAlarmSound != "Gentle Chime" {
		t.Fatalf("expected NewAlarmSound Gentle Chime, got %s", state.NewAlarmSound)
	}
	if state.Audio.PreviewKey() != "alarm:Gentle Chime" {
		t.Fatalf("expected audio preview key alarm:Gentle Chime, got %s", state.Audio.PreviewKey())
	}
	clickID(t, driver, IDBtnCloseNewAlarm)

	// 3. Test Sleep ambient sound preview & active loop
	clickID(t, driver, TabSleepID)
	if !clickID(t, driver, IDBtnSleepStream) {
		t.Fatal("failed to click Forest Stream ambient sound")
	}
	if state.Audio.PreviewKey() != "ambient:Forest Stream" {
		t.Fatalf("expected audio preview key ambient:Forest Stream, got %s", state.Audio.PreviewKey())
	}
	clickID(t, driver, IDBtnSleepStart)
	if state.Audio.ActiveKey() != "ambient:Forest Stream" {
		t.Fatalf("expected active audio loop ambient:Forest Stream, got %s", state.Audio.ActiveKey())
	}
	clickID(t, driver, IDBtnSleepStop)
	if state.Audio.ActiveKey() != "" {
		t.Fatalf("expected active audio loop stopped, got %s", state.Audio.ActiveKey())
	}

	// 4. Test On This Day in History ("历史上的今天") screen in Chinese
	clickID(t, driver, TabNewsID)
	snapHist := driver.Snapshot()
	if !hasLabel(snapHist, "历史上的今天") {
		t.Fatal("expected '历史上的今天' title on screen")
	}
}

func TestTimerPersistenceCustomCreateEditAndDelete(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	clickID(t, driver, TabTimersID)
	initialTimers := state.TimerMgr.Timers()
	if len(initialTimers) != 3 {
		t.Fatalf("expected 3 default timers initially, got %d", len(initialTimers))
	}

	// 1. Delete the first timer ("timer-tea") and verify it stays deleted after reloading from disk
	if !clickID(t, driver, TimerDeleteID("timer-tea")) {
		t.Fatal("failed to click delete button on timer-tea")
	}
	if len(state.TimerMgr.Timers()) != 2 {
		t.Fatalf("expected 2 timers after deleting timer-tea, got %d", len(state.TimerMgr.Timers()))
	}

	reloadedStore1 := storage.NewStore(state.Store.FilePath())
	reloadedState1 := NewAppState(reloadedStore1)
	if len(reloadedState1.TimerMgr.Timers()) != 2 {
		t.Fatalf("deleted timer reappeared on reload: expected 2 timers, got %d", len(reloadedState1.TimerMgr.Timers()))
	}

	// 2. Edit an existing timer ("timer-break") via the Edit button
	if !clickID(t, driver, TimerEditID("timer-break")) {
		t.Fatal("failed to click Edit button on timer-break")
	}
	if !state.ShowTimerModal || state.EditingTimerID != "timer-break" {
		t.Fatal("expected Timer modal to open in edit mode for timer-break")
	}
	state.TimerNameField.SetText("Custom Reading Task")
	clickID(t, driver, IDBtnTimerMin10Up) // 5m + 10m = 15m
	if !clickID(t, driver, IDBtnSaveTimerModal) {
		t.Fatal("failed to click Save button in Timer modal")
	}

	// Verify edit persisted to disk
	reloadedStore2 := storage.NewStore(state.Store.FilePath())
	reloadedState2 := NewAppState(reloadedStore2)
	var foundEdited bool
	for _, tm := range reloadedState2.TimerMgr.Timers() {
		if tm.ID == "timer-break" && tm.Name == "Custom Reading Task" && tm.Duration == 15*time.Minute {
			foundEdited = true
			break
		}
	}
	if !foundEdited {
		t.Fatal("expected edited timer-break ('Custom Reading Task', 15m) to persist across reload")
	}

	// 3. Create a brand new custom timer via "+ Custom Timer" modal
	if !clickID(t, driver, IDBtnOpenTimerModal) {
		t.Fatal("failed to click + Custom Timer button")
	}
	clickID(t, driver, IDBtnTimerPresetNap) // 20m preset
	state.TimerNameField.SetText("Afternoon Nap")
	if !clickID(t, driver, IDBtnSaveTimerModal) {
		t.Fatal("failed to click Save New Timer button")
	}
	if len(state.TimerMgr.Timers()) != 3 {
		t.Fatalf("expected 3 timers after adding custom timer, got %d", len(state.TimerMgr.Timers()))
	}

	reloadedStore3 := storage.NewStore(state.Store.FilePath())
	reloadedState3 := NewAppState(reloadedStore3)
	if len(reloadedState3.TimerMgr.Timers()) != 3 {
		t.Fatalf("expected 3 persisted timers on reload, got %d", len(reloadedState3.TimerMgr.Timers()))
	}
}

func TestClockDualCalendarAndNonDuplicateText(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Switch language to zh via UI
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	if !clickID(t, driver, IDLangZH) {
		t.Fatal("failed to click ZH language button")
	}
	if state.Lang() != "zh" {
		t.Fatalf("expected zh language, got %s", state.Lang())
	}
	if !clickID(t, driver, TabClockID) {
		t.Fatal("failed to navigate back to TabClock")
	}

	snap := driver.Snapshot()

	// Assert presence of both Solar (公历) and Lunar (农历) calendars
	if !hasLabel(snap, "公历") {
		t.Fatal("expected '公历' (Solar calendar) in Chinese main clock view")
	}
	if !hasLabel(snap, "农历") {
		t.Fatal("expected '农历' (Lunar calendar) in Chinese main clock view")
	}

	// Verify no duplicate location text
	badgeCount := 0
	for _, l := range snap.Labels {
		if strings.Contains(l.Value, "所在国家/地区") {
			badgeCount++
		}
	}
	if badgeCount != 1 {
		t.Fatalf("expected location badge to appear exactly once, got %d times", badgeCount)
	}

	// Switch to Popout mode and verify dual calendar
	if !clickID(t, driver, TabPopoutID) {
		t.Fatal("failed to click Popout tab")
	}
	popSnap := driver.Snapshot()

	if !hasLabel(popSnap, "公历") {
		t.Fatal("expected '公历' in Popout mode")
	}
	if !hasLabel(popSnap, "农历") {
		t.Fatal("expected '农历' in Popout mode")
	}
}

func TestPopoutClockPositionButtons(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Switch to Popout mode
	if !clickID(t, driver, TabPopoutID) {
		t.Fatal("failed to click Popout tab")
	}

	// 1. Initial position should be center
	if state.Store.Get().PopoutPosition != storage.PopoutPosCenter {
		t.Fatalf("expected initial PopoutPosition to be 'center', got '%s'", state.Store.Get().PopoutPosition)
	}

	// 2. Click Top position button
	if !clickID(t, driver, IDBtnPopoutPosTop) {
		t.Fatal("failed to click IDBtnPopoutPosTop button")
	}
	if state.Store.Get().PopoutPosition != storage.PopoutPosTop {
		t.Fatalf("expected PopoutPosition to be 'top', got '%s'", state.Store.Get().PopoutPosition)
	}

	// 3. Click Bottom position button
	if !clickID(t, driver, IDBtnPopoutPosBottom) {
		t.Fatal("failed to click IDBtnPopoutPosBottom button")
	}
	if state.Store.Get().PopoutPosition != storage.PopoutPosBottom {
		t.Fatalf("expected PopoutPosition to be 'bottom', got '%s'", state.Store.Get().PopoutPosition)
	}

	// 4. Click Center position button
	if !clickID(t, driver, IDBtnPopoutPosCenter) {
		t.Fatal("failed to click IDBtnPopoutPosCenter button")
	}
	if state.Store.Get().PopoutPosition != storage.PopoutPosCenter {
		t.Fatalf("expected PopoutPosition to be 'center', got '%s'", state.Store.Get().PopoutPosition)
	}

	// 5. Restore back to main window and check settings screen
	if !clickID(t, driver, IDBtnExitPopout) {
		t.Fatal("failed to click Restore button")
	}
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabClock)

	if !clickID(t, driver, IDOptPopoutPosTop) {
		t.Fatal("failed to click IDOptPopoutPosTop in settings")
	}
	if state.Store.Get().PopoutPosition != storage.PopoutPosTop {
		t.Fatalf("expected PopoutPosition to be 'top' via settings, got '%s'", state.Store.Get().PopoutPosition)
	}
}

func TestHolidayDisplayAndSettingsToggle(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Switch language to zh
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangZH)
	clickID(t, driver, TabClockID)

	// Verify holiday badge ("今日节日：...") is displayed when ShowHolidays is true
	snap := driver.Snapshot()
	if !hasLabel(snap, "今日节日：") {
		t.Fatal("expected holiday badge '今日节日：' to be displayed when ShowHolidays is true")
	}

	// Navigate to Settings -> Clock and toggle ShowHolidays OFF
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabClock)
	if !clickID(t, driver, IDOptShowHolidays) {
		t.Fatal("failed to click IDOptShowHolidays toggle button")
	}
	if state.Store.Get().ShowHolidays {
		t.Fatal("expected ShowHolidays to be false")
	}

	// Verify holiday badge is now hidden
	clickID(t, driver, TabClockID)
	snapHidden := driver.Snapshot()
	if hasLabel(snapHidden, "今日节日：") {
		t.Fatal("expected holiday badge to be hidden when ShowHolidays is false")
	}
}

func TestHolidayDisplayWhenNoHoliday(t *testing.T) {
	badgeZH := clock.FormatHolidayBadge(nil, "zh")
	if !strings.Contains(badgeZH, "今日节日：无") {
		t.Fatalf("expected '今日节日：无' when no holidays, got '%s'", badgeZH)
	}

	badgeEN := clock.FormatHolidayBadge(nil, "en")
	if !strings.Contains(badgeEN, "Today's Holiday: None") {
		t.Fatalf("expected 'Today's Holiday: None' when no holidays, got '%s'", badgeEN)
	}

	badgeJA := clock.FormatHolidayBadge(nil, "ja")
	if !strings.Contains(badgeJA, "なし") {
		t.Fatalf("expected 'なし' when no holidays, got '%s'", badgeJA)
	}
}

func TestAutoStartSettingToggle(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	var hookCalls []bool
	state.AutostartHook = func(enabled bool) error {
		hookCalls = append(hookCalls, enabled)
		return nil
	}

	// Default AutoStart should be false
	if state.Store.Get().AutoStart {
		t.Fatal("expected AutoStart to be false by default")
	}

	// Navigate to Settings -> Clock sub-tab
	clickID(t, driver, TabSettingsID)
	if !clickID(t, driver, IDSettingsTabClock) {
		t.Fatal("failed to click Clock settings sub-tab")
	}

	// Click AutoStart toggle -> ON
	if !clickID(t, driver, IDOptAutoStart) {
		t.Fatal("failed to click IDOptAutoStart button")
	}
	if !state.Store.Get().AutoStart {
		t.Fatal("expected AutoStart to be true after clicking toggle")
	}
	if len(hookCalls) != 1 || !hookCalls[0] {
		t.Fatalf("expected AutostartHook called with true, got %v", hookCalls)
	}

	// Click explicit Disable button -> OFF
	if !clickID(t, driver, IDOptAutoStartDisable) {
		t.Fatal("failed to click IDOptAutoStartDisable button")
	}
	if state.Store.Get().AutoStart {
		t.Fatal("expected AutoStart to be false after clicking Disable button")
	}
	if len(hookCalls) != 2 || hookCalls[1] {
		t.Fatalf("expected AutostartHook called with false, got %v", hookCalls)
	}

	// Click explicit Enable button -> ON
	if !clickID(t, driver, IDOptAutoStartEnable) {
		t.Fatal("failed to click IDOptAutoStartEnable button")
	}
	if !state.Store.Get().AutoStart {
		t.Fatal("expected AutoStart to be true after clicking Enable button")
	}
	if len(hookCalls) != 3 || !hookCalls[2] {
		t.Fatalf("expected AutostartHook called with true, got %v", hookCalls)
	}
}

func TestHeaderAndFooterVisibilityToggle(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Switch to Settings -> Language subtab
	if !clickID(t, driver, TabSettingsID) {
		t.Fatal("failed to click TabSettingsID")
	}
	if !clickID(t, driver, IDSettingsTabLanguage) {
		t.Fatal("failed to click IDSettingsTabLanguage")
	}

	// 1. Initially both should be false (visible)
	if state.Store.Get().HideHeader {
		t.Fatal("expected default HideHeader to be false")
	}
	if state.Store.Get().HideFooter {
		t.Fatal("expected default HideFooter to be false")
	}

	// 2. Click Hide Header
	if !clickID(t, driver, IDOptHeaderHide) {
		t.Fatal("failed to click IDOptHeaderHide")
	}
	if !state.Store.Get().HideHeader {
		t.Fatal("expected HideHeader to be true after clicking IDOptHeaderHide")
	}

	// 3. Click Show Header
	if !clickID(t, driver, IDOptHeaderShow) {
		t.Fatal("failed to click IDOptHeaderShow")
	}
	if state.Store.Get().HideHeader {
		t.Fatal("expected HideHeader to be false after clicking IDOptHeaderShow")
	}

	// 4. Click Hide Footer
	if !clickID(t, driver, IDOptFooterHide) {
		t.Fatal("failed to click IDOptFooterHide")
	}
	if !state.Store.Get().HideFooter {
		t.Fatal("expected HideFooter to be true after clicking IDOptFooterHide")
	}

	// 5. Click Show Footer
	if !clickID(t, driver, IDOptFooterShow) {
		t.Fatal("failed to click IDOptFooterShow")
	}
	if state.Store.Get().HideFooter {
		t.Fatal("expected HideFooter to be false after clicking IDOptFooterShow")
	}

	// 6. Test One-Click Hide All Status Bars
	if !clickID(t, driver, IDOptBarsAllHide) {
		t.Fatal("failed to click IDOptBarsAllHide")
	}
	if !state.Store.Get().HideHeader || !state.Store.Get().HideFooter {
		t.Fatal("expected both HideHeader and HideFooter to be true")
	}

	// 7. Verify Popout mode reflects hidden bars
	state.IsPopout = true
	state.ActiveTab = TabPopout
	_ = driver.Resize(900, 1201)

	// In Popout mode with HideHeader=true, full header (IDBtnPinPopout) is hidden, but mini restore button (IDBtnExitPopout) remains
	snap := driver.Snapshot()
	hasPin := false
	hasExit := false
	has24H := false
	for _, hit := range snap.Hits {
		if hit.ID == IDBtnPinPopout {
			hasPin = true
		}
		if hit.ID == IDBtnExitPopout {
			hasExit = true
		}
		if hit.ID == IDBtnToggle24H {
			has24H = true
		}
	}
	if hasPin {
		t.Fatal("expected IDBtnPinPopout to be hidden when HideHeader is true")
	}
	if !hasExit {
		t.Fatal("expected IDBtnExitPopout to remain accessible for restoring from popout")
	}
	if has24H {
		t.Fatal("expected IDBtnToggle24H to be hidden when HideFooter is true")
	}

	// Click Exit Popout to return to clock
	if !clickID(t, driver, IDBtnExitPopout) {
		t.Fatal("failed to click IDBtnExitPopout in immersive popout mode")
	}
	// Use mini settings button from clock (since header is hidden) to navigate to settings
	if !clickID(t, driver, IDBtnClockMiniSettings) {
		t.Fatal("failed to click IDBtnClockMiniSettings")
	}
	clickID(t, driver, IDSettingsTabLanguage)

	// 8. Test One-Click Show All Status Bars
	if !clickID(t, driver, IDOptBarsAllShow) {
		t.Fatal("failed to click IDOptBarsAllShow")
	}
	if state.Store.Get().HideHeader || state.Store.Get().HideFooter {
		t.Fatal("expected both HideHeader and HideFooter to be false")
	}
}

func TestEdgeBarAutoShrinkAndRestoreButton(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Switch to Chinese UI
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangZH)
	clickID(t, driver, TabClockID)

	// 1. Window in normal interior screen position -> should NOT be in Edge Bar mode
	state.HandleFrameChange(vui.Rect{X: 220, Y: 160, Width: 840, Height: 580})
	if state.IsEdgeBar {
		t.Fatal("expected IsEdgeBar to be false when window is away from screen edges")
	}

	// 2. Window approaches top edge of screen -> should automatically shrink into Edge Bar mode
	state.HandleFrameChange(vui.Rect{X: 220, Y: 8, Width: 840, Height: 580})
	if !state.IsEdgeBar {
		t.Fatal("expected IsEdgeBar to be true when window approaches top screen edge")
	}
	if state.DockedEdge != "top" {
		t.Fatalf("expected DockedEdge 'top', got '%s'", state.DockedEdge)
	}

	// Advance frame to repaint Edge Bar UI
	driver.Advance(50 * time.Millisecond)
	snapBar := driver.Snapshot()
	if !hasLabel(snapBar, "恢复正常") {
		t.Fatal("expected Edge Bar to display '恢复正常' restore button")
	}
	if !hasLabel(snapBar, "公历") || !hasLabel(snapBar, "农历") {
		t.Fatal("expected Edge Bar to concisely display solar and lunar date")
	}
	if hasLabel(snapBar, "世界时钟") {
		t.Fatal("expected main navigation bar to be hidden while in compact Edge Bar mode")
	}

	// 3. Click '恢复正常' button -> restores normal full UI
	if !clickID(t, driver, IDBtnRestoreEdgeBar) {
		t.Fatal("failed to click IDBtnRestoreEdgeBar ('恢复正常')")
	}
	if state.IsEdgeBar {
		t.Fatal("expected IsEdgeBar to be false after clicking restore button")
	}
	snapRestored := driver.Snapshot()
	if !hasLabel(snapRestored, "世界时钟") {
		t.Fatal("expected normal navigation bar to be restored after clicking restore button")
	}

	// 4. Verify debounce: while still near edge right after restore, it doesn't immediately re-collapse
	state.HandleFrameChange(vui.Rect{X: 220, Y: 10, Width: 840, Height: 580})
	if state.IsEdgeBar {
		t.Fatal("expected debounce to prevent immediate re-shrinking before moving away from edge")
	}

	// 5. Move window back inside screen, then move near left edge -> shrinks again
	state.HandleFrameChange(vui.Rect{X: 240, Y: 180, Width: 840, Height: 580})
	state.HandleFrameChange(vui.Rect{X: 4, Y: 180, Width: 840, Height: 580})
	if !state.IsEdgeBar || state.DockedEdge != "left" {
		t.Fatalf("expected window to shrink at left edge after moving away and approaching left edge, got IsEdgeBar=%v edge=%s", state.IsEdgeBar, state.DockedEdge)
	}

	// Restore again via button
	driver.Advance(50 * time.Millisecond)
	if !clickID(t, driver, IDBtnRestoreEdgeBar) {
		t.Fatal("failed to click restore button from left edge bar")
	}

	// 6. Test quick Edge Bar button on main clock dock
	if !clickID(t, driver, IDBtnQuickEdgeBar) {
		t.Fatal("failed to click IDBtnQuickEdgeBar on main clock dock")
	}
	if !state.IsEdgeBar {
		t.Fatal("expected IsEdgeBar to be true after clicking quick edge bar button")
	}
	if !clickID(t, driver, IDBtnRestoreEdgeBar) {
		t.Fatal("failed to restore from quick edge bar")
	}

	// 7. Test explicit Disable ('🚫 不启用') and Enable ('✅ 启用状态条收缩') buttons in Settings -> Clock
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabClock)
	if !clickID(t, driver, IDOptEdgeDockDisable) {
		t.Fatal("failed to click IDOptEdgeDockDisable ('🚫 不启用') in settings")
	}
	if state.Store.Get().EdgeDockEnabled {
		t.Fatal("expected EdgeDockEnabled to be false after clicking '🚫 不启用'")
	}
	state.HandleFrameChange(vui.Rect{X: 240, Y: 180, Width: 840, Height: 580})
	state.HandleFrameChange(vui.Rect{X: 240, Y: 4, Width: 840, Height: 580})
	if state.IsEdgeBar {
		t.Fatal("expected IsEdgeBar to remain false when EdgeDockEnabled is disabled in settings")
	}

	// Re-enable via explicit Enable ('✅ 启用状态条收缩') button
	if !clickID(t, driver, IDOptEdgeDockEnable) {
		t.Fatal("failed to click IDOptEdgeDockEnable ('✅ 启用状态条收缩') in settings")
	}
	if !state.Store.Get().EdgeDockEnabled {
		t.Fatal("expected EdgeDockEnabled to be true after clicking '✅ 启用状态条收缩'")
	}

	// Test preview button ('➖ 立即缩为状态条') directly inside Settings
	if !clickID(t, driver, IDBtnSettingsEdgeBarPreview) {
		t.Fatal("failed to click IDBtnSettingsEdgeBarPreview in settings")
	}
	if !state.IsEdgeBar {
		t.Fatal("expected IsEdgeBar to be true after clicking preview button in settings")
	}
	if !clickID(t, driver, IDBtnRestoreEdgeBar) {
		t.Fatal("failed to restore from settings preview edge bar")
	}
}

func TestProSubscriptionAndTrialCenter(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Switch to Chinese UI first
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangZH)
	clickID(t, driver, TabClockID)

	// 1. Default state is 7-Day Free Trial, with all features 100% unlocked without payment
	if state.License.Plan() != license.PlanTrial {
		t.Fatalf("expected initial plan to be PlanTrial, got %s", state.License.Plan())
	}
	if !state.License.IsFeatureUnlocked(state.Now) {
		t.Fatal("expected all features to be unlocked in full test build")
	}
	if !hasLabel(driver.Snapshot(), "试用 7 天") {
		t.Fatal("expected top navigation bar to display '👑 试用 7 天 · Pro' button")
	}

	// 2. Click prominent top navigation bar Pro button -> opens Pro Subscription & Trial Modal
	if !clickID(t, driver, TabProID) {
		t.Fatal("failed to click TabProID in top navigation bar")
	}
	if !state.ShowProModal {
		t.Fatal("expected ShowProModal to be true after clicking TabProID")
	}
	snapPro := driver.Snapshot()
	if !hasLabel(snapPro, "¥3 / 月") || !hasLabel(snapPro, "¥28 / 年") {
		t.Fatal("expected both ¥3/月 and ¥28/年 subscription plans on Pro modal")
	}

	// 3. Click Monthly subscription (¥3/月) -> activates immediately for free testing and persists
	if !clickID(t, driver, IDBtnProSubMonthly) {
		t.Fatal("failed to click IDBtnProSubMonthly")
	}
	if state.License.Plan() != license.PlanMonthly || !state.License.IsProActive() {
		t.Fatalf("expected PlanMonthly active, got %s", state.License.Plan())
	}
	if state.Store.Get().License.Plan != license.PlanMonthly {
		t.Fatalf("expected persisted plan %s, got %s", license.PlanMonthly, state.Store.Get().License.Plan)
	}
	if !hasLabel(driver.Snapshot(), "👑 Pro 会员") {
		t.Fatal("expected top navigation button to update to '👑 Pro 会员'")
	}

	// 4. Click Yearly subscription (¥28/年) -> activates immediately for free testing and persists
	if !clickID(t, driver, IDBtnProSubYearly) {
		t.Fatal("failed to click IDBtnProSubYearly")
	}
	if state.License.Plan() != license.PlanYearly {
		t.Fatalf("expected PlanYearly active, got %s", state.License.Plan())
	}
	if state.Store.Get().License.Plan != license.PlanYearly {
		t.Fatalf("expected persisted plan %s, got %s", license.PlanYearly, state.Store.Get().License.Plan)
	}

	// 5. Test Full Test Mode & 7-Day Trial Simulation buttons + Restore Purchases
	if !clickID(t, driver, IDBtnProSimFullTest) {
		t.Fatal("failed to click IDBtnProSimFullTest")
	}
	if state.License.Plan() != license.PlanFullTest {
		t.Fatalf("expected PlanFullTest, got %s", state.License.Plan())
	}
	if !clickID(t, driver, IDBtnProSimTrial) {
		t.Fatal("failed to click IDBtnProSimTrial")
	}
	if state.License.Plan() != license.PlanTrial {
		t.Fatalf("expected PlanTrial after simulation reset, got %s", state.License.Plan())
	}
	if !clickID(t, driver, IDBtnProRestore) {
		t.Fatal("failed to click IDBtnProRestore")
	}
	if state.License.LastNotice() != "restored" {
		t.Fatalf("expected notice restored, got %s", state.License.LastNotice())
	}

	// 6. Close Pro Modal via close button
	if !clickID(t, driver, IDBtnCloseProModal) {
		t.Fatal("failed to click IDBtnCloseProModal")
	}
	if state.ShowProModal {
		t.Fatal("expected ShowProModal to be false after closing modal")
	}

	// 7. Navigate to Settings -> 👑 会员 sub-tab and verify Pro Center renders inside Settings
	clickID(t, driver, TabSettingsID)
	if !clickID(t, driver, IDSettingsTabPro) {
		t.Fatal("failed to click IDSettingsTabPro ('👑 会员') in Settings")
	}
	if state.SettingsTab != "pro" {
		t.Fatalf("expected SettingsTab 'pro', got '%s'", state.SettingsTab)
	}
	snapSettingsPro := driver.Snapshot()
	if !hasLabel(snapSettingsPro, "¥3 / 月") || !hasLabel(snapSettingsPro, "¥28 / 年") {
		t.Fatal("expected subscription plans to render inside Settings -> Pro sub-tab")
	}
}

func TestTrialExpiredLockoutEnforced(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Switch to Chinese UI
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangZH)
	clickID(t, driver, TabClockID)

	baseNow := time.Now()
	mockNow := baseNow
	state.NowHook = func() time.Time {
		return mockNow
	}

	// 1. Initial state (Day 0): 7-Day trial is active, app functions normally
	if !state.License.IsFeatureUnlocked(mockNow) {
		t.Fatal("expected trial to be active initially")
	}
	if !hasLabel(driver.Snapshot(), "试用 7 天") {
		t.Fatal("expected '试用 7 天' banner in header")
	}

	// 2. Advance time past 7 days (e.g. 8 days)
	mockNow = baseNow.Add(8 * 24 * time.Hour)

	// Advance driver to paint a new frame with updated time
	_ = driver.Resize(900, 1201)

	// Verify license manager reports trial expired and locked
	if state.License.TrialDaysRemaining(mockNow) != 0 {
		t.Fatalf("expected 0 trial days remaining, got %d", state.License.TrialDaysRemaining(mockNow))
	}
	if state.License.IsFeatureUnlocked(mockNow) {
		t.Fatal("expected IsFeatureUnlocked to be false after 8 days")
	}

	// 3. Verify that the UI displays the Lockout View and hides all normal tabs
	snapLockout := driver.Snapshot()
	if !hasLabel(snapLockout, "7 天全功能免费试用已结束") {
		var lbls []string
		for _, l := range snapLockout.Labels {
			lbls = append(lbls, l.Value)
		}
		t.Fatalf("expected lockout header '7 天全功能免费试用已结束', got labels: %v", lbls)
	}
	// Normal navigation tabs should be completely absent from hit targets
	for _, hit := range snapLockout.Hits {
		if hit.ID == TabClockID || hit.ID == TabWorldID || hit.ID == TabAlarmsID || hit.ID == TabTimersID {
			t.Fatalf("normal navigation tab %v should NOT be rendered when locked out", hit.ID)
		}
	}

	// 4. Verify clicking '退出应用程序' (IDBtnLockoutExit) triggers application exit
	if !clickID(t, driver, IDBtnLockoutExit) {
		t.Fatal("failed to click IDBtnLockoutExit")
	}
	if !state.ForceQuit {
		t.Fatal("expected ForceQuit to be true after clicking exit on lockout screen")
	}

	// 5. Test unlock by subscribing: user clicks '开通按月订阅' (IDBtnProSubMonthly)
	if !clickID(t, driver, IDBtnProSubMonthly) {
		t.Fatal("failed to click IDBtnProSubMonthly on lockout screen")
	}
	if !state.License.IsFeatureUnlocked(mockNow) {
		t.Fatal("expected features to be unlocked immediately after subscribing")
	}

	// Advance driver to paint the unlocked frame
	_ = driver.Resize(900, 1200)

	// Normal navigation tabs should now be restored
	snapUnlocked := driver.Snapshot()
	if hasLabel(snapUnlocked, "7 天全功能免费试用已结束") {
		t.Fatal("lockout screen should disappear after subscribing")
	}
	var clockTabFound bool
	for _, hit := range snapUnlocked.Hits {
		if hit.ID == TabClockID {
			clockTabFound = true
			break
		}
	}
	if !clockTabFound {
		t.Fatal("expected TabClockID to be restored and interactive after subscribing")
	}
}

func TestIPLocationAndWeatherDisplay(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Switch to Chinese UI
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangZH)
	clickID(t, driver, TabClockID)

	// Inject IP location
	state.IPLoc.SetCached(iploc.Info{
		IP:          "120.244.11.22",
		Country:     "中国",
		CountryCode: "CN",
		RegionName:  "广东",
		City:        "广州",
		Latitude:    23.1291,
		Longitude:   113.2644,
		IsSuccess:   true,
	}, "zh")

	// Repaint
	_ = driver.Resize(900, 1202)
	snap := driver.Snapshot()

	// 1. Verify IP country badge with full IP
	if !hasLabel(snap, "IP 所在国家：中国") || !hasLabel(snap, "120.244.11.22") {
		var lbls []string
		for _, l := range snap.Labels {
			lbls = append(lbls, l.Value)
		}
		t.Fatalf("expected IP country badge with full IP, got labels: %v", lbls)
	}

	// 2. Verify IP Country, Province, and City in front of the weather temperature
	if !hasLabel(snap, "中国 广东 广州") {
		t.Fatal("expected '中国 广东 广州' in front of weather temperature")
	}
}

func TestPreciseTimeSteppersInAlarmTimerSleepAndTracker(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	// Switch to Chinese UI
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	clickID(t, driver, IDLangZH)

	// 1. Alarm Creation Modal: test +1分钟 / -1分钟 and hour/midnight carry & borrow
	clickID(t, driver, TabAlarmsID)
	if !clickID(t, driver, IDBtnNewAlarmModal) {
		t.Fatal("failed to open New Alarm modal")
	}
	if !clickID(t, driver, IDBtnPresetMorning) { // 07:30
		t.Fatal("failed to click Morning preset")
	}
	if !clickID(t, driver, IDBtnAlarmMin1Up) {
		t.Fatal("failed to click +1分钟 button in Alarm modal")
	}
	if state.NewAlarmHour != 7 || state.NewAlarmMinute != 31 {
		t.Fatalf("expected 07:31 after +1m, got %02d:%02d", state.NewAlarmHour, state.NewAlarmMinute)
	}
	if !clickID(t, driver, IDBtnAlarmMin1Down) {
		t.Fatal("failed to click -1分钟 button in Alarm modal")
	}
	if state.NewAlarmHour != 7 || state.NewAlarmMinute != 30 {
		t.Fatalf("expected 07:30 after -1m, got %02d:%02d", state.NewAlarmHour, state.NewAlarmMinute)
	}

	// Test boundary carry: 07:59 + 1m -> 08:00, and 08:00 - 1m -> 07:59
	state.NewAlarmHour, state.NewAlarmMinute = 7, 59
	driver.Advance(10 * time.Millisecond)
	clickID(t, driver, IDBtnAlarmMin1Up)
	if state.NewAlarmHour != 8 || state.NewAlarmMinute != 0 {
		t.Fatalf("expected 07:59 + 1m -> 08:00, got %02d:%02d", state.NewAlarmHour, state.NewAlarmMinute)
	}
	clickID(t, driver, IDBtnAlarmMin1Down)
	if state.NewAlarmHour != 7 || state.NewAlarmMinute != 59 {
		t.Fatalf("expected 08:00 - 1m -> 07:59, got %02d:%02d", state.NewAlarmHour, state.NewAlarmMinute)
	}

	// Test midnight wrap: 23:59 + 1m -> 00:00, and 00:00 - 1m -> 23:59
	if h, m := StepAlarmTime(23, 59, 0, 1); h != 0 || m != 0 {
		t.Fatalf("expected 23:59 + 1m -> 00:00, got %02d:%02d", h, m)
	}
	if h, m := StepAlarmTime(0, 0, 0, -1); h != 23 || m != 59 {
		t.Fatalf("expected 00:00 - 1m -> 23:59, got %02d:%02d", h, m)
	}
	clickID(t, driver, IDBtnCloseNewAlarm)

	// 2. Timer Creation Modal: test +1m/-1m, +1s/-1s, and carry/borrow across minute & hour boundaries
	clickID(t, driver, TabTimersID)
	if !clickID(t, driver, IDBtnOpenTimerModal) {
		t.Fatal("failed to open Timer modal")
	}
	// Initial is 00:15:00 -> click +1s -> 00:15:01 -> click -1s -> 00:15:00 -> click -1s -> 00:14:59
	if !clickID(t, driver, IDBtnTimerSec1Up) {
		t.Fatal("failed to click +1秒 in Timer modal")
	}
	if state.TimerModalMinutes != 15 || state.TimerModalSeconds != 1 {
		t.Fatalf("expected 00:15:01, got %02d:%02d:%02d", state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds)
	}
	clickID(t, driver, IDBtnTimerSec1Down)
	clickID(t, driver, IDBtnTimerSec1Down)
	if state.TimerModalMinutes != 14 || state.TimerModalSeconds != 59 {
		t.Fatalf("expected 00:15:00 - 1s -> 00:14:59 (minute borrow), got %02d:%02d:%02d", state.TimerModalHours, state.TimerModalMinutes, state.TimerModalSeconds)
	}
	clickID(t, driver, IDBtnTimerSec1Up) // back to 00:15:00

	// Test hour carry & borrow: 00:59:00 + 1m -> 01:00:00, and 01:00:00 - 1m -> 00:59:00
	if h, m, s := StepTimerHMS(0, 59, 0, 60); h != 1 || m != 0 || s != 0 {
		t.Fatalf("expected 00:59:00 + 1m -> 01:00:00, got %02d:%02d:%02d", h, m, s)
	}
	if h, m, s := StepTimerHMS(1, 0, 0, -60); h != 0 || m != 59 || s != 0 {
		t.Fatalf("expected 01:00:00 - 1m -> 00:59:00, got %02d:%02d:%02d", h, m, s)
	}
	if h, m, s := StepTimerHMS(0, 0, 10, -15); h != 0 || m != 0 || s != 0 {
		t.Fatalf("expected 00:00:10 - 15s to clamp to 00:00:00, got %02d:%02d:%02d", h, m, s)
	}
	clickID(t, driver, IDBtnCloseTimerModal)

	// 3. Sleep Mode: test +1分钟 / -1分钟 / +5分钟 / -5分钟 steppers
	clickID(t, driver, TabSleepID)
	clickID(t, driver, IDBtnSleep30m)
	if !clickID(t, driver, IDBtnSleepMin1Up) {
		t.Fatal("failed to click +1分钟 in Sleep Mode")
	}
	if state.SleepDuration != 31*time.Minute {
		t.Fatalf("expected 31m sleep duration after +1m, got %v", state.SleepDuration)
	}
	if !clickID(t, driver, IDBtnSleepMin1Down) {
		t.Fatal("failed to click -1分钟 in Sleep Mode")
	}
	if state.SleepDuration != 30*time.Minute {
		t.Fatalf("expected 30m sleep duration after -1m, got %v", state.SleepDuration)
	}

	// 4. Productivity Tracker (Pomodoro): test +1分钟 / -1分钟 steppers when Idle
	clickID(t, driver, TabTrackerID)
	if !clickID(t, driver, IDBtnTrackerMin1Up) {
		t.Fatal("failed to click +1分钟 in Tracker")
	}
	if state.Tracker.Config().WorkDuration != 26*time.Minute || state.Tracker.FormatRemaining() != "26:00" {
		t.Fatalf("expected 26:00 work duration after +1m, got %v (%s)", state.Tracker.Config().WorkDuration, state.Tracker.FormatRemaining())
	}
	if !clickID(t, driver, IDBtnTrackerMin1Down) {
		t.Fatal("failed to click -1分钟 in Tracker")
	}
	if state.Tracker.Config().WorkDuration != 25*time.Minute || state.Tracker.FormatRemaining() != "25:00" {
		t.Fatalf("expected 25:00 work duration after -1m, got %v (%s)", state.Tracker.Config().WorkDuration, state.Tracker.FormatRemaining())
	}
}

func TestTopmostPinAcrossAllModes(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	var topmostCalls []bool
	state.TopmostHook = func(enabled bool) {
		topmostCalls = append(topmostCalls, enabled)
	}

	// 1. Toggle Topmost from Top Header Bar (`IDBtnHeaderTopmost`)
	if state.Store.Get().Topmost {
		t.Fatal("expected initial Topmost to be false")
	}
	if !clickID(t, driver, IDBtnHeaderTopmost) {
		t.Fatal("failed to click IDBtnHeaderTopmost in top navigation bar")
	}
	if !state.Store.Get().Topmost || len(topmostCalls) != 1 || !topmostCalls[0] {
		t.Fatalf("expected Topmost true and TopmostHook(true), got Topmost=%v calls=%v", state.Store.Get().Topmost, topmostCalls)
	}

	// 2. Toggle Topmost inside Popout Mode (`IDBtnPinPopout`)
	if !clickID(t, driver, TabPopoutID) {
		t.Fatal("failed to enter Popout mode")
	}
	if !clickID(t, driver, IDBtnPinPopout) {
		t.Fatal("failed to click IDBtnPinPopout in Popout mode")
	}
	if state.Store.Get().Topmost || len(topmostCalls) != 2 || topmostCalls[1] {
		t.Fatalf("expected Topmost false and TopmostHook(false), got Topmost=%v calls=%v", state.Store.Get().Topmost, topmostCalls)
	}
	clickID(t, driver, IDBtnExitPopout)

	// 3. Toggle Topmost inside Settings -> Clock (`IDOptTopmost`)
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabClock)
	if !clickID(t, driver, IDOptTopmost) {
		t.Fatal("failed to click IDOptTopmost in Clock settings")
	}
	if !state.Store.Get().Topmost || len(topmostCalls) != 3 || !topmostCalls[2] {
		t.Fatalf("expected Topmost true and TopmostHook(true) from settings, got Topmost=%v calls=%v", state.Store.Get().Topmost, topmostCalls)
	}

	// 4. Toggle Topmost inside Edge Bar mode (`IDBtnEdgeBarTopmost`)
	clickID(t, driver, TabClockID)
	clickID(t, driver, IDBtnQuickEdgeBar)
	if !clickID(t, driver, IDBtnEdgeBarTopmost) {
		t.Fatal("failed to click IDBtnEdgeBarTopmost in Edge Bar mode")
	}
	if state.Store.Get().Topmost || len(topmostCalls) != 4 || topmostCalls[3] {
		t.Fatalf("expected Topmost false and TopmostHook(false) from edge bar, got Topmost=%v calls=%v", state.Store.Get().Topmost, topmostCalls)
	}
	clickID(t, driver, IDBtnRestoreEdgeBar)
}

func TestSystemFontSelectionAndCrossPCSafeFallback(t *testing.T) {
	state, driver := setupTestApp(t)
	defer driver.Close()

	var appliedFonts []string
	state.FontFamilyHook = func(family string) {
		appliedFonts = append(appliedFonts, family)
	}

	// 1. Verify symbol / dingbat font blacklist and readable font filter
	rawFonts := []string{"Wingdings", "Webdings", "Marlett", "Symbol", "Segoe MDL2 Assets", "@SimSun", "Microsoft YaHei", "Consolas", "Arial"}
	filtered := FilterReadableSystemFonts(rawFonts)
	if len(filtered) != 3 {
		t.Fatalf("expected 3 readable fonts after filtering symbol/vertical fonts, got %v", filtered)
	}
	for _, f := range filtered {
		if f == "Wingdings" || f == "Webdings" || f == "Marlett" || f == "Symbol" || f == "@SimSun" {
			t.Fatalf("symbol or vertical font should have been filtered out: %s", f)
		}
	}

	// 2. Verify cross-PC safe font resolution (installed font vs missing font from another PC)
	if got := ResolveSafeFontFamily("microsoft yahei", filtered); got != "Microsoft YaHei" {
		t.Fatalf("expected canonical 'Microsoft YaHei', got %q", got)
	}
	if got := ResolveSafeFontFamily("UninstalledCustomFontOnAnotherPC", filtered); got != "" {
		t.Fatalf("expected missing font to safely fall back to empty string (system default), got %q", got)
	}
	if got := ResolveSafeFontFamily("Wingdings", filtered); got != "" {
		t.Fatalf("expected symbol font to safely fall back to empty string, got %q", got)
	}

	// 3. Navigate to Settings -> 🌐 语言 and select a curated system font ("Microsoft YaHei")
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
	if !clickID(t, driver, CuratedFontID("Microsoft YaHei")) {
		t.Fatal("failed to click CuratedFontID('Microsoft YaHei') in Language settings")
	}
	if state.Store.Get().FontFamily != "Microsoft YaHei" {
		t.Fatalf("expected persisted FontFamily 'Microsoft YaHei', got %q", state.Store.Get().FontFamily)
	}
	if len(appliedFonts) == 0 || appliedFonts[len(appliedFonts)-1] != "Microsoft YaHei" {
		t.Fatalf("expected FontFamilyHook called with 'Microsoft YaHei', got %v", appliedFonts)
	}

	// 4. Expand full system font library (`IDBtnToggleSystemFontList`) and select "Consolas"
	if !clickID(t, driver, IDBtnToggleSystemFontList) {
		t.Fatal("failed to click IDBtnToggleSystemFontList to expand full system font list")
	}
	if !state.ShowAllSystemFonts {
		t.Fatal("expected ShowAllSystemFonts to be true after clicking expand button")
	}
	if !clickID(t, driver, SystemFontItemID("Consolas")) {
		t.Fatal("failed to click SystemFontItemID('Consolas') in expanded system font list")
	}
	if state.Store.Get().FontFamily != "Consolas" {
		t.Fatalf("expected persisted FontFamily 'Consolas', got %q", state.Store.Get().FontFamily)
	}

	// 5. Reset to System Default font (`IDBtnFontDefault`)
	if !clickID(t, driver, IDBtnFontDefault) {
		t.Fatal("failed to click IDBtnFontDefault ('系统默认')")
	}
	if state.Store.Get().FontFamily != "" {
		t.Fatalf("expected FontFamily to reset to '', got %q", state.Store.Get().FontFamily)
	}

	// 6. Simulate migrating settings.json from another PC with an uninstalled font -> SyncFontFamily auto-recovers
	_ = state.Store.Update(func(s *storage.Settings) {
		s.FontFamily = "NonExistentPCFont_2099"
	})
	state.SyncFontFamily()
	if state.Store.Get().FontFamily != "" {
		t.Fatalf("expected SyncFontFamily to automatically fall back to '' when font is missing on current PC, got %q", state.Store.Get().FontFamily)
	}
}
