// Package weather 提供城市真实天气状况在线获取（集成 Open-Meteo 全球开放气象 API）、
// 6 小时本地内存自动更新缓存、离线自愈兜底以及摄氏度/华氏度单位转换。
package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
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
	IsLive      bool      `json:"isLive"`      // 是否来自真实在线气象观测
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

// CoordResolverFunc 根据城市 ID 解析对应的经纬度坐标。
type CoordResolverFunc func(cityID string) (lat, lon float64, ok bool)

// Service 管理城市天气快照内存缓存、异步后台天气拉取与离线气象兜底。
type Service struct {
	mu            sync.RWMutex
	cache         map[string]Snapshot
	inFlight      map[string]bool
	coordResolver CoordResolverFunc
	httpClient    *http.Client
}

// NewService 创建并初始化全球气象服务。
func NewService() *Service {
	s := &Service{
		cache:    make(map[string]Snapshot),
		inFlight: make(map[string]bool),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	s.populateDefaults()
	return s
}

// SetCoordResolver 注入经纬度坐标解析器，用于驱动 Open-Meteo 精确经纬度查询。
func (s *Service) SetCoordResolver(fn CoordResolverFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.coordResolver = fn
}

// GetWeather 获取指定城市的缓存天气快照；若尚未缓存，则按城市 ID 返回初始参考快照并触发异步拉取。
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

// EnsureFresh 获取指定城市的天气快照；若缓存已超过 6 小时或尚未获取过真实在线天气，
// 会在后台协程异步请求 Open-Meteo 真实气象 API，获取完成后调用 onUpdate 触发单次重绘。
func (s *Service) EnsureFresh(cityID string, onUpdate func()) Snapshot {
	s.mu.RLock()
	snap, exists := s.cache[cityID]
	s.mu.RUnlock()

	needsFetch := !exists || !snap.IsLive || time.Since(snap.UpdatedAt) > 6*time.Hour
	if needsFetch {
		s.TriggerFetchAsync(cityID, onUpdate)
	}

	if exists {
		return snap
	}
	return s.GetWeather(cityID)
}

// EnsureFreshCoords 根据指定经纬度获取实时天气快照（支持按 IP 定位坐标获取当地天气），严格每 6 小时自动更新一次。
func (s *Service) EnsureFreshCoords(id string, lat, lon float64, onUpdate func()) Snapshot {
	s.mu.RLock()
	snap, exists := s.cache[id]
	s.mu.RUnlock()

	needsFetch := !exists || !snap.IsLive || time.Since(snap.UpdatedAt) > 6*time.Hour
	if needsFetch {
		s.TriggerFetchCoordsAsync(id, lat, lon, onUpdate)
	}

	if exists {
		return snap
	}
	return s.GetWeather(id)
}

// TriggerFetchCoordsAsync 异步向 Open-Meteo 开放气象 API 发起精确经纬度实时天气观测请求。
func (s *Service) TriggerFetchCoordsAsync(id string, lat, lon float64, onComplete func()) {
	s.mu.Lock()
	if s.inFlight[id] {
		s.mu.Unlock()
		return
	}
	s.inFlight[id] = true
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.inFlight, id)
			s.mu.Unlock()
		}()

		snap, err := s.fetchOpenMeteo(id, lat, lon)
		if err == nil {
			s.mu.Lock()
			s.cache[id] = snap
			s.mu.Unlock()
			if onComplete != nil {
				onComplete()
			}
		}
	}()
}

// TriggerFetchAsync 异步向 Open-Meteo 开放气象 API 发起实时天气观测请求。
func (s *Service) TriggerFetchAsync(cityID string, onComplete func()) {
	s.mu.Lock()
	if s.inFlight[cityID] {
		s.mu.Unlock()
		return
	}
	resolver := s.coordResolver
	if resolver == nil {
		s.mu.Unlock()
		return
	}
	lat, lon, ok := resolver(cityID)
	if !ok {
		s.mu.Unlock()
		return
	}
	s.inFlight[cityID] = true
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.inFlight, cityID)
			s.mu.Unlock()
		}()

		snap, err := s.fetchOpenMeteo(cityID, lat, lon)
		if err == nil {
			s.mu.Lock()
			s.cache[cityID] = snap
			s.mu.Unlock()
			if onComplete != nil {
				onComplete()
			}
		}
	}()
}

type openMeteoResponse struct {
	Current struct {
		Temperature2m      float64 `json:"temperature_2m"`
		RelativeHumidity2m int     `json:"relative_humidity_2m"`
		WeatherCode        int     `json:"weather_code"`
		WindSpeed10m       float64 `json:"wind_speed_10m"`
	} `json:"current"`
}

func (s *Service) fetchOpenMeteo(cityID string, lat, lon float64) (Snapshot, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,weather_code,wind_speed_10m",
		lat, lon,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Snapshot{}, err
	}
	req.Header.Set("User-Agent", "XiaoruIntegratedClockDesktop/1.0 (+https://github.com/vicky-clair/world_Clock_Pro_Desktop_Time_Alarms-Focus_Timer)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return Snapshot{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Snapshot{}, fmt.Errorf("open-meteo http status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return Snapshot{}, err
	}

	var data openMeteoResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return Snapshot{}, err
	}

	cond, icon := parseWMOCode(data.Current.WeatherCode)
	return Snapshot{
		CityID:      cityID,
		Condition:   cond,
		Icon:        icon,
		TempCelsius: float32(data.Current.Temperature2m),
		Humidity:    data.Current.RelativeHumidity2m,
		WindKmH:     float32(data.Current.WindSpeed10m),
		UpdatedAt:   time.Now(),
		IsLive:      true,
	}, nil
}

// parseWMOCode 将世界气象组织 (WMO) 标准天气代码映射为应用的 Condition 与 Emoji 图标。
func parseWMOCode(code int) (Condition, string) {
	switch code {
	case 0:
		return ConditionSunny, "☀️"
	case 1, 2:
		return ConditionPartlyCloudy, "⛅"
	case 3:
		return ConditionCloudy, "☁️"
	case 45, 48:
		return ConditionCloudy, "🌫️"
	case 51, 53, 55, 56, 57, 61, 63, 65, 66, 67, 80, 81, 82:
		return ConditionRainy, "🌧️"
	case 71, 73, 75, 77, 85, 86:
		return ConditionSnowy, "❄️"
	case 95, 96, 99:
		return ConditionThunder, "⛈️"
	default:
		return ConditionSunny, "☀️"
	}
}

// populateDefaults 预填国际主要大都市的基准气象参考数据。
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

	now := time.Now().Add(-2 * time.Hour) // 设为 2 小时前，启动后可自动触发真实刷新
	for _, d := range defaults {
		d.UpdatedAt = now
		s.cache[d.CityID] = d
	}
}

// generateForCity 基于城市 ID 与地理哈希生成断网离线环境下的平滑兜底气象快照。
func (s *Service) generateForCity(cityID string) Snapshot {
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
		UpdatedAt:   time.Now().Add(-2 * time.Hour),
		IsLive:      false,
	}
}
