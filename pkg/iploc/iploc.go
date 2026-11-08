// Package iploc 提供基于公网 IP 的地理位置（国家、城市、经纬度）异步探测与缓存。
package iploc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Info 保存公网 IP 探测结果。
type Info struct {
	IP          string    `json:"query"`
	Country     string    `json:"country"`     // 本地化国家名（如 "中国", "新加坡", "United States"）
	CountryCode string    `json:"countryCode"` // 如 "CN", "SG", "US"
	RegionName  string    `json:"regionName"`  // 省/州名
	City        string    `json:"city"`        // 城市名
	Latitude    float64   `json:"lat"`         // 纬度
	Longitude   float64   `json:"lon"`         // 经度
	Timezone    string    `json:"timezone"`    // 时区标识
	ISP         string    `json:"isp"`         // 网络运营商
	UpdatedAt   time.Time `json:"updatedAt"`   // 更新时间
	IsSuccess   bool      `json:"isSuccess"`   // 是否探测成功
}

// MaskIP 对 IP 地址进行脱敏显示（如 "45.129.228.101" -> "45.129.***.101"），保护用户网络隐私。
func MaskIP(ip string) string {
	ip = strings.TrimSpace(ip)
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.***.%s", parts[0], parts[1], parts[3])
	}
	if strings.Contains(ip, ":") {
		colonParts := strings.Split(ip, ":")
		if len(colonParts) >= 4 {
			return fmt.Sprintf("%s:%s:****:%s", colonParts[0], colonParts[1], colonParts[len(colonParts)-1])
		}
	}
	return ip
}

// FormatLocation 返回包含“国家 省份 城市”的完整地理位置文本。
// 例如："中国 广东 广州"、"中国 北京"、"新加坡"、"美国 加利福尼亚 洛杉矶"。
func (info Info) FormatLocation() string {
	if !info.IsSuccess || info.Country == "" {
		return ""
	}
	// 城市国家（如新加坡、摩纳哥等），国家名与城市名相同，直接返回国家名
	if strings.EqualFold(info.City, info.Country) {
		return info.Country
	}

	var parts []string
	parts = append(parts, info.Country)

	// 若省份存在且不与国家名或城市名重复，则追加省份
	if info.RegionName != "" &&
		!strings.EqualFold(info.RegionName, info.Country) &&
		!strings.EqualFold(info.RegionName, info.City) {
		parts = append(parts, info.RegionName)
	}

	// 追加城市名
	if info.City != "" && !strings.EqualFold(info.City, info.Country) {
		parts = append(parts, info.City)
	}

	return strings.Join(parts, " ")
}

// FormatBadge 根据当前界面语言返回 IP 所在国家的展示徽章文本（包含完整公网 IP 地址）。
func (info Info) FormatBadge(lang string) string {
	if !info.IsSuccess || info.Country == "" {
		return ""
	}
	ipStr := strings.TrimSpace(info.IP)
	switch lang {
	case "zh":
		if ipStr != "" {
			return fmt.Sprintf("🌐 IP 所在国家：%s (%s)", info.Country, ipStr)
		}
		return fmt.Sprintf("🌐 IP 所在国家：%s", info.Country)
	case "ja":
		if ipStr != "" {
			return fmt.Sprintf("🌐 IP の国：%s (%s)", info.Country, ipStr)
		}
		return fmt.Sprintf("🌐 IP の国：%s", info.Country)
	default:
		if ipStr != "" {
			return fmt.Sprintf("🌐 IP Country: %s (%s)", info.Country, ipStr)
		}
		return fmt.Sprintf("🌐 IP Country: %s", info.Country)
	}
}

// Service 管理公网 IP 定位的后台异步拉取与内存缓存。
type Service struct {
	mu         sync.RWMutex
	cached     Info
	cachedLang string
	inFlight   bool
	httpClient *http.Client
	fetchFunc  func(lang string) (Info, error) // 用于单元测试注入 mock
}

// NewService 创建 IP 定位服务实例。
func NewService() *Service {
	return &Service{
		httpClient: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

// Info 返回当前已缓存的 IP 信息快照。
func (s *Service) Info() Info {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cached
}

// SetMockFetcher 用于测试环境注入模拟数据。
func (s *Service) SetMockFetcher(fn func(lang string) (Info, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fetchFunc = fn
}

// SetCached 用于直接设置或预加载已解析的 IP 定位信息（供测试与特定离线预设使用）。
func (s *Service) SetCached(info Info, lang string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	info.IsSuccess = true
	info.UpdatedAt = time.Now()
	s.cached = info
	s.cachedLang = lang
}

// EnsureFresh 获取当前缓存的 IP 定位信息；若未获取过或缓存超过 6 小时，后台异步触发一次刷新。
func (s *Service) EnsureFresh(lang string, onUpdate func()) Info {
	s.mu.RLock()
	info := s.cached
	langChanged := s.cachedLang != "" && s.cachedLang != lang
	needsFetch := !info.IsSuccess || langChanged || time.Since(info.UpdatedAt) > 6*time.Hour
	s.mu.RUnlock()

	if needsFetch {
		s.TriggerFetchAsync(lang, onUpdate)
	}
	return info
}

// TriggerFetchAsync 后台异步拉取 IP 定位信息。
func (s *Service) TriggerFetchAsync(lang string, onComplete func()) {
	s.mu.Lock()
	if s.inFlight {
		s.mu.Unlock()
		return
	}
	s.inFlight = true
	fetcher := s.fetchFunc
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			s.inFlight = false
			s.mu.Unlock()
		}()

		var info Info
		var err error
		if fetcher != nil {
			info, err = fetcher(lang)
		} else {
			info, err = s.doFetch(lang)
		}

		if err == nil && info.Country != "" {
			info.IsSuccess = true
			info.UpdatedAt = time.Now()
			s.mu.Lock()
			s.cached = info
			s.cachedLang = lang
			s.mu.Unlock()

			if onComplete != nil {
				onComplete()
			}
		}
	}()
}

func (s *Service) doFetch(lang string) (Info, error) {
	apiLang := "zh-CN"
	switch lang {
	case "en":
		apiLang = "en"
	case "ja":
		apiLang = "ja"
	}

	url := fmt.Sprintf("http://ip-api.com/json/?lang=%s", apiLang)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Info{}, err
	}
	req.Header.Set("User-Agent", "XiaoruIntegratedClockDesktop/1.0 (+https://github.com/vicky-clair/world_Clock_Pro_Desktop_Time_Alarms-Focus_Timer)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return Info{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Info{}, fmt.Errorf("ip-api status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Info{}, err
	}

	var raw struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		RegionName  string  `json:"regionName"`
		City        string  `json:"city"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Timezone    string  `json:"timezone"`
		ISP         string  `json:"isp"`
		Query       string  `json:"query"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return Info{}, err
	}
	if raw.Status != "success" {
		return Info{}, fmt.Errorf("ip-api returned status %s", raw.Status)
	}

	return Info{
		IP:          raw.Query,
		Country:     raw.Country,
		CountryCode: raw.CountryCode,
		RegionName:  raw.RegionName,
		City:        raw.City,
		Latitude:    raw.Lat,
		Longitude:   raw.Lon,
		Timezone:    raw.Timezone,
		ISP:         raw.ISP,
	}, nil
}
