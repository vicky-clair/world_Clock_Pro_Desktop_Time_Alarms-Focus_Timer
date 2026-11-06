package ui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"worldclock/pkg/alarm"
	"worldclock/pkg/clock"
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
	cfg.Height = 1200
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

	// 3. Toggle Style Digital -> Analog
	if state.Store.Get().ClockStyle != string(clock.StyleDigital) {
		t.Fatal("expected default digital clock style")
	}
	if !clickID(t, driver, IDBtnToggleStyle) {
		t.Fatal("failed to click style toggle button")
	}
	if state.Store.Get().ClockStyle != string(clock.StyleAnalog) {
		t.Fatal("expected style to switch to analog")
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

	// Navigate to News tab
	if !clickID(t, driver, TabNewsID) {
		t.Fatal("failed to click News tab")
	}
	if !hasLabel(driver.Snapshot(), "News & Headlines") {
		t.Fatal("expected News screen title")
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

	// Navigate to Settings -> Language tab
	clickID(t, driver, TabSettingsID)
	clickID(t, driver, IDSettingsTabLanguage)
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

	// Navigate to Settings -> Clock sub-tab
	clickID(t, driver, TabSettingsID)
	if !clickID(t, driver, IDSettingsTabFont) {
		t.Fatal("failed to click Font sub-tab")
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
				if strings.Contains(strings.ToLower(lbl.Value), "olive clock") {
					t.Fatalf("found residual 'Olive Clock' text in label %q on tab %v", lbl.Value, state.ActiveTab)
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
