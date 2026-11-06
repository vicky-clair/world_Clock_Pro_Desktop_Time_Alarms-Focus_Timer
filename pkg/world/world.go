package world

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"worldclock/pkg/clock"
	"worldclock/pkg/i18n"
)

// CityTimeSnapshot holds computed time info for a city at a given instant.
type CityTimeSnapshot struct {
	City           City
	DisplayName    string
	DisplayCountry string
	LocalTime      time.Time
	TimeDisplay    string
	Period         string
	DateDisplay    string
	OffsetStr      string
	DayDiff        string // "Today", "Tomorrow", "Yesterday"
	IsFavorite     bool
	CustomLabel    string
}

// UserCity represents an added city in the user's list.
type UserCity struct {
	CityID      string `json:"cityId"`
	CustomLabel string `json:"customLabel,omitempty"`
	IsFavorite  bool   `json:"isFavorite"`
	SortOrder   int    `json:"sortOrder"`
}

// Manager manages user's world clocks and timezone conversions.
type Manager struct {
	mu       sync.RWMutex
	locCache map[string]*time.Location
	cities   []UserCity
}

// DefaultCities returns the initial default world clock cities.
func DefaultCities() []UserCity {
	return []UserCity{
		{CityID: "beijing", IsFavorite: true, SortOrder: 0},
		{CityID: "tokyo", IsFavorite: false, SortOrder: 1},
		{CityID: "london", IsFavorite: true, SortOrder: 2},
		{CityID: "newyork", IsFavorite: true, SortOrder: 3},
	}
}

// NewManager creates a world clock manager with default cities when initialCities is nil.
// If initialCities is a non-nil empty slice, the user intentionally removed all cities.
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

// Cities returns a copy of current user cities.
func (m *Manager) Cities() []UserCity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]UserCity, len(m.cities))
	copy(res, m.cities)
	return res
}

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
				// Capital alias already represented in Global193Countries
				continue
			}
			if cityMatches(b, q, cont) {
				res = append(res, b)
			}
		}
	}
	return res
}
