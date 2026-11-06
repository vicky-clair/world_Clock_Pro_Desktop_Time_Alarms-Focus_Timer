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
		Topmost:        false,
		PopoutMode:     false,
		TempFahrenheit: false,
		Language:       "zh", // Default to Chinese as user prefers
		Wallpaper:      wallpaper.DefaultConfig(),
		ClockFontSize:  64,
		ClockFontColor: "white",
		UIFontColor:    "default",
		FontScale:      1.0,
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

// Get returns a copy of current settings.
func (s *Store) Get() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

// Update updates the current settings and persists them.
func (s *Store) Update(fn func(*Settings)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.settings)
	return s.saveLocked()
}

// Load loads settings from file if it exists.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var loaded Settings
	if err := json.Unmarshal(data, &loaded); err != nil {
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

	return os.WriteFile(s.filePath, data, 0644)
}
