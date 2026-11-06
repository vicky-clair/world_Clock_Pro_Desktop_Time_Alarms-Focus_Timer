// Package storage 负责用户配置的并发安全读取、深拷贝快照隔离、JSON 持久化以及 .bak 自动容灾备份与损坏恢复。
package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"worldclock/pkg/alarm"
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
)

// Settings 聚合所有需要持久化到磁盘的用户偏好设置、世界时钟城市列表、闹钟与壁纸轮播配置。
// Settings represents all persistent user preferences and data.
type Settings struct {
	TimeFormat24H  bool             `json:"timeFormat24H"`  // 是否使用 24 小时制
	ShowSeconds    bool             `json:"showSeconds"`    // 是否显示秒数
	ShowFullDate   bool             `json:"showFullDate"`   // 是否显示完整日期格式
	ClockStyle     string           `json:"clockStyle"`     // 时钟样式："digital"（数字）或 "analog"（模拟表盘）
	Theme          ThemeType        `json:"theme"`          // 当前视觉主题
	ActiveProfile  string           `json:"activeProfile"`  // 当前激活的配置文件名称（"Default" 或 "Work"）
	WorldCities    []world.UserCity `json:"worldCities"`    // 用户已添加的世界时钟城市列表
	Alarms         []alarm.Alarm    `json:"alarms"`         // 用户已配置的闹钟列表
	Topmost        bool             `json:"topmost"`        // 悬浮小窗是否置顶
	PopoutMode     bool             `json:"popoutMode"`     // 是否处于紧凑型桌面悬浮小窗模式
	TempFahrenheit bool             `json:"tempFahrenheit"` // 是否使用华氏度 (°F) 显示气温
	Language       string           `json:"language"`       // 界面语言："zh"（中文）、"en"（英文）、"ja"（日文）
	Wallpaper      wallpaper.Config `json:"wallpaper"`      // 壁纸与轮播配置
	ClockFontSize  int              `json:"clockFontSize"`  // 主时钟字号像素（如 48, 64, 76, 88）
	ClockFontColor string           `json:"clockFontColor"` // 主时钟文字配色（如 "white", "gold", "olive", "cyan", "purple", "coral"）
	UIFontColor    string           `json:"uiFontColor"`    // 全局界面文字颜色（如 "default", "white", "black" 等）
	FontScale      float32          `json:"fontScale"`      // 字体缩放比例（如 0.9, 1.0, 1.1, 1.2）
	CloseBehavior  string           `json:"closeBehavior"`  // 关闭按钮行为："minimize"（最小化到托盘）或 "exit"（直接退出）
}

// DefaultSettings 返回包含默认世界时钟城市与推荐偏好的初始配置。
// DefaultSettings returns fresh default settings.
func DefaultSettings() Settings {
	return Settings{
		TimeFormat24H:  true,
		ShowSeconds:    true,
		ShowFullDate:   true,
		ClockStyle:     "digital",
		Theme:          ThemeOlive,
		ActiveProfile:  "Default",
		WorldCities:    world.DefaultCities(),
		Topmost:        false,
		PopoutMode:     false,
		TempFahrenheit: false,
		Language:       "zh", // 默认使用简体中文界面
		Wallpaper:      wallpaper.DefaultConfig(),
		ClockFontSize:  64,
		ClockFontColor: "white",
		UIFontColor:    "default",
		FontScale:      1.0,
		CloseBehavior:  CloseBehaviorMinimize,
	}
}

// Store 负责并发安全地加载、更新和持久化用户配置。
// Store handles thread-safe loading and persisting of settings.
type Store struct {
	mu       sync.RWMutex
	filePath string
	settings Settings
}

// NewStore 创建配置存储实例；若 filePath 为空，则默认使用用户配置目录下的 settings.json。
// NewStore creates a new storage store. If path is empty, uses the user config directory.
func NewStore(filePath string) *Store {
	if filePath == "" {
		cfgDir, err := os.UserConfigDir()
		if err != nil || cfgDir == "" {
			cfgDir = "."
		}
		filePath = filepath.Join(cfgDir, "OliveClock", "settings.json")
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
	if err := json.Unmarshal(data, &loaded); err != nil {
		return Settings{}, err
	}
	if loaded.WorldCities == nil {
		loaded.WorldCities = world.DefaultCities()
	}
	if loaded.CloseBehavior != CloseBehaviorMinimize && loaded.CloseBehavior != CloseBehaviorExit {
		loaded.CloseBehavior = CloseBehaviorMinimize
	}
	if loaded.ClockFontSize <= 0 {
		loaded.ClockFontSize = 64
	}
	if loaded.FontScale <= 0 {
		loaded.FontScale = 1.0
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
				_ = s.saveLocked()
				return nil
			}
		}
		return err
	}

	s.settings = loaded
	return nil
}

// Save 将当前内存中的配置持久化写入磁盘。
// Save persists current settings to disk.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

// saveLocked 在覆写主配置文件前，先将上一份有效配置备份至 .bak 文件，确保任何时刻至少有一份完整有效的数据。
func (s *Store) saveLocked() error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return err
	}

	bakPath := s.filePath + ".bak"
	if prev, readErr := os.ReadFile(s.filePath); readErr == nil && len(prev) > 0 {
		_ = os.WriteFile(bakPath, prev, 0644)
	}
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return err
	}
	if _, statErr := os.Stat(bakPath); os.IsNotExist(statErr) {
		_ = os.WriteFile(bakPath, data, 0644)
	}
	return nil
}
