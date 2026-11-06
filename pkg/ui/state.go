package ui

import (
	"time"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
	"worldclock/pkg/alarm"
	"worldclock/pkg/i18n"
	"worldclock/pkg/rss"
	"worldclock/pkg/sleep"
	"worldclock/pkg/storage"
	"worldclock/pkg/timer"
	"worldclock/pkg/tracker"
	"worldclock/pkg/wallpaper"
	"worldclock/pkg/weather"
	"worldclock/pkg/world"
)

// Tab represents an active navigation screen.
type Tab string

const (
	TabClock    Tab = "clock"
	TabWorld    Tab = "world"
	TabAlarms   Tab = "alarms"
	TabTimers   Tab = "timers"
	TabTracker  Tab = "tracker"
	TabSleep    Tab = "sleep"
	TabNews     Tab = "news"
	TabSettings Tab = "settings"
	TabPopout   Tab = "popout"
)

// AppState holds the full runtime application state.
type AppState struct {
	Store     *storage.Store
	World     *world.Manager
	Scheduler *alarm.Scheduler
	TimerMgr  *timer.Manager
	Stopwatch *timer.Stopwatch
	Weather   *weather.Service
	Tracker   *tracker.Tracker
	Sleep     *sleep.Session
	News      *rss.FeedStore
	Wallpaper *wallpaper.Manager

	Invalidator *app.Invalidator

	ActiveTab Tab
	IsPopout  bool

	// Auxiliary UI state
	NewsCategory      string
	SleepDuration     time.Duration
	SleepSound        sleep.AmbientSound
	SettingsTab       string // "font", "wallpaper", "lang", "clock", "about"
	ScrollY           float32
	CityModalScrollY  float32
	ShowAddCityModal  bool
	SelectedContinent string
	CitySearchQuery   string
	CitySearchField   flex.Field
	lastFieldText     string
	ShowNewAlarmModal bool
	NewAlarmTitle     string
	NewAlarmHour      int
	NewAlarmMinute    int
	NewAlarmSchedule  alarm.ScheduleType
	NewAlarmWeekdays  [7]bool

	// Flex layout view (keeps layout buffers across frames)
	View flex.View

	// Cached clock instant for deterministic per-frame rendering
	Now time.Time
	UTC time.Time
}

// NewAppState initializes application state with persistent storage.
func NewAppState(store *storage.Store) *AppState {
	if store == nil {
		store = storage.NewStore("")
	}
	s := store.Get()

	// Initialize managers with persisted data
	worldMgr := world.NewManager(s.WorldCities)
	alarmSched := alarm.NewScheduler(s.Alarms)
	timerMgr := timer.NewManager()
	sw := timer.NewStopwatch()
	wSvc := weather.NewService()
	tr := tracker.NewTracker(nil)
	sl := sleep.NewSession()
	news := rss.NewFeedStore()
	wpMgr := wallpaper.NewManager(s.Wallpaper)
	var weekdays [7]bool
	weekdays[2] = true
	weekdays[3] = true
	weekdays[4] = true
	weekdays[5] = true

	activeTab := TabClock
	if s.PopoutMode {
		activeTab = TabPopout
	}

	return &AppState{
		Store:            store,
		World:            worldMgr,
		Scheduler:        alarmSched,
		TimerMgr:         timerMgr,
		Stopwatch:        sw,
		Weather:          wSvc,
		Tracker:          tr,
		Sleep:            sl,
		News:             news,
		Wallpaper:        wpMgr,
		ActiveTab:        activeTab,
		IsPopout:         s.PopoutMode,
		NewsCategory:     "All",
		SleepDuration:    30 * time.Minute,
		SleepSound:       sleep.SoundRain,
		SettingsTab:      "font",
		ScrollY:          0,
		CityModalScrollY: 0,
		NewAlarmHour:     8,
		NewAlarmMinute:   0,
		NewAlarmTitle:    "New Alarm",
		NewAlarmSchedule: alarm.ScheduleDaily,
		NewAlarmWeekdays: weekdays,
	}
}

// Persist syncs active in-memory lists back into storage.
// Lang returns current selected language.
func (a *AppState) Lang() i18n.Language {
	l := i18n.Language(a.Store.Get().Language)
	if l == "" {
		return i18n.LangZH
	}
	return l
}

// T returns translated string for key in current language.
func (a *AppState) T(key string) string {
	return i18n.T(a.Lang(), key)
}

// Persist syncs active in-memory lists back into storage.
func (a *AppState) Persist() {
	_ = a.Store.Update(func(cfg *storage.Settings) {
		cfg.WorldCities = a.World.Cities()
		cfg.Alarms = a.Scheduler.Alarms()
		cfg.PopoutMode = a.IsPopout
		cfg.Wallpaper = a.Wallpaper.Config()
	})
}

// ScrollOffset returns current vertical scroll offset.
func (a *AppState) ScrollOffset() float32 {
	return a.ScrollY
}

// ScrollBy adjusts scroll offset by delta pixels, clamped between 0 and 1200.
func (a *AppState) ScrollBy(delta float32) {
	a.ScrollY += delta
	if a.ScrollY < 0 {
		a.ScrollY = 0
	}
	if a.ScrollY > 1200 {
		a.ScrollY = 1200
	}
}

// ResetScroll resets scroll offset back to top.
func (a *AppState) ResetScroll() {
	a.ScrollY = 0
}

// ModalScrollOffset returns current vertical scroll offset for city modal.
func (a *AppState) ModalScrollOffset() float32 {
	return a.CityModalScrollY
}

// ModalScrollBy adjusts city modal scroll offset by delta.
func (a *AppState) ModalScrollBy(delta float32) {
	a.CityModalScrollY += delta
	if a.CityModalScrollY < 0 {
		a.CityModalScrollY = 0
	}
	if a.CityModalScrollY > 6000 {
		a.CityModalScrollY = 6000
	}
}

// ResetModalScroll resets city modal scroll back to top.
func (a *AppState) ResetModalScroll() {
	a.CityModalScrollY = 0
}

// SetCitySearchQuery updates the city search query and synchronizes the interactive text field.
func (a *AppState) SetCitySearchQuery(q string) {
	a.CitySearchQuery = q
	a.CitySearchField.SetText(q)
	a.lastFieldText = q
	a.ResetModalScroll()
}

// ClockColor returns user-customized clock font color or theme foreground.
func (a *AppState) ClockColor(theme ui.Theme) ui.Color {
	col := a.Store.Get().ClockFontColor
	switch col {
	case "white":
		return ui.RGB24(0xEDEDED)
	case "gold":
		return ui.RGB24(0xFFBE0B)
	case "olive":
		return ui.RGB24(0x708238)
	case "cyan":
		return ui.RGB24(0x4CC9F0)
	case "purple":
		return ui.RGB24(0xB5838D)
	case "coral":
		return ui.RGB24(0xE76F51)
	default:
		return theme.Foreground
	}
}

// EffectiveClockFontSize returns user-customized font size scaled by FontScale.
func (a *AppState) EffectiveClockFontSize() float32 {
	cfg := a.Store.Get()
	sz := float32(cfg.ClockFontSize)
	if sz <= 0 {
		sz = 64
	}
	scale := cfg.FontScale
	if scale <= 0 {
		scale = 1.0
	}
	return sz * scale
}

// Invalidate requests a repaint across goroutine boundaries.
func (a *AppState) Invalidate() {
	if a.Invalidator != nil {
		a.Invalidator.Invalidate()
	}
}
