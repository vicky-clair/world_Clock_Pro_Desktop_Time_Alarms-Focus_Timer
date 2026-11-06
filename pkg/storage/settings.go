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

// ThemeType names the visual theme.
type ThemeType string

const (
	ThemeDark  ThemeType = "dark"
	ThemeLight ThemeType = "light"
	ThemeOlive ThemeType = "olive"
	ThemeNord  ThemeType = "nord"
)

// CloseBehaviorType defines what happens when the user clicks the window close (X) button.
const (
	CloseBehaviorMinimize = "minimize"
	CloseBehaviorExit     = "exit"
)

// Settings represents all persistent user preferences and data.
type Settings struct {
	TimeFormat24H  bool             `json:"timeFormat24H"`
	ShowSeconds    bool             `json:"showSeconds"`
	ShowFullDate   bool             `json:"showFullDate"`
	ClockStyle     string           `json:"clockStyle"` // "digital" or "analog"
	Theme          ThemeType        `json:"theme"`
	ActiveProfile  string           `json:"activeProfile"`
	WorldCities    []world.UserCity `json:"worldCities"`
	Alarms         []alarm.Alarm    `json:"alarms"`
	Topmost        bool             `json:"topmost"`
	PopoutMode     bool             `json:"popoutMode"`
	TempFahrenheit bool             `json:"tempFahrenheit"`
	Language       string           `json:"language"` // "zh", "en", "ja"
	Wallpaper      wallpaper.Config `json:"wallpaper"`
	ClockFontSize  int              `json:"clockFontSize"`  // e.g. 48, 64, 76, 88
	ClockFontColor string           `json:"clockFontColor"` // e.g. "white", "gold", "olive", "cyan", "purple", "coral"
	UIFontColor    string           `json:"uiFontColor"`    // e.g. "default", "white", "black", "gold", "cyan", "purple", "olive"
	FontScale      float32          `json:"fontScale"`      // e.g. 0.9, 1.0, 1.1, 1.2
	CloseBehavior  string           `json:"closeBehavior"`  // "minimize" or "exit"
}

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
		Language:       "zh", // Default to Chinese as user prefers
		Wallpaper:      wallpaper.DefaultConfig(),
		ClockFontSize:  64,
		ClockFontColor: "white",
		UIFontColor:    "default",
		FontScale:      1.0,
		CloseBehavior:  CloseBehaviorMinimize,
	}
}

// Store handles thread-safe loading and persisting of settings.
type Store struct {
	mu       sync.RWMutex
	filePath string
	settings Settings
}

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

// FilePath returns the path to the settings JSON file.
func (s *Store) FilePath() string {
	return s.filePath
}

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

// Update updates the current settings and persists them.
func (s *Store) Update(fn func(*Settings)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.settings)
	return s.saveLocked()
}

func unmarshalSettingsWithDefaults(data []byte) (Settings, error) {
	loaded := DefaultSettings()
	loaded.WorldCities = nil // allow distinguishing JSON "null"/missing vs explicit "[]"
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

// Load loads settings from file if it exists (with automatic backup recovery).
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		// Try backup file if primary does not exist or failed to read
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
		// Primary file corrupted; attempt recovery from backup
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

// Save persists current settings to disk.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked()
}

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
