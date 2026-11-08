// Package storage 负责用户配置的并发安全读取、深拷贝快照隔离、JSON 持久化以及 .bak 自动容灾备份与损坏恢复。
package storage

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"worldclock/pkg/alarm"
	"worldclock/pkg/license"
	"worldclock/pkg/timer"
	"worldclock/pkg/wallpaper"
	"worldclock/pkg/world"
)

// ThemeType 定义应用视觉主题名称。
// ThemeType names the visual theme.
type ThemeType string

const (
	ThemeDark  ThemeType = "dark"  // 深邃暗黑主题
	ThemeLight ThemeType = "light" // 明亮清爽主题
	ThemeOlive ThemeType = "olive" // 经典橄榄绿主题
	ThemeNord  ThemeType = "nord"  // 极地冰蓝 (Nord) 主题
)

// CloseBehaviorMinimize 与 CloseBehaviorExit 定义用户点击窗口右上角关闭按钮（✕）时的行为。
// CloseBehaviorType defines what happens when the user clicks the window close (X) button.
const (
	CloseBehaviorMinimize = "minimize" // 隐藏窗口并最小化到后台系统托盘
	CloseBehaviorExit     = "exit"     // 直接退出整个应用程序

	PopoutPosTop    = "top"    // 悬浮时钟靠顶显示
	PopoutPosCenter = "center" // 悬浮时钟居中显示
	PopoutPosBottom = "bottom" // 悬浮时钟靠底显示
)

// Settings 聚合所有需要持久化到磁盘的用户偏好设置、世界时钟城市列表、闹钟与壁纸轮播配置。
// Settings represents all persistent user preferences and data.
type Settings struct {
	TimeFormat24H   bool             `json:"timeFormat24H"`   // 是否使用 24 小时制
	ShowSeconds     bool             `json:"showSeconds"`     // 是否显示秒数
	ShowFullDate    bool             `json:"showFullDate"`    // 是否显示完整日期格式
	ShowHolidays    bool             `json:"showHolidays"`    // 是否在主时钟、悬浮时钟与世界时钟中显示今日公历、农历节气与网络节日
	ClockStyle      string           `json:"clockStyle"`      // 时钟样式："digital"（数字）或 "analog"（模拟表盘）
	Theme           ThemeType        `json:"theme"`           // 当前视觉主题
	WorldCities     []world.UserCity `json:"worldCities"`     // 用户已添加的世界时钟城市列表
	Alarms          []alarm.Alarm    `json:"alarms"`          // 用户已配置的闹钟列表
	Timers          []timer.Timer    `json:"timers"`          // 用户已配置的倒计时任务列表（支持自定义增删改持久化）
	Topmost         bool             `json:"topmost"`         // 悬浮小窗是否置顶
	PopoutMode      bool             `json:"popoutMode"`      // 是否处于紧凑型桌面悬浮小窗模式
	PopoutPosition  string           `json:"popoutPosition"`  // 悬浮时钟显示位置："top"（顶部）、"center"（居中）、"bottom"（底部）
	TempFahrenheit  bool             `json:"tempFahrenheit"`  // 是否使用华氏度 (°F) 显示气温
	Language        string           `json:"language"`        // 界面语言："zh"（中文）、"en"（英文）、"ja"（日文）
	FontFamily      string           `json:"fontFamily"`      // 应用全局系统字体族名称（空字符串 "" 表示使用系统默认字体并自动回退）
	Wallpaper       wallpaper.Config `json:"wallpaper"`       // 壁纸与轮播配置
	ClockFontSize   int              `json:"clockFontSize"`   // 主时钟字号像素（如 48, 64, 76, 88）
	ClockFontColor  string           `json:"clockFontColor"`  // 主时钟文字配色（如 "white", "gold", "olive", "cyan", "purple", "coral"）
	UIFontColor     string           `json:"uiFontColor"`     // 全局界面文字颜色（如 "default", "white", "black" 等）
	FontScale       float32          `json:"fontScale"`       // 字体缩放比例（如 0.9, 1.0, 1.1, 1.2）
	CloseBehavior   string           `json:"closeBehavior"`   // 关闭按钮行为："minimize"（最小化到托盘）或 "exit"（直接退出）
	AutoStart       bool             `json:"autoStart"`       // 是否在操作系统登录时开机自动启动
	EdgeDockEnabled bool             `json:"edgeDockEnabled"` // 窗口靠近屏幕边缘后是否自动缩小成一个状态条
	License         license.Config   `json:"license"`         // 7天免费试用与 Pro 会员订阅配置（支持月付3元/年付28元及免付测试版）
}

// DefaultSettings 返回包含默认世界时钟城市与推荐偏好的初始配置。
// DefaultSettings returns fresh default settings.
func DefaultSettings() Settings {
	return Settings{
		TimeFormat24H:   true,
		ShowSeconds:     true,
		ShowFullDate:    true,
		ShowHolidays:    true,
		ClockStyle:      "digital",
		Theme:           ThemeOlive,
		WorldCities:     world.DefaultCities(),
		Timers:          timer.DefaultTimers(),
		Topmost:         false,
		PopoutMode:      false,
		PopoutPosition:  PopoutPosCenter,
		TempFahrenheit:  false,
		Language:        "zh", // 默认使用简体中文界面
		Wallpaper:       wallpaper.DefaultConfig(),
		ClockFontSize:   64,
		ClockFontColor:  "white",
		UIFontColor:     "default",
		FontScale:       1.0,
		CloseBehavior:   CloseBehaviorMinimize,
		AutoStart:       false,
		EdgeDockEnabled: true,
		License: license.Config{
			Plan: license.PlanTrial,
		},
	}
}

// Store 负责并发安全地加载、更新和持久化用户配置。
// Store handles thread-safe loading and persisting of settings.
type Store struct {
	mu        sync.RWMutex
	filePath  string
	settings  Settings
	lastSaved []byte // 最近一次成功落盘或加载的 JSON 字节快照，用于去重免写与内存级 .bak 备份（降低硬盘读写磨损）
}

// NewStore 创建配置存储实例；若 filePath 为空，则默认使用用户配置目录下的 XiaoruIntegratedClockDesktop/settings.json。
// NewStore creates a new storage store. If path is empty, uses the user config directory.
func NewStore(filePath string) *Store {
	if filePath == "" {
		cfgDir, err := os.UserConfigDir()
		if err != nil || cfgDir == "" {
			cfgDir = "."
		}
		filePath = filepath.Join(cfgDir, "XiaoruIntegratedClockDesktop", "settings.json")
	}

	s := &Store{
		filePath: filePath,
		settings: DefaultSettings(),
	}
	_ = s.Load()
	return s
}

// FilePath 返回当前配置文件在磁盘上的完整路径。
// FilePath returns the path to the settings JSON file.
func (s *Store) FilePath() string {
	return s.filePath
}

// GetPrefs 返回当前标量配置偏好的零内存分配快照（不拷贝切片字段，专供 UI 每帧高频渲染与语言查询使用）。
func (s *Store) GetPrefs() Settings {
	s.mu.RLock()
	cp := s.settings
	s.mu.RUnlock()
	cp.WorldCities = nil
	cp.Alarms = nil
	cp.Timers = nil
	cp.Wallpaper.ImagePaths = nil
	return cp
}

// Get 返回当前配置的深拷贝快照（对切片执行独立拷贝，防止外部修改污染内部状态）。
// Get returns a deep copy of current settings.
func (s *Store) Get() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := s.settings
	if s.settings.WorldCities != nil {
		cp.WorldCities = make([]world.UserCity, len(s.settings.WorldCities))
		copy(cp.WorldCities, s.settings.WorldCities)
	}
	if s.settings.Alarms != nil {
		cp.Alarms = make([]alarm.Alarm, len(s.settings.Alarms))
		copy(cp.Alarms, s.settings.Alarms)
	}
	if s.settings.Timers != nil {
		cp.Timers = make([]timer.Timer, len(s.settings.Timers))
		copy(cp.Timers, s.settings.Timers)
	}
	if len(s.settings.Wallpaper.ImagePaths) > 0 {
		cp.Wallpaper.ImagePaths = make([]string, len(s.settings.Wallpaper.ImagePaths))
		copy(cp.Wallpaper.ImagePaths, s.settings.Wallpaper.ImagePaths)
	}
	return cp
}

// Update 在写锁保护下执行配置修改回调 fn，并立即持久化到磁盘。
// Update updates the current settings and persists them.
func (s *Store) Update(fn func(*Settings)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.settings)
	return s.saveLocked()
}

// unmarshalSettingsWithDefaults 解析 JSON 字节流并补齐缺失字段的默认值；
// 同时严格区分 JSON 中的 null/未定义（回退为默认 4 个城市）与显式空数组 "[]"（用户主动清空了所有城市）。
func unmarshalSettingsWithDefaults(data []byte) (Settings, error) {
	loaded := DefaultSettings()
	loaded.WorldCities = nil // 用于区分 JSON 中缺省/null 与显式空切片 "[]"
	loaded.Timers = nil      // 用于区分 JSON 中缺省/null 与显式空切片 "[]"（用户删空所有倒计时任务）
	if err := json.Unmarshal(data, &loaded); err != nil {
		return Settings{}, err
	}
	if loaded.WorldCities == nil {
		loaded.WorldCities = world.DefaultCities()
	}
	if loaded.Timers == nil {
		loaded.Timers = timer.DefaultTimers()
	}
	if loaded.CloseBehavior != CloseBehaviorMinimize && loaded.CloseBehavior != CloseBehaviorExit {
		loaded.CloseBehavior = CloseBehaviorMinimize
	}
	if loaded.PopoutPosition != PopoutPosTop && loaded.PopoutPosition != PopoutPosBottom {
		loaded.PopoutPosition = PopoutPosCenter
	}
	if loaded.ClockFontSize <= 0 {
		loaded.ClockFontSize = 64
	}
	if loaded.FontScale <= 0 {
		loaded.FontScale = 1.0
	}
	if loaded.License.Plan == "" {
		loaded.License.Plan = license.PlanTrial
	}
	return loaded, nil
}

// Load 从磁盘加载配置文件；若主文件缺失或因异常断电损坏，自动尝试从 .bak 备份文件自愈恢复。
// Load loads settings from file if it exists (with automatic backup recovery).
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		// 主文件读取失败时尝试从 .bak 备份文件恢复
		if bakData, bakErr := os.ReadFile(s.filePath + ".bak"); bakErr == nil && len(bakData) > 0 {
			if loaded, uErr := unmarshalSettingsWithDefaults(bakData); uErr == nil {
				s.settings = loaded
				s.lastSaved = nil
				_ = s.saveLocked()
				return nil
			}
		}
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	loaded, err := unmarshalSettingsWithDefaults(data)
	if err != nil {
		// 主文件 JSON 损坏时自动从 .bak 备份恢复并写回主文件
		if bakData, bakErr := os.ReadFile(s.filePath + ".bak"); bakErr == nil && len(bakData) > 0 {
			if bakLoaded, uErr := unmarshalSettingsWithDefaults(bakData); uErr == nil {
				s.settings = bakLoaded
				s.lastSaved = nil
				_ = s.saveLocked()
				return nil
			}
		}
		return err
	}

	s.settings = loaded
	s.lastSaved = data
	return nil
}

// Save 将当前内存中的配置持久化写入磁盘。
// Save persists current settings to disk.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

// saveLocked 在覆写主配置文件前，先将上一份有效配置备份至 .bak 文件，确保任何时刻至少有一份完整有效的数据；
// 同时对比内存快照 s.lastSaved，若配置内容未发生任何实质变化则直接跳过写盘，保护用户 SSD/机械硬盘寿命。
func (s *Store) saveLocked() error {
	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return err
	}

	// 若序列化结果与上一次已落盘数据完全一致，则跳过磁盘写入（零冗余磁盘 I/O）
	if len(s.lastSaved) > 0 && bytes.Equal(data, s.lastSaved) {
		return nil
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	bakPath := s.filePath + ".bak"
	if len(s.lastSaved) > 0 {
		// 直接将内存中的上一份有效 JSON 写入 .bak 备份，确保断电或异常损坏时可 100% 自愈恢复
		_ = os.WriteFile(bakPath, s.lastSaved, 0644)
	} else if prev, readErr := os.ReadFile(s.filePath); readErr == nil && len(prev) > 0 {
		_ = os.WriteFile(bakPath, prev, 0644)
	}
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return err
	}
	if len(s.lastSaved) == 0 {
		if _, statErr := os.Stat(bakPath); os.IsNotExist(statErr) {
			_ = os.WriteFile(bakPath, data, 0644)
		}
	}
	s.lastSaved = data
	return nil
}
