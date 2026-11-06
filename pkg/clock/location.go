package clock

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// SystemLocation 描述根据当前电脑操作系统区域设置与系统时区自动识别出的所在国家/地区及代表城市信息。
type SystemLocation struct {
	CountryCode string // ISO 3166-1 国家/地区代码（如 "CN", "US", "JP"）
	CountryZH   string // 国家/地区中文名
	CountryEN   string // 国家/地区英文名
	CountryJA   string // 国家/地区日文名
	CityZH      string // 本地时区代表城市中文名
	CityEN      string // 本地时区代表城市英文名
	CityJA      string // 本地时区代表城市日文名
	WeatherID   string // 关联的本地天气查询城市 ID
	UTCOffset   string // 格式化后的 UTC 偏移量（如 "UTC+08:00"）
}

// CountryName 根据当前界面语言 ("zh", "en", "ja") 返回所在国家/地区的本地化名称。
func (s SystemLocation) CountryName(lang string) string {
	switch lang {
	case "zh":
		return s.CountryZH
	case "ja":
		return s.CountryJA
	default:
		return s.CountryEN
	}
}

// CityName 根据当前界面语言 ("zh", "en", "ja") 返回所在时区代表城市的本地化名称。
func (s SystemLocation) CityName(lang string) string {
	switch lang {
	case "zh":
		return s.CityZH
	case "ja":
		return s.CityJA
	default:
		return s.CityEN
	}
}

// FormatBadge 返回用于主时钟和悬浮窗展示的“所在国家/地区 • 本地时区”完整徽章文本。
func (s SystemLocation) FormatBadge(lang string) string {
	switch lang {
	case "zh":
		return fmt.Sprintf("📍 所在国家/地区：%s · %s (%s)", s.CountryZH, s.CityZH, s.UTCOffset)
	case "ja":
		return fmt.Sprintf("📍 現在の国・地域：%s · %s (%s)", s.CountryJA, s.CityJA, s.UTCOffset)
	default:
		return fmt.Sprintf("📍 Country/Region: %s · %s (%s)", s.CountryEN, s.CityEN, s.UTCOffset)
	}
}

type countryMeta struct {
	code      string
	zh        string
	en        string
	ja        string
	cityZH    string
	cityEN    string
	cityJA    string
	weatherID string
}

var knownCountries = map[string]countryMeta{
	"CN": {"CN", "中国", "China", "中国", "北京", "Beijing", "北京", "beijing"},
	"TW": {"TW", "中国台湾", "Taiwan, China", "台湾", "台北", "Taipei", "台北", "taipei"},
	"HK": {"HK", "中国香港", "Hong Kong, China", "香港", "香港", "Hong Kong", "香港", "hongkong"},
	"MO": {"MO", "中国澳门", "Macau, China", "マカオ", "澳门", "Macau", "マカオ", "hongkong"},
	"JP": {"JP", "日本", "Japan", "日本", "东京", "Tokyo", "東京", "tokyo"},
	"KR": {"KR", "韩国", "South Korea", "韓国", "首尔", "Seoul", "ソウル", "seoul"},
	"SG": {"SG", "新加坡", "Singapore", "シンガポール", "新加坡", "Singapore", "シンガポール", "singapore"},
	"MY": {"MY", "马来西亚", "Malaysia", "マレーシア", "吉隆坡", "Kuala Lumpur", "クアラルンプール", "singapore"},
	"TH": {"TH", "泰国", "Thailand", "タイ", "曼谷", "Bangkok", "バンコク", "bangkok"},
	"VN": {"VN", "越南", "Vietnam", "ベトナム", "河内", "Hanoi", "ハノイ", "bangkok"},
	"IN": {"IN", "印度", "India", "インド", "新德里", "New Delhi", "ニューデリー", "mumbai"},
	"AU": {"AU", "澳大利亚", "Australia", "オーストラリア", "悉尼", "Sydney", "シドニー", "sydney"},
	"NZ": {"NZ", "新西兰", "New Zealand", "ニュージーランド", "奥克兰", "Auckland", "オークランド", "sydney"},
	"GB": {"GB", "英国", "United Kingdom", "イギリス", "伦敦", "London", "ロンドン", "london"},
	"UK": {"GB", "英国", "United Kingdom", "イギリス", "伦敦", "London", "ロンドン", "london"},
	"FR": {"FR", "法国", "France", "フランス", "巴黎", "Paris", "パリ", "paris"},
	"DE": {"DE", "德国", "Germany", "ドイツ", "柏林", "Berlin", "ベルリン", "berlin"},
	"IT": {"IT", "意大利", "Italy", "イタリア", "罗马", "Rome", "ローマ", "paris"},
	"ES": {"ES", "西班牙", "Spain", "スペイン", "马德里", "Madrid", "マドリード", "paris"},
	"NL": {"NL", "荷兰", "Netherlands", "オランダ", "阿姆斯特丹", "Amsterdam", "アムステルダム", "paris"},
	"CH": {"CH", "瑞士", "Switzerland", "スイス", "苏黎世", "Zurich", "チューリッヒ", "paris"},
	"SE": {"SE", "瑞典", "Sweden", "スウェーデン", "斯德哥尔摩", "Stockholm", "ストックホルム", "berlin"},
	"RU": {"RU", "俄罗斯", "Russia", "ロシア", "莫斯科", "Moscow", "モスクワ", "moscow"},
	"AE": {"AE", "阿联酋", "United Arab Emirates", "アラブ首長国連邦", "迪拜", "Dubai", "ドバイ", "dubai"},
	"US": {"US", "美国", "United States", "アメリカ合衆国", "纽约", "New York", "ニューヨーク", "newyork"},
	"CA": {"CA", "加拿大", "Canada", "カナダ", "多伦多", "Toronto", "トロント", "toronto"},
	"BR": {"BR", "巴西", "Brazil", "ブラジル", "圣保罗", "Sao Paulo", "サンパウロ", "saopaulo"},
	"MX": {"MX", "墨西哥", "Mexico", "メキシコ", "墨西哥城", "Mexico City", "メキシコシティ", "newyork"},
}

var (
	osRegionOnce sync.Once
	cachedGeo    string
	cachedLocale string
	cachedWinTZ  string
)

// DetectSystemLocation 根据电脑当前的系统时区、Windows 区域设置及偏移量，返回所在国家/地区与代表城市信息。
func DetectSystemLocation(now time.Time) SystemLocation {
	osRegionOnce.Do(func() {
		cachedGeo, cachedLocale, cachedWinTZ = platformDetectOSRegion()
	})
	return resolveLocation(now.Local(), cachedGeo, cachedLocale, cachedWinTZ)
}

func resolveLocation(localTime time.Time, geoCode, localeName, winTZName string) SystemLocation {
	_, offsetSec := localTime.Zone()
	offsetStr := FormatUTCOffset(offsetSec)

	// 1. 优先结合系统时区与 Windows 国家/地区代码判定（避免英文系统但处于中国时区时误判为美国）
	tzLower := strings.ToLower(winTZName + " " + localTime.Location().String())
	code := strings.ToUpper(strings.TrimSpace(geoCode))
	if code == "" {
		code = extractCountryFromLocale(localeName)
	}

	// 根据明确的系统时区标识修正国家与城市
	switch {
	case strings.Contains(tzLower, "china standard time") || strings.Contains(tzLower, "asia/shanghai") || strings.Contains(tzLower, "asia/chongqing"):
		if code != "HK" && code != "MO" && code != "TW" {
			code = "CN"
		}
	case strings.Contains(tzLower, "taipei"):
		code = "TW"
	case strings.Contains(tzLower, "tokyo") || strings.Contains(tzLower, "japan"):
		code = "JP"
	case strings.Contains(tzLower, "korea") || strings.Contains(tzLower, "seoul"):
		code = "KR"
	case strings.Contains(tzLower, "singapore"):
		code = "SG"
	case strings.Contains(tzLower, "gmt standard time") || strings.Contains(tzLower, "europe/london"):
		if code == "" || code == "US" {
			code = "GB"
		}
	}

	// 若系统时区为 UTC+8 且未明确指定其他 UTC+8 国家（或英文系统默认 US），则归属为中国 (CN)
	if offsetSec == 8*3600 && (code == "" || code == "US") {
		code = "CN"
	} else if offsetSec == 9*3600 && (code == "" || code == "US") {
		code = "JP"
	}

	if meta, ok := knownCountries[code]; ok {
		loc := SystemLocation{
			CountryCode: meta.code,
			CountryZH:   meta.zh,
			CountryEN:   meta.en,
			CountryJA:   meta.ja,
			CityZH:      meta.cityZH,
			CityEN:      meta.cityEN,
			CityJA:      meta.cityJA,
			WeatherID:   meta.weatherID,
			UTCOffset:   offsetStr,
		}
		// 对美国/加拿大等多时区国家，按电脑实际 UTC 偏移细分代表城市
		if meta.code == "US" {
			switch offsetSec / 3600 {
			case -8, -7:
				loc.CityZH, loc.CityEN, loc.CityJA, loc.WeatherID = "洛杉矶", "Los Angeles", "ロサンゼルス", "losangeles"
			case -6:
				loc.CityZH, loc.CityEN, loc.CityJA, loc.WeatherID = "芝加哥", "Chicago", "シカゴ", "chicago"
			default:
				loc.CityZH, loc.CityEN, loc.CityJA, loc.WeatherID = "纽约", "New York", "ニューヨーク", "newyork"
			}
		}
		return loc
	}

	// 根据 UTC 偏移量兜底推断所在国家与代表城市
	switch offsetSec / 3600 {
	case 8:
		m := knownCountries["CN"]
		return SystemLocation{m.code, m.zh, m.en, m.ja, m.cityZH, m.cityEN, m.cityJA, m.weatherID, offsetStr}
	case 9:
		m := knownCountries["JP"]
		return SystemLocation{m.code, m.zh, m.en, m.ja, m.cityZH, m.cityEN, m.cityJA, m.weatherID, offsetStr}
	case 0:
		m := knownCountries["GB"]
		return SystemLocation{m.code, m.zh, m.en, m.ja, m.cityZH, m.cityEN, m.cityJA, m.weatherID, offsetStr}
	case 1, 2:
		m := knownCountries["DE"]
		return SystemLocation{m.code, m.zh, m.en, m.ja, m.cityZH, m.cityEN, m.cityJA, m.weatherID, offsetStr}
	case 3:
		m := knownCountries["RU"]
		return SystemLocation{m.code, m.zh, m.en, m.ja, m.cityZH, m.cityEN, m.cityJA, m.weatherID, offsetStr}
	case 4:
		m := knownCountries["AE"]
		return SystemLocation{m.code, m.zh, m.en, m.ja, m.cityZH, m.cityEN, m.cityJA, m.weatherID, offsetStr}
	case 5:
		m := knownCountries["IN"]
		return SystemLocation{m.code, m.zh, m.en, m.ja, m.cityZH, m.cityEN, m.cityJA, m.weatherID, offsetStr}
	case 7:
		m := knownCountries["TH"]
		return SystemLocation{m.code, m.zh, m.en, m.ja, m.cityZH, m.cityEN, m.cityJA, m.weatherID, offsetStr}
	case 10, 11:
		m := knownCountries["AU"]
		return SystemLocation{m.code, m.zh, m.en, m.ja, m.cityZH, m.cityEN, m.cityJA, m.weatherID, offsetStr}
	case -5, -4, -6, -7, -8:
		m := knownCountries["US"]
		return SystemLocation{m.code, m.zh, m.en, m.ja, m.cityZH, m.cityEN, m.cityJA, m.weatherID, offsetStr}
	case -3:
		m := knownCountries["BR"]
		return SystemLocation{m.code, m.zh, m.en, m.ja, m.cityZH, m.cityEN, m.cityJA, m.weatherID, offsetStr}
	default:
		m := knownCountries["CN"]
		return SystemLocation{m.code, m.zh, m.en, m.ja, m.cityZH, m.cityEN, m.cityJA, m.weatherID, offsetStr}
	}
}

func extractCountryFromLocale(locale string) string {
	if locale == "" {
		return ""
	}
	// Strip encoding like ".UTF-8"
	if idx := strings.IndexByte(locale, '.'); idx > 0 {
		locale = locale[:idx]
	}
	locale = strings.ReplaceAll(locale, "_", "-")
	parts := strings.Split(locale, "-")
	for i := len(parts) - 1; i >= 1; i-- {
		p := strings.ToUpper(parts[i])
		if len(p) == 2 {
			return p
		}
	}
	return ""
}

// FormatUTCOffset 将秒级时区偏移量格式化为标准 "UTC+08:00" 或 "UTC-05:00" 字符串。
func FormatUTCOffset(offsetSec int) string {
	sign := "+"
	if offsetSec < 0 {
		sign = "-"
		offsetSec = -offsetSec
	}
	hours := offsetSec / 3600
	mins := (offsetSec % 3600) / 60
	return fmt.Sprintf("UTC%s%02d:%02d", sign, hours, mins)
}
