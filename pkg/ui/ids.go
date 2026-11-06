package ui

import "github.com/vivaui/viva/ui"

// Tree 保存全局唯一的 Viva 控件身份标识树（ui.Tree），用于在跨帧绘制时稳定追踪焦点、悬停与点击命中。
// Tree holds the static identity tree for the World Clock application.
var (
	Tree = ui.Tree{}

	// 功能分区（Features / Areas）
	featNav     = Tree.Feature("nav")      // 顶部导航栏区域
	featHeader  = Tree.Feature("header")   // 顶部闹钟提醒横幅区域
	featClock   = Tree.Feature("clock")    // 主时钟区域
	featWorld   = Tree.Feature("world")    // 世界时钟与选城面板区域
	featAlarms  = Tree.Feature("alarms")   // 闹钟列表与新建面板区域
	featTimers  = Tree.Feature("timers")   // 倒计时与秒表区域
	featPopout  = Tree.Feature("popout")   // 桌面悬浮小窗区域
	featTracker = Tree.Feature("tracker")  // 专注计时器（番茄钟）区域
	featSleep   = Tree.Feature("sleep")    // 睡眠与环境音区域
	featNews    = Tree.Feature("news")     // 资讯阅读区域
	featSetting = Tree.Feature("settings") // 设置与偏好区域
	featDialog  = Tree.Feature("dialog")   // 全局模态对话框区域

	// Navigation Tabs
	TabClockID    = Tree.ID(featNav, "tab-clock")
	TabWorldID    = Tree.ID(featNav, "tab-world")
	TabAlarmsID   = Tree.ID(featNav, "tab-alarms")
	TabTimersID   = Tree.ID(featNav, "tab-timers")
	TabTrackerID  = Tree.ID(featNav, "tab-tracker")
	TabSleepID    = Tree.ID(featNav, "tab-sleep")
	TabNewsID     = Tree.ID(featNav, "tab-news")
	TabSettingsID = Tree.ID(featNav, "tab-settings")
	TabPopoutID   = Tree.ID(featNav, "tab-popout")

	// Header controls
	IDBtnSnooze  = Tree.ID(featHeader, "btn-snooze")
	IDBtnDismiss = Tree.ID(featHeader, "btn-dismiss")

	// Main Clock Controls
	IDBtnToggle24H     = Tree.ID(featClock, "btn-toggle-24h")
	IDBtnToggleSeconds = Tree.ID(featClock, "btn-toggle-seconds")
	IDBtnToggleStyle   = Tree.ID(featClock, "btn-toggle-style")
	IDBtnQuickPopout   = Tree.ID(featClock, "btn-quick-popout")

	// World Clock Controls
	IDBtnAddCityModal         = Tree.ID(featWorld, "btn-add-city-modal")
	IDBtnCloseAddCity         = Tree.ID(featWorld, "btn-close-add-city")
	IDInputCitySearch         = Tree.ID(featWorld, "input-city-search")
	IDBtnClearCitySearch      = Tree.ID(featWorld, "btn-clear-city-search")
	IDScrollCityModal         = Tree.ID(featWorld, "scroll-city-modal")
	IDScrollWorldList         = Tree.ID(featWorld, "scroll-world-list")
	IDDialogDupCityScope      = Tree.ID(featDialog, "dup-city-scope")
	IDDialogDupCityBackdrop   = Tree.ID(featDialog, "dup-city-backdrop")
	IDDialogDupCityClose      = Tree.ID(featDialog, "dup-city-close")
	IDBtnDismissDuplicateCity = Tree.ID(featDialog, "btn-dismiss-dup-city")

	IDQuickFilterCN = Tree.ID(featWorld, "qf-cn")
	IDQuickFilterUS = Tree.ID(featWorld, "qf-us")
	IDQuickFilterJP = Tree.ID(featWorld, "qf-jp")
	IDQuickFilterUK = Tree.ID(featWorld, "qf-uk")
	IDQuickFilterFR = Tree.ID(featWorld, "qf-fr")
	IDQuickFilterDE = Tree.ID(featWorld, "qf-de")
	// Continent filter tabs
	IDContAll          = Tree.ID(featWorld, "cont-all")
	IDContAsia         = Tree.ID(featWorld, "cont-asia")
	IDContEurope       = Tree.ID(featWorld, "cont-europe")
	IDContAfrica       = Tree.ID(featWorld, "cont-africa")
	IDContNorthAmerica = Tree.ID(featWorld, "cont-na")
	IDContSouthAmerica = Tree.ID(featWorld, "cont-sa")
	IDContOceania      = Tree.ID(featWorld, "cont-oc")

	// Alarm Controls
	IDBtnNewAlarmModal     = Tree.ID(featAlarms, "btn-new-alarm-modal")
	IDBtnCloseNewAlarm     = Tree.ID(featAlarms, "btn-close-new-alarm")
	IDBtnSaveNewAlarm      = Tree.ID(featAlarms, "btn-save-new-alarm")
	IDBtnPresetMorning     = Tree.ID(featAlarms, "preset-morning")
	IDBtnPresetNoon        = Tree.ID(featAlarms, "preset-noon")
	IDBtnPresetEvening     = Tree.ID(featAlarms, "preset-evening")
	IDBtnAlarmHourDown     = Tree.ID(featAlarms, "alarm-hr-down")
	IDBtnAlarmHourUp       = Tree.ID(featAlarms, "alarm-hr-up")
	IDBtnAlarmMinDown      = Tree.ID(featAlarms, "alarm-min-down")
	IDBtnAlarmMinUp        = Tree.ID(featAlarms, "alarm-min-up")
	IDBtnRecurOnce         = Tree.ID(featAlarms, "recur-once")
	IDBtnRecurDaily        = Tree.ID(featAlarms, "recur-daily")
	IDBtnRecurWeekdays     = Tree.ID(featAlarms, "recur-weekdays")
	IDBtnAlarmSoundGentle  = Tree.ID(featAlarms, "alarm-snd-gentle")
	IDBtnAlarmSoundDigital = Tree.ID(featAlarms, "alarm-snd-digital")
	IDBtnAlarmSoundClassic = Tree.ID(featAlarms, "alarm-snd-classic")
	IDBtnAlarmPreviewSound = Tree.ID(featAlarms, "alarm-snd-preview")

	// Timers Controls
	IDBtnAddTimerCustom = Tree.ID(featTimers, "btn-add-timer-custom")
	IDBtnSwStart        = Tree.ID(featTimers, "btn-sw-start")
	IDBtnSwPause        = Tree.ID(featTimers, "btn-sw-pause")
	IDBtnSwReset        = Tree.ID(featTimers, "btn-sw-reset")
	IDBtnSwLap          = Tree.ID(featTimers, "btn-sw-lap")
	// Tracker Controls
	IDBtnTrackerStart    = Tree.ID(featTracker, "btn-tracker-start")
	IDBtnTrackerPause    = Tree.ID(featTracker, "btn-tracker-pause")
	IDBtnTrackerResume   = Tree.ID(featTracker, "btn-tracker-resume")
	IDBtnTrackerReset    = Tree.ID(featTracker, "btn-tracker-reset")
	IDBtnTrackerSkip     = Tree.ID(featTracker, "btn-tracker-skip")
	IDBtnTrackerPostpone = Tree.ID(featTracker, "btn-tracker-postpone")
	IDBtnTrackerDND      = Tree.ID(featTracker, "btn-tracker-dnd")

	// Sleep Controls
	IDBtnSleepStart   = Tree.ID(featSleep, "btn-sleep-start")
	IDBtnSleepStop    = Tree.ID(featSleep, "btn-sleep-stop")
	IDBtnSleep15m     = Tree.ID(featSleep, "btn-sleep-15m")
	IDBtnSleep30m     = Tree.ID(featSleep, "btn-sleep-30m")
	IDBtnSleep60m     = Tree.ID(featSleep, "btn-sleep-60m")
	IDBtnSleepRain    = Tree.ID(featSleep, "btn-sleep-rain")
	IDBtnSleepStream  = Tree.ID(featSleep, "btn-sleep-stream")
	IDBtnSleepWaves   = Tree.ID(featSleep, "btn-sleep-waves")
	IDBtnSleepWhite   = Tree.ID(featSleep, "btn-sleep-white")
	IDBtnSleepPreview = Tree.ID(featSleep, "btn-sleep-preview")

	// News / History Today Controls
	IDNewsCatAll        = Tree.ID(featNews, "cat-all")
	IDNewsCatTech       = Tree.ID(featNews, "cat-tech")
	IDNewsCatWorld      = Tree.ID(featNews, "cat-world")
	IDNewsCatScience    = Tree.ID(featNews, "cat-science")
	IDBtnRefreshHistory = Tree.ID(featNews, "btn-refresh-history")

	// Popout Controls
	IDBtnExitPopout = Tree.ID(featPopout, "btn-exit-popout")
	IDBtnPinPopout  = Tree.ID(featPopout, "btn-pin-popout")
	// Settings Controls
	IDThemeOlive = Tree.ID(featSetting, "theme-olive")
	IDThemeDark  = Tree.ID(featSetting, "theme-dark")
	IDThemeLight = Tree.ID(featSetting, "theme-light")
	IDThemeNord  = Tree.ID(featSetting, "theme-nord")

	IDOptShowSeconds        = Tree.ID(featSetting, "opt-show-seconds")
	IDOptFormat24H          = Tree.ID(featSetting, "opt-format-24h")
	IDOptFullDate           = Tree.ID(featSetting, "opt-full-date")
	IDOptStyleDigital       = Tree.ID(featSetting, "opt-style-digital")
	IDOptStyleAnalog        = Tree.ID(featSetting, "opt-style-analog")
	IDProfileDefault        = Tree.ID(featSetting, "profile-default")
	IDProfileWork           = Tree.ID(featSetting, "profile-work")
	IDOptTempFahrenheit     = Tree.ID(featSetting, "opt-temp-unit")
	IDCloseBehaviorMinimize = Tree.ID(featSetting, "close-behavior-minimize")
	IDCloseBehaviorExit     = Tree.ID(featSetting, "close-behavior-exit")

	// Settings Sub-Tabs
	IDSettingsTabFont      = Tree.ID(featSetting, "subtab-font")
	IDSettingsTabWallpaper = Tree.ID(featSetting, "subtab-wallpaper")
	IDSettingsTabLanguage  = Tree.ID(featSetting, "subtab-language")
	IDSettingsTabClock     = Tree.ID(featSetting, "subtab-clock")
	IDSettingsTabAbout     = Tree.ID(featSetting, "subtab-about")
	// Language Controls
	IDLangZH = Tree.ID(featSetting, "lang-zh")
	IDLangEN = Tree.ID(featSetting, "lang-en")
	IDLangJA = Tree.ID(featSetting, "lang-ja")

	// Wallpaper Controls
	IDWallpaperToggle          = Tree.ID(featSetting, "wp-toggle")
	IDBtnBrowseWallpaper       = Tree.ID(featSetting, "wp-browse")
	IDBtnBrowseFolder          = Tree.ID(featSetting, "wp-browse-folder")
	IDBtnSampleWallpaper       = Tree.ID(featSetting, "wp-sample")
	IDBtnSampleSlideshow       = Tree.ID(featSetting, "wp-sample-slideshow")
	IDBtnClearWallpaper        = Tree.ID(featSetting, "wp-clear")
	IDWallpaperSlideshowToggle = Tree.ID(featSetting, "wp-slideshow-toggle")
	IDWallpaperSlidePrev       = Tree.ID(featSetting, "wp-slide-prev")
	IDWallpaperSlideNext       = Tree.ID(featSetting, "wp-slide-next")
	IDWallpaperInt1m           = Tree.ID(featSetting, "wp-int-1m")
	IDWallpaperInt5m           = Tree.ID(featSetting, "wp-int-5m")
	IDWallpaperInt10m          = Tree.ID(featSetting, "wp-int-10m")
	IDWallpaperInt15m          = Tree.ID(featSetting, "wp-int-15m")
	IDWallpaperOrderShuffle    = Tree.ID(featSetting, "wp-order-shuffle")
	IDWallpaperOrderSeq        = Tree.ID(featSetting, "wp-order-seq")
	IDPresetGrid               = Tree.ID(featSetting, "wp-preset-grid")
	IDPresetAurora             = Tree.ID(featSetting, "wp-preset-aurora")
	IDPresetForest             = Tree.ID(featSetting, "wp-preset-forest")
	IDPresetSunset             = Tree.ID(featSetting, "wp-preset-sunset")
	IDWallpaperFitFill         = Tree.ID(featSetting, "wp-fit-fill")
	IDWallpaperFitFit          = Tree.ID(featSetting, "wp-fit-fit")
	IDWallpaperFitStretch      = Tree.ID(featSetting, "wp-fit-stretch")
	IDWallpaperFitTile         = Tree.ID(featSetting, "wp-fit-tile")
	IDWallpaperFitCenter       = Tree.ID(featSetting, "wp-fit-center")
	IDWallpaperOpacityDown     = Tree.ID(featSetting, "wp-opacity-down")
	IDWallpaperOpacityUp       = Tree.ID(featSetting, "wp-opacity-up")
	IDCardGlassDown            = Tree.ID(featSetting, "card-glass-down")
	IDCardGlassUp              = Tree.ID(featSetting, "card-glass-up")

	// ScrollArea IDs
	IDScrollSettings = Tree.ID(featSetting, "scroll-settings")
	IDScrollAlarms   = Tree.ID(featAlarms, "scroll-alarms")
	IDScrollTimers   = Tree.ID(featTimers, "scroll-timers")
	IDScrollNews     = Tree.ID(featNews, "scroll-news")
	IDScrollTracker  = Tree.ID(featTracker, "scroll-tracker")
	// Font Size & Color Controls
	IDFontSize48         = Tree.ID(featSetting, "fs-48")
	IDFontSize64         = Tree.ID(featSetting, "fs-64")
	IDFontSize76         = Tree.ID(featSetting, "fs-76")
	IDFontSize88         = Tree.ID(featSetting, "fs-88")
	IDFontSize96         = Tree.ID(featSetting, "fs-96")
	IDFontSizeCustomDown = Tree.ID(featSetting, "fs-custom-down")
	IDFontSizeCustomUp   = Tree.ID(featSetting, "fs-custom-up")

	IDFontColorWhite  = Tree.ID(featSetting, "fc-white")
	IDFontColorGold   = Tree.ID(featSetting, "fc-gold")
	IDFontColorOlive  = Tree.ID(featSetting, "fc-olive")
	IDFontColorCyan   = Tree.ID(featSetting, "fc-cyan")
	IDFontColorPurple = Tree.ID(featSetting, "fc-purple")
	IDFontColorCoral  = Tree.ID(featSetting, "fc-coral")

	IDUIColorDefault = Tree.ID(featSetting, "uic-default")
	IDUIColorWhite   = Tree.ID(featSetting, "uic-white")
	IDUIColorBlack   = Tree.ID(featSetting, "uic-black")
	IDUIColorGold    = Tree.ID(featSetting, "uic-gold")
	IDUIColorCyan    = Tree.ID(featSetting, "uic-cyan")
	IDUIColorPurple  = Tree.ID(featSetting, "uic-purple")
	IDUIColorOlive   = Tree.ID(featSetting, "uic-olive")

	IDFontScale90  = Tree.ID(featSetting, "fscale-90")
	IDFontScale100 = Tree.ID(featSetting, "fscale-100")
	IDFontScale110 = Tree.ID(featSetting, "fscale-110")
	IDFontScale120 = Tree.ID(featSetting, "fscale-120")
)

// Dynamic IDs generation helpers with stable names inside feature scopes
func CityToggleFavID(cityID string) ui.ID {
	return Tree.ID(featWorld, "fav-"+cityID)
}

func CityRemoveID(cityID string) ui.ID {
	return Tree.ID(featWorld, "rm-"+cityID)
}

func CityMoveUpID(cityID string) ui.ID {
	return Tree.ID(featWorld, "up-"+cityID)
}

func CityMoveDownID(cityID string) ui.ID {
	return Tree.ID(featWorld, "down-"+cityID)
}

func CityAddCatalogID(cityID string) ui.ID {
	return Tree.ID(featWorld, "add-cat-"+cityID)
}

func AlarmToggleID(alarmID string) ui.ID {
	return Tree.ID(featAlarms, "tog-"+alarmID)
}

func AlarmDeleteID(alarmID string) ui.ID {
	return Tree.ID(featAlarms, "del-"+alarmID)
}

func AlarmPreviewID(alarmID string) ui.ID {
	return Tree.ID(featAlarms, "prev-"+alarmID)
}

func TimerStartID(timerID string) ui.ID {
	return Tree.ID(featTimers, "start-"+timerID)
}

func TimerPauseID(timerID string) ui.ID {
	return Tree.ID(featTimers, "pause-"+timerID)
}

func TimerResetID(timerID string) ui.ID {
	return Tree.ID(featTimers, "reset-"+timerID)
}

func TimerDeleteID(timerID string) ui.ID {
	return Tree.ID(featTimers, "del-"+timerID)
}
