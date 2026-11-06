// Package world 提供全球 193 个国家与核心城市的时区换算、夏令时识别、多语言模糊搜索、判重及排序管理。
package world

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"worldclock/pkg/clock"
	"worldclock/pkg/i18n"
)

// CityTimeSnapshot 保存某个城市在特定时刻计算出的本地时间、UTC 偏移、跨日状态与本地化名称快照。
// CityTimeSnapshot holds computed time info for a city at a given instant.
type CityTimeSnapshot struct {
	City           City      // 城市基础地理与时区元数据
	DisplayName    string    // 当前语言下的城市/首都显示名（或用户自定义标签）
	DisplayCountry string    // 当前语言下的所属国家显示名
	LocalTime      time.Time // 转换至该城市 IANA 时区后的本地时间
	TimeDisplay    string    // 格式化后的时间字符串（如 "14:30:05"）
	Period         string    // 12 小时制下的 "AM" / "PM"（24 小时制下为空）
	DateDisplay    string    // 本地化日期字符串
	OffsetStr      string    // 相对 UTC 的偏移字符串（如 "UTC+8"、"UTC+5.5"）
	DayDiff        string    // 相对本地日期的跨日标识："Today"、"Tomorrow"、"Yesterday"
	IsFavorite     bool      // 是否被用户标星收藏
	CustomLabel    string    // 用户自定义备注名
}

// UserCity 表示用户添加到世界时钟列表中的单条城市记录。
// UserCity represents an added city in the user's list.
type UserCity struct {
	CityID      string `json:"cityId"`                // 城市或国家唯一 ID
	CustomLabel string `json:"customLabel,omitempty"` // 自定义显示标签（可选）
	IsFavorite  bool   `json:"isFavorite"`            // 是否收藏置顶显示
	SortOrder   int    `json:"sortOrder"`             // 排序序号
}

// Manager 负责管理用户的世界时钟城市列表、IANA 时区对象缓存以及时间快照计算。
// Manager manages user's world clocks and timezone conversions.
type Manager struct {
	mu       sync.RWMutex
	locCache map[string]*time.Location // IANA 时区缓存，避免每帧重复调用 time.LoadLocation
	cities   []UserCity                // 当前用户的世界时钟城市列表
}

// DefaultCities 返回首次启动时的 4 个默认世界时钟城市（北京、东京、伦敦、纽约）。
// DefaultCities returns the initial default world clock cities.
func DefaultCities() []UserCity {
	return []UserCity{
		{CityID: "beijing", IsFavorite: true, SortOrder: 0},
		{CityID: "tokyo", IsFavorite: false, SortOrder: 1},
		{CityID: "london", IsFavorite: true, SortOrder: 2},
		{CityID: "newyork", IsFavorite: true, SortOrder: 3},
	}
}

// NewManager 创建世界时钟管理器；仅当 initialCities 为 nil 时填充默认城市，
// 若 initialCities 是非 nil 的空切片（[]UserCity{}），则保留为空（代表用户主动删除了所有城市）。
// NewManager creates a world clock manager with default cities when initialCities is nil.
func NewManager(initialCities []UserCity) *Manager {
	cities := initialCities
	if cities == nil {
		cities = DefaultCities()
	}
	m := &Manager{
		locCache: make(map[string]*time.Location),
		cities:   cities,
	}
	return m
}

// getLocation 从缓存中获取或加载指定的 IANA 时区对象（若加载失败则安全回退至 UTC）。
func (m *Manager) getLocation(ianaID string) *time.Location {
	m.mu.RLock()
	loc, ok := m.locCache[ianaID]
	m.mu.RUnlock()
	if ok {
		return loc
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if loc, ok := m.locCache[ianaID]; ok {
		return loc
	}

	loadedLoc, err := time.LoadLocation(ianaID)
	if err != nil {
		loadedLoc = time.UTC
	}
	m.locCache[ianaID] = loadedLoc
	return loadedLoc
}

// Cities 返回当前用户世界时钟城市列表的深拷贝切片。
// Cities returns a copy of current user cities.
func (m *Manager) Cities() []UserCity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]UserCity, len(m.cities))
	copy(res, m.cities)
	return res
}

// HasCity 检查指定 cityID（或其首都/国家同义词别名）是否已经存在于用户的世界时钟列表中。
// HasCity reports whether cityID (or an alias of it) is already in the user's world clock list.
func (m *Manager) HasCity(cityID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, c := range m.cities {
		if SameCityID(c.CityID, cityID) {
			return true
		}
	}
	return false
}

// AddCity 将城市添加到用户世界时钟列表；若该城市（或同义词 ID）已存在则返回 false 且不重复添加。
// AddCity adds a city to the user list if not already present.
func (m *Manager) AddCity(cityID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.cities {
		if SameCityID(c.CityID, cityID) {
			return false
		}
	}
	m.cities = append(m.cities, UserCity{
		CityID:    cityID,
		SortOrder: len(m.cities),
	})
	return true
}

// RemoveCity 根据城市 ID（兼容首都/国家别名）从列表中移除对应城市，并重新整理 SortOrder。
// RemoveCity removes a city by ID (supporting both canonical and legacy alias IDs).
func (m *Manager) RemoveCity(cityID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	idx := -1
	for i, c := range m.cities {
		if SameCityID(c.CityID, cityID) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false
	}
	m.cities = append(m.cities[:idx], m.cities[idx+1:]...)
	for i := range m.cities {
		m.cities[i].SortOrder = i
	}
	return true
}

// ToggleFavorite 切换指定城市的星标收藏状态。
// ToggleFavorite toggles the favorite status of a city.
func (m *Manager) ToggleFavorite(cityID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.cities {
		if SameCityID(m.cities[i].CityID, cityID) {
			m.cities[i].IsFavorite = !m.cities[i].IsFavorite
			return
		}
	}
}

// MoveUp 将指定城市在列表中向上移动一位。
// MoveUp moves a city up in sort order.
func (m *Manager) MoveUp(cityID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.cities {
		if SameCityID(m.cities[i].CityID, cityID) && i > 0 {
			m.cities[i], m.cities[i-1] = m.cities[i-1], m.cities[i]
			m.cities[i].SortOrder = i
			m.cities[i-1].SortOrder = i - 1
			return
		}
	}
}

// MoveDown 将指定城市在列表中向下移动一位。
// MoveDown moves a city down in sort order.
func (m *Manager) MoveDown(cityID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.cities {
		if SameCityID(m.cities[i].CityID, cityID) && i < len(m.cities)-1 {
			m.cities[i], m.cities[i+1] = m.cities[i+1], m.cities[i]
			m.cities[i].SortOrder = i
			m.cities[i+1].SortOrder = i + 1
			return
		}
	}
}

// Snapshots 根据给定的 UTC 时间、本地当前时间、12/24 小时制和语言设置，计算所有用户城市的实时显示快照。
// Snapshots computes the time snapshot for all user cities in the requested language.
func (m *Manager) Snapshots(utcInstant time.Time, localNow time.Time, is24H, showSeconds bool, lang string) []CityTimeSnapshot {
	m.mu.RLock()
	userList := make([]UserCity, len(m.cities))
	copy(userList, m.cities)
	m.mu.RUnlock()

	snapshots := make([]CityTimeSnapshot, 0, len(userList))
	for _, uc := range userList {
		city, ok := FindCityByID(uc.CityID)
		if !ok {
			continue
		}
		city.ID = uc.CityID
		loc := m.getLocation(city.IanaTimeZoneID)
		localCityTime := utcInstant.In(loc)

		timeStr, period := clock.FormatTime(localCityTime, is24H, showSeconds)
		dateStr := i18n.FormatDateLocalized(i18n.Language(lang), int(localCityTime.Month()), localCityTime.Day(), localCityTime.Year(), int(localCityTime.Weekday()), false)

		_, offsetSec := localCityTime.Zone()
		offsetHours := float32(offsetSec) / 3600.0
		var offsetStr string
		if offsetHours >= 0 {
			if float32(int(offsetHours)) == offsetHours {
				offsetStr = fmt.Sprintf("UTC+%d", int(offsetHours))
			} else {
				offsetStr = fmt.Sprintf("UTC+%.1f", offsetHours)
			}
		} else {
			if float32(int(offsetHours)) == offsetHours {
				offsetStr = fmt.Sprintf("UTC%d", int(offsetHours))
			} else {
				offsetStr = fmt.Sprintf("UTC%.1f", offsetHours)
			}
		}

		// 计算该城市相对本地日期是今天、明天还是昨天
		dayDiff := "Today"
		cityDay := localCityTime.YearDay()
		cityYear := localCityTime.Year()
		localDay := localNow.YearDay()
		localYear := localNow.Year()

		if cityYear > localYear || (cityYear == localYear && cityDay > localDay) {
			dayDiff = "Tomorrow"
		} else if cityYear < localYear || (cityYear == localYear && cityDay < localDay) {
			dayDiff = "Yesterday"
		}

		cityName := city.LocalizedName(lang)
		if uc.CustomLabel != "" {
			cityName = uc.CustomLabel
		}
		countryName := city.LocalizedCountry(lang)

		snapshots = append(snapshots, CityTimeSnapshot{
			City:           city,
			DisplayName:    cityName,
			DisplayCountry: countryName,
			LocalTime:      localCityTime,
			TimeDisplay:    timeStr,
			Period:         period,
			DateDisplay:    dateStr,
			OffsetStr:      offsetStr,
			DayDiff:        dayDiff,
			IsFavorite:     uc.IsFavorite,
			CustomLabel:    uc.CustomLabel,
		})
	}
	return snapshots
}

// cityMatches 判断城市 c 是否匹配小写的搜索关键词 q（支持中/英/日名称、国家代码、IANA 时区）与大洲筛选条件 cont。
// cityMatches reports whether city c matches lowercase query q and optional continent filter.
func cityMatches(c City, q, cont string) bool {
	if cont != "" && cont != "All" && cont != "全部" && cont != "すべて" {
		if !strings.EqualFold(c.Continent, cont) && !strings.Contains(c.Continent, cont) {
			return false
		}
	}
	if q == "" {
		return true
	}
	return strings.Contains(strings.ToLower(c.Name), q) ||
		strings.Contains(strings.ToLower(c.NameZH), q) ||
		strings.Contains(strings.ToLower(c.NameJA), q) ||
		strings.Contains(strings.ToLower(c.Country), q) ||
		strings.Contains(strings.ToLower(c.CountryZH), q) ||
		strings.Contains(strings.ToLower(c.CountryJA), q) ||
		strings.Contains(strings.ToLower(c.CountryCode), q) ||
		strings.Contains(strings.ToLower(c.IanaTimeZoneID), q) ||
		strings.Contains(strings.ToLower(c.ID), q)
}

// SearchCatalog 在全球 193 个国家（以及用户输入搜索词时的内置非首都重要城市，如上海、香港、大阪、旧金山等）中检索匹配项。
// SearchCatalog searches the 193 global countries (and major world cities when querying) matching query and optional continent.
func SearchCatalog(query string, continent string) []City {
	q := strings.ToLower(strings.TrimSpace(query))
	cont := strings.TrimSpace(continent)
	all := Global193Countries()

	var res []City
	for _, c := range all {
		if cityMatches(c, q, cont) {
			res = append(res, c)
		}
	}
	if q != "" {
		for _, b := range BuiltinCities() {
			if CanonicalCityID(b.ID) != b.ID {
				// 首都别名已在 Global193Countries 中包含，跳过以避免搜索列表出现重复项
				continue
			}
			if cityMatches(b, q, cont) {
				res = append(res, b)
			}
		}
	}
	return res
}
