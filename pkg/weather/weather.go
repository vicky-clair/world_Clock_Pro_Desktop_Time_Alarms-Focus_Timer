// Package weather 提供城市天气状况快照缓存、摄氏度/华氏度转换以及基于城市名的离线确定性气象生成。
package weather

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// Condition 表示标准天气状况枚举。
// Condition represents standard atmospheric condition.
type Condition string

const (
	ConditionSunny        Condition = "Sunny"         // 晴朗
	ConditionPartlyCloudy Condition = "Partly Cloudy" // 多云转晴
	ConditionCloudy       Condition = "Cloudy"        // 阴天
	ConditionRainy        Condition = "Rainy"         // 降雨
	ConditionSnowy        Condition = "Snowy"         // 降雪
	ConditionThunder      Condition = "Thunderstorm"  // 雷阵雨
)

// Snapshot 表示单个城市的天气观测快照数据。
// Snapshot represents weather observations for a location.
type Snapshot struct {
	CityID      string    `json:"cityId"`      // 城市唯一 ID
	Condition   Condition `json:"condition"`   // 天气状况枚举
	Icon        string    `json:"icon"`        // 天气 Emoji 图标
	TempCelsius float32   `json:"tempCelsius"` // 摄氏温度 (°C)
	Humidity    int       `json:"humidity"`    // 相对湿度百分比
	WindKmH     float32   `json:"windKmH"`     // 风速 (km/h)
	UpdatedAt   time.Time `json:"updatedAt"`   // 更新时间戳
}

// FormatTemperature 根据用户偏好将摄氏温度格式化为摄氏度 (°C) 或华氏度 (°F) 字符串。
// FormatTemperature formats temperature in Celsius or Fahrenheit.
func FormatTemperature(celsius float32, fahrenheit bool) string {
	if fahrenheit {
		f := celsius*9.0/5.0 + 32.0
		return fmt.Sprintf("%.0f°F", f)
	}
	return fmt.Sprintf("%.0f°C", celsius)
}

// Service 管理城市天气快照内存缓存与离线气象生成。
// Service manages weather data, local caching, and offline generation.
type Service struct {
	mu    sync.RWMutex
	cache map[string]Snapshot
}

// NewService 创建并预加载全球核心城市天气快照的气象服务。
// NewService creates a weather service with preloaded weather snapshots for world cities.
func NewService() *Service {
	s := &Service{
		cache: make(map[string]Snapshot),
	}
	s.populateDefaults()
	return s
}

// GetWeather 获取指定城市的缓存天气快照；若尚未缓存，则按城市 ID 生成稳定的确定性气象快照。
// GetWeather retrieves cached weather for a city or generates an initial snapshot.
func (s *Service) GetWeather(cityID string) Snapshot {
	s.mu.RLock()
	snap, ok := s.cache[cityID]
	s.mu.RUnlock()
	if ok {
		return snap
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if snap, ok := s.cache[cityID]; ok {
		return snap
	}

	snap = s.generateForCity(cityID)
	s.cache[cityID] = snap
	return snap
}

// populateDefaults 预填常见国际大都市的基准天气数据。
func (s *Service) populateDefaults() {
	defaults := []Snapshot{
		{CityID: "beijing", Condition: ConditionSunny, Icon: "☀️", TempCelsius: 16.0, Humidity: 45, WindKmH: 12.0},
		{CityID: "tokyo", Condition: ConditionPartlyCloudy, Icon: "⛅", TempCelsius: 18.0, Humidity: 60, WindKmH: 10.0},
		{CityID: "london", Condition: ConditionRainy, Icon: "🌧️", TempCelsius: 12.0, Humidity: 85, WindKmH: 18.0},
		{CityID: "newyork", Condition: ConditionSunny, Icon: "☀️", TempCelsius: 15.0, Humidity: 50, WindKmH: 14.0},
		{CityID: "losangeles", Condition: ConditionSunny, Icon: "☀️", TempCelsius: 24.0, Humidity: 40, WindKmH: 8.0},
		{CityID: "paris", Condition: ConditionCloudy, Icon: "☁️", TempCelsius: 13.0, Humidity: 70, WindKmH: 15.0},
		{CityID: "berlin", Condition: ConditionPartlyCloudy, Icon: "⛅", TempCelsius: 11.0, Humidity: 65, WindKmH: 16.0},
		{CityID: "sydney", Condition: ConditionSunny, Icon: "☀️", TempCelsius: 22.0, Humidity: 55, WindKmH: 20.0},
		{CityID: "singapore", Condition: ConditionThunder, Icon: "⛈️", TempCelsius: 30.0, Humidity: 88, WindKmH: 6.0},
		{CityID: "dubai", Condition: ConditionSunny, Icon: "☀️", TempCelsius: 32.0, Humidity: 35, WindKmH: 15.0},
		{CityID: "toronto", Condition: ConditionPartlyCloudy, Icon: "⛅", TempCelsius: 10.0, Humidity: 58, WindKmH: 14.0},
	}

	now := time.Now()
	for _, d := range defaults {
		d.UpdatedAt = now
		s.cache[d.CityID] = d
	}
}

// generateForCity 基于城市 ID 字符哈希生成稳定的离线天气参考快照。
func (s *Service) generateForCity(cityID string) Snapshot {
	// Stable pseudo-hash based on city ID characters
	var sum int
	for _, c := range cityID {
		sum += int(c)
	}
	temp := 12.0 + float32(math.Mod(float64(sum), 16.0))
	cond := ConditionSunny
	icon := "☀️"
	switch sum % 5 {
	case 1:
		cond = ConditionPartlyCloudy
		icon = "⛅"
	case 2:
		cond = ConditionCloudy
		icon = "☁️"
	case 3:
		cond = ConditionRainy
		icon = "🌧️"
	case 4:
		cond = ConditionThunder
		icon = "⛈️"
	}

	return Snapshot{
		CityID:      cityID,
		Condition:   cond,
		Icon:        icon,
		TempCelsius: temp,
		Humidity:    50 + (sum % 30),
		WindKmH:     10 + float32(sum%15),
		UpdatedAt:   time.Now(),
	}
}
