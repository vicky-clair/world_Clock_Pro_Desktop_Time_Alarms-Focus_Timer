package ui

import (
	"sync"
	"time"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
	"worldclock/pkg/alarm"
	"worldclock/pkg/audio"
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
	News      *rss.FeedStore     // 历史上的今天与精选事件流存储
	Wallpaper *wallpaper.Manager // 壁纸与轮播引擎
	Audio     *audio.Engine      // 原生程序化合成音频播放引擎（闹钟铃声与睡眠白噪音）

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
	IsEdgeBar bool // 是否处于靠近屏幕边缘自动缩小的状态条模式

	// 屏幕边缘收缩状态条几何与防抖状态
	DockedEdge                   string  // 当前停靠边缘方向 ("top", "left", "right", "bottom")
	LastNormalFrame              ui.Rect // 最近一次正常窗口的内容区屏幕坐标与尺寸
	SavedNormalFrame             ui.Rect // 进入边缘状态条前保存的正常窗口矩形
	SuppressEdgeBarUntilMoveAway bool    // 点击“恢复正常”按钮后，在窗口移出边缘前暂时屏蔽自动收缩

	// 开机自动启动同步钩子（测试可注入自定义回调，默认调用 SyncSystemAutostart）
	AutostartHook func(enabled bool) error

	// 界面辅助交互状态
	NewsCategory       string             // 资讯页当前选中的分类过滤
	SleepDuration      time.Duration      // 睡眠页选中的倒计时时长
	SleepSound         sleep.AmbientSound // 睡眠页选中的环境音
	SettingsTab        string             // 设置页当前选中的子分类 ("font", "wallpaper", "lang", "clock", "about")
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
	NewAlarmSound      string             // 新建闹钟选中的铃声名称 ("Classic Chime", "Gentle Chime", "Digital Bell")

	// 倒计时任务新建/编辑面板状态
	ShowTimerModal    bool       // 是否展开“新建/编辑倒计时任务”面板
	EditingTimerID    string     // 正在编辑的任务 ID（为空表示新建自定义任务）
	TimerNameField    flex.Field // 倒计时任务名称输入框组件状态
	TimerModalHours   int        // 倒计时设定小时 (0-23)
	TimerModalMinutes int        // 倒计时设定分钟 (0-59)
	TimerModalSeconds int        // 倒计时设定秒数 (0-59)

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
	timerMgr := timer.NewManagerWithTimers(s.Timers)
	sw := timer.NewStopwatch()
	wSvc := weather.NewService()
	tr := tracker.NewTracker(nil)
	sl := sleep.NewSession()
	news := rss.NewFeedStore()
	wpMgr := wallpaper.NewManager(s.Wallpaper)
	audioEng := audio.NewEngine()
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
		Store:             store,
		World:             worldMgr,
		Scheduler:         alarmSched,
		TimerMgr:          timerMgr,
		Stopwatch:         sw,
		Weather:           wSvc,
		Tracker:           tr,
		Sleep:             sl,
		News:              news,
		Wallpaper:         wpMgr,
		Audio:             audioEng,
		ActiveTab:         activeTab,
		IsPopout:          s.PopoutMode,
		NewsCategory:      "All",
		SleepDuration:     30 * time.Minute,
		SleepSound:        sleep.SoundRain,
		SettingsTab:       "font",
		NewAlarmHour:      8,
		NewAlarmMinute:    0,
		NewAlarmTitle:     "New Alarm",
		NewAlarmSchedule:  alarm.ScheduleDaily,
		NewAlarmWeekdays:  weekdays,
		NewAlarmSound:     "Classic Chime",
		TimerModalHours:   0,
		TimerModalMinutes: 10,
		TimerModalSeconds: 0,
	}
}

// Lang 返回当前选中的界面显示语言（默认回退为简体中文）。
// Lang returns current selected language.
func (a *AppState) Lang() i18n.Language {
	l := i18n.Language(a.Store.Get().Language)
	if l == "" {
		return i18n.LangZH
	}
	return l
}

// T 根据当前激活语言翻译指定词条 key。
// T returns translated string for key in current language.
func (a *AppState) T(key string) string {
	return i18n.T(a.Lang(), key)
}

// Persist 将内存中的世界城市、闹钟、倒计时任务、悬浮窗状态与壁纸配置同步持久化至磁盘。
// Persist syncs active in-memory lists back into storage.
func (a *AppState) Persist() {
	_ = a.Store.Update(func(cfg *storage.Settings) {
		cfg.WorldCities = a.World.Cities()
		cfg.Alarms = a.Scheduler.Alarms()
		cfg.Timers = a.TimerMgr.ExportTimers()
		cfg.PopoutMode = a.IsPopout
		cfg.Wallpaper = a.Wallpaper.Config()
	})
}

// UpdateSettings 在同步当前内存列表状态的前提下执行自定义设置修改回调 fn 并立即持久化。
// UpdateSettings updates settings in Store while guaranteeing that live WorldCities, Alarms,
// Timers, PopoutMode, and Wallpaper state are always synchronized before writing to disk.
func (a *AppState) UpdateSettings(fn func(*storage.Settings)) {
	_ = a.Store.Update(func(cfg *storage.Settings) {
		cfg.WorldCities = a.World.Cities()
		cfg.Alarms = a.Scheduler.Alarms()
		cfg.Timers = a.TimerMgr.ExportTimers()
		cfg.PopoutMode = a.IsPopout
		cfg.Wallpaper = a.Wallpaper.Config()
		if fn != nil {
			fn(cfg)
		}
	})
}

// RequestShowWindow 从任意后台协程安全请求在 UI 主线程恢复并显示主窗口。
// RequestShowWindow queues a request to show/restore the main window on the UI thread.
func (a *AppState) RequestShowWindow() {
	a.wakeMu.Lock()
	a.pendingShowWindow = true
	a.wakeMu.Unlock()
	a.Invalidate()
}

// QueueWallpaperFiles 将异步文件选择框选中的图片路径列表入队，交由 UI 主线程加载。
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

// QueueWallpaperFolder 将异步文件夹选择框选中的目录路径入队，交由 UI 主线程加载。
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

// ProcessPendingWake 在 UI 主线程的 OnWake 回调中安全消费跨线程唤醒事件（恢复窗口、加载新选中的壁纸/文件夹）。
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

// ResetScroll 在切换标签页或子标签页时将对应 ScrollArea 的滚动位置重置回顶部。
// ResetScroll resets scroll offset back to top.
func (a *AppState) ResetScroll() {
	if ss := a.View.ScrollState(IDScrollSettings); ss != nil {
		ss.Offset = 0
	}
}

// ResetModalScroll 将世界时钟选城弹窗的滚动位置重置回顶部。
// ResetModalScroll resets city modal scroll back to top.
func (a *AppState) ResetModalScroll() {
	if ss := a.View.ScrollState(IDScrollCityModal); ss != nil {
		ss.Offset = 0
	}
}

// SetCitySearchQuery 更新选城面板搜索关键词并同步刷新输入框组件文本与滚动条位置。
// SetCitySearchQuery updates the city search query and synchronizes the interactive text field.
func (a *AppState) SetCitySearchQuery(q string) {
	a.CitySearchQuery = q
	a.CitySearchField.SetText(q)
	a.lastFieldText = q
	a.ResetModalScroll()
}

// ClockColor 返回用户自定义的时钟数字高亮颜色；若未指定则回退为主题前景色。
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

// EffectiveClockFontSize 计算经全局界面缩放系数 (FontScale) 加权后的主时钟有效字号。
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

// Invalidate 跨协程安全地通知 Viva 渲染管线触发下一帧重绘。
// Invalidate requests a repaint across goroutine boundaries.
func (a *AppState) Invalidate() {
	if a.Invalidator != nil {
		a.Invalidator.Invalidate()
	}
}

// SetAutoStart 更新“开机自动启动”设置并同步写入操作系统当前用户的登录启动项。
func (a *AppState) SetAutoStart(enabled bool) {
	a.UpdateSettings(func(s *storage.Settings) {
		s.AutoStart = enabled
	})
	if a.AutostartHook != nil {
		_ = a.AutostartHook(enabled)
	} else {
		_ = SyncSystemAutostart(enabled)
	}
}

// EnterEdgeBar 进入屏幕边缘精简状态条模式，并在桌面环境下将窗口自动缩小为顶部/边缘横条。
func (a *AppState) EnterEdgeBar(edge string) {
	if edge == "" {
		edge = "top"
	}
	if a.LastNormalFrame.Width > 0 && a.LastNormalFrame.Height > 0 {
		a.SavedNormalFrame = a.LastNormalFrame
	} else if a.SavedNormalFrame.Width <= 0 {
		a.SavedNormalFrame = ui.Rect{X: 120, Y: 90, Width: 840, Height: 580}
	}
	a.IsEdgeBar = true
	a.DockedEdge = edge
	platformShrinkWindowToEdgeBar(edge, a.SavedNormalFrame)
	a.Invalidate()
}

// ExitEdgeBar 退出屏幕边缘状态条模式，恢复正常窗口界面与尺寸，并设置防抖锁避免立即再次触发收缩。
func (a *AppState) ExitEdgeBar() {
	a.IsEdgeBar = false
	a.SuppressEdgeBarUntilMoveAway = true
	platformRestoreWindowFromEdgeBar(a.SavedNormalFrame)
	a.Invalidate()
}

// HandleFrameChange 处理窗口位置或尺寸变化事件（由 Viva Config.OnFrameChange 回调驱动）。
// 当检测到窗口靠近屏幕边缘且已启用边缘收缩功能时，自动将窗口收缩为状态条模式。
func (a *AppState) HandleFrameChange(frame ui.Rect) {
	if frame.Width <= 0 || frame.Height <= 0 {
		return
	}
	if a.Window != nil && (a.Window.Maximized() || a.Window.Minimized() || a.Window.Fullscreen()) {
		return
	}
	if a.IsEdgeBar {
		return
	}

	workArea := platformGetMonitorWorkArea(frame)
	edge := DetectScreenEdge(frame, workArea)
	if edge == "" {
		a.LastNormalFrame = frame
		if IsClearlyInsideScreen(frame, workArea) {
			a.SuppressEdgeBarUntilMoveAway = false
		}
		return
	}

	if !a.Store.Get().EdgeDockEnabled {
		a.LastNormalFrame = frame
		return
	}
	if a.SuppressEdgeBarUntilMoveAway {
		return
	}

	if a.LastNormalFrame.Width <= 0 || a.LastNormalFrame.Height <= 0 {
		a.LastNormalFrame = ui.Rect{X: frame.X, Y: frame.Y, Width: 840, Height: 580}
	}
	a.EnterEdgeBar(edge)
}
