package ui

import (
	"sync"
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

// Tab 表示顶部导航栏激活的功能屏幕枚举。
// Tab represents an active navigation screen.
type Tab string

const (
	TabClock    Tab = "clock"    // 主时钟页
	TabWorld    Tab = "world"    // 世界时钟页
	TabAlarms   Tab = "alarms"   // 闹钟管理页
	TabTimers   Tab = "timers"   // 倒计时与秒表页
	TabTracker  Tab = "tracker"  // 生产力专注计时（番茄钟）页
	TabSleep    Tab = "sleep"    // 睡眠与环境音页
	TabNews     Tab = "news"     // 每日资讯页
	TabSettings Tab = "settings" // 设置与偏好页
	TabPopout   Tab = "popout"   // 桌面悬浮小窗模式
)

// AppState 聚合应用程序运行时的所有领域管理器、窗口句柄、跨线程唤醒队列以及界面临时交互状态。
// AppState holds the full runtime application state.
type AppState struct {
	Store     *storage.Store     // 配置持久化存储
	World     *world.Manager     // 世界时钟管理器
	Scheduler *alarm.Scheduler   // 闹钟调度器
	TimerMgr  *timer.Manager     // 多倒计时管理器
	Stopwatch *timer.Stopwatch   // 高精度秒表
	Weather   *weather.Service   // 城市天气服务
	Tracker   *tracker.Tracker   // 番茄专注计时器
	Sleep     *sleep.Session     // 睡眠助眠会话
	News      *rss.FeedStore     // 资讯流存储
	Wallpaper *wallpaper.Manager // 壁纸与轮播引擎

	Invalidator *app.Invalidator // 跨线程触发窗口重绘与 OnWake 的通知器
	Window      *app.Window      // 当前主窗口引用
	Application *app.Application // 当前 Viva 应用实例引用
	ForceQuit   bool             // 是否由系统托盘菜单发起强制退出

	wakeMu                 sync.Mutex // 保护跨线程唤醒队列的互斥锁
	pendingShowWindow      bool       // 是否存在待处理的恢复显示主窗口请求
	pendingWallpaperFiles  []string   // 异步文件选择框选中的待加载壁纸文件列表
	pendingWallpaperFolder string     // 异步文件夹选择框选中的待加载壁纸目录

	ActiveTab Tab  // 当前激活的功能标签页
	IsPopout  bool // 是否处于紧凑型悬浮小窗模式

	// 界面辅助交互状态
	NewsCategory       string             // 资讯页当前选中的分类过滤
	SleepDuration      time.Duration      // 睡眠页选中的倒计时时长
	SleepSound         sleep.AmbientSound // 睡眠页选中的环境音
	SettingsTab        string             // 设置页当前选中的子分类 ("font", "wallpaper", "lang", "clock", "about")
	ScrollY            float32            // 通用垂直滚动偏移量
	CityModalScrollY   float32            // 选城弹窗滚动偏移量
	ShowAddCityModal   bool               // 是否展开“添加世界时钟城市”面板
	DuplicateCityAlert string             // 重复添加城市提醒弹窗的城市名称（非空时展示模态对话框）
	SelectedContinent  string             // 选城面板当前选中的大洲过滤条件
	CitySearchQuery    string             // 选城面板当前输入的搜索关键词
	CitySearchField    flex.Field         // 选城搜索文本输入框组件状态
	lastFieldText      string             // 上一帧搜索框文本快照，用于检测用户输入变化
	ShowNewAlarmModal  bool               // 是否展开“新建闹钟”面板
	NewAlarmTitle      string             // 新建闹钟的标题
	NewAlarmHour       int                // 新建闹钟的小时 (0-23)
	NewAlarmMinute     int                // 新建闹钟的分钟 (0-59)
	NewAlarmSchedule   alarm.ScheduleType // 新建闹钟的重复规则
	NewAlarmWeekdays   [7]bool            // 新建闹钟的自定义星期掩码

	// Flex 声明式布局视图容器（跨帧复用内部布局缓冲区以实现零额外分配）
	View flex.View

	// 当前帧统一的本地时间与 UTC 时间戳快照
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

// UpdateSettings updates settings in Store while guaranteeing that live WorldCities, Alarms,
// PopoutMode, and Wallpaper state are always synchronized before writing to disk.
func (a *AppState) UpdateSettings(fn func(*storage.Settings)) {
	_ = a.Store.Update(func(cfg *storage.Settings) {
		cfg.WorldCities = a.World.Cities()
		cfg.Alarms = a.Scheduler.Alarms()
		cfg.PopoutMode = a.IsPopout
		cfg.Wallpaper = a.Wallpaper.Config()
		if fn != nil {
			fn(cfg)
		}
	})
}

// RequestShowWindow queues a request to show/restore the main window on the UI thread.
func (a *AppState) RequestShowWindow() {
	a.wakeMu.Lock()
	a.pendingShowWindow = true
	a.wakeMu.Unlock()
	a.Invalidate()
}

// QueueWallpaperFiles queues selected wallpaper file paths from an async file dialog for the UI thread.
func (a *AppState) QueueWallpaperFiles(paths []string) {
	if len(paths) == 0 {
		return
	}
	a.wakeMu.Lock()
	a.pendingWallpaperFiles = append([]string(nil), paths...)
	a.wakeMu.Unlock()
	a.Invalidate()
}

// QueueWallpaperFolder queues a selected wallpaper folder path from an async folder dialog for the UI thread.
func (a *AppState) QueueWallpaperFolder(dir string) {
	if dir == "" {
		return
	}
	a.wakeMu.Lock()
	a.pendingWallpaperFolder = dir
	a.wakeMu.Unlock()
	a.Invalidate()
}

// ProcessPendingWake runs on the window thread inside OnWake before painting.
func (a *AppState) ProcessPendingWake() {
	a.wakeMu.Lock()
	showWin := a.pendingShowWindow
	files := a.pendingWallpaperFiles
	folder := a.pendingWallpaperFolder
	a.pendingShowWindow = false
	a.pendingWallpaperFiles = nil
	a.pendingWallpaperFolder = ""
	a.wakeMu.Unlock()

	if showWin && a.Window != nil {
		a.Window.Show()
	}
	if len(files) > 0 {
		if len(files) == 1 {
			_ = a.Wallpaper.LoadImage(files[0])
		} else {
			a.Wallpaper.LoadImages(files)
		}
		a.Persist()
	}
	if folder != "" {
		_, _ = a.Wallpaper.LoadFolder(folder)
		a.Persist()
	}
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
