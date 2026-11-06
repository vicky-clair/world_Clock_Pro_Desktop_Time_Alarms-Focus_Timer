package world

// City represents a city entry with geographic and timezone metadata.
type City struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	NameZH         string  `json:"nameZH,omitempty"`
	NameJA         string  `json:"nameJA,omitempty"`
	Country        string  `json:"country"`
	CountryZH      string  `json:"countryZH,omitempty"`
	CountryJA      string  `json:"countryJA,omitempty"`
	CountryCode    string  `json:"countryCode"`
	Continent      string  `json:"continent,omitempty"`
	IanaTimeZoneID string  `json:"ianaTimeZoneId"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
}

// LocalizedCountry returns the country name in the requested language ("zh", "ja", "en").
func (c City) LocalizedCountry(lang string) string {
	switch lang {
	case "zh":
		if c.CountryZH != "" {
			return c.CountryZH
		}
	case "ja":
		if c.CountryJA != "" {
			return c.CountryJA
		}
		if c.CountryZH != "" {
			return c.CountryZH
		}
	case "en":
		if c.Country != "" {
			return c.Country
		}
	}
	if lang == "zh" && c.CountryZH != "" {
		return c.CountryZH
	}
	if c.Country != "" {
		return c.Country
	}
	return c.Name
}

// LocalizedName returns the city/capital name in the requested language ("zh", "ja", "en").
func (c City) LocalizedName(lang string) string {
	switch lang {
	case "zh":
		if c.NameZH != "" {
			return c.NameZH
		}
	case "ja":
		if c.NameJA != "" {
			return c.NameJA
		}
		if c.NameZH != "" {
			return c.NameZH
		}
	case "en":
		if c.Name != "" {
			return c.Name
		}
	}
	if lang == "zh" && c.NameZH != "" {
		return c.NameZH
	}
	if c.Name != "" {
		return c.Name
	}
	return c.ID
}

// LocalizedContinent returns the continent name in the requested language.
func (c City) LocalizedContinent(lang string) string {
	switch lang {
	case "zh":
		switch c.Continent {
		case "Asia", "亚洲":
			return "亚洲"
		case "Europe", "欧洲":
			return "欧洲"
		case "Africa", "非洲":
			return "非洲"
		case "North America", "北美洲":
			return "北美洲"
		case "South America", "南美洲":
			return "南美洲"
		case "Oceania", "大洋洲":
			return "大洋洲"
		}
		return c.Continent
	case "ja":
		switch c.Continent {
		case "Asia", "亚洲":
			return "アジア"
		case "Europe", "欧洲":
			return "ヨーロッパ"
		case "Africa", "非洲":
			return "アフリカ"
		case "North America", "北美洲":
			return "北米"
		case "South America", "南美洲":
			return "南米"
		case "Oceania", "大洋洲":
			return "オセアニア"
		}
		return c.Continent
	default:
		switch c.Continent {
		case "亚洲":
			return "Asia"
		case "欧洲":
			return "Europe"
		case "非洲":
			return "Africa"
		case "北美洲":
			return "North America"
		case "南美洲":
			return "South America"
		case "大洋洲":
			return "Oceania"
		}
		return c.Continent
	}
}

// BuiltinCities returns a curated catalog of major world cities across all continents.
func BuiltinCities() []City {
	return []City{
		{
			ID:             "beijing",
			Name:           "Beijing",
			NameZH:         "北京",
			NameJA:         "北京",
			Country:        "China",
			CountryZH:      "中国",
			CountryJA:      "中国",
			CountryCode:    "CN",
			Continent:      "亚洲",
			IanaTimeZoneID: "Asia/Shanghai",
			Latitude:       39.9042,
			Longitude:      116.4074,
		},
		{
			ID:             "tokyo",
			Name:           "Tokyo",
			NameZH:         "东京",
			NameJA:         "東京",
			Country:        "Japan",
			CountryZH:      "日本",
			CountryJA:      "日本",
			CountryCode:    "JP",
			Continent:      "亚洲",
			IanaTimeZoneID: "Asia/Tokyo",
			Latitude:       35.6762,
			Longitude:      139.6503,
		},
		{
			ID:             "london",
			Name:           "London",
			NameZH:         "伦敦",
			NameJA:         "ロンドン",
			Country:        "United Kingdom",
			CountryZH:      "英国",
			CountryJA:      "イギリス",
			CountryCode:    "GB",
			Continent:      "欧洲",
			IanaTimeZoneID: "Europe/London",
			Latitude:       51.5074,
			Longitude:      -0.1278,
		},
		{
			ID:             "newyork",
			Name:           "New York",
			NameZH:         "纽约",
			NameJA:         "ニューヨーク",
			Country:        "United States",
			CountryZH:      "美国",
			CountryJA:      "アメリカ",
			CountryCode:    "US",
			Continent:      "北美洲",
			IanaTimeZoneID: "America/New_York",
			Latitude:       40.7128,
			Longitude:      -74.0060,
		},
		{
			ID:             "losangeles",
			Name:           "Los Angeles",
			NameZH:         "洛杉矶",
			NameJA:         "ロサンゼルス",
			Country:        "United States",
			CountryZH:      "美国",
			CountryJA:      "アメリカ",
			CountryCode:    "US",
			Continent:      "北美洲",
			IanaTimeZoneID: "America/Los_Angeles",
			Latitude:       34.0522,
			Longitude:      -118.2437,
		},
		{
			ID:             "paris",
			Name:           "Paris",
			NameZH:         "巴黎",
			NameJA:         "パリ",
			Country:        "France",
			CountryZH:      "法国",
			CountryJA:      "フランス",
			CountryCode:    "FR",
			Continent:      "欧洲",
			IanaTimeZoneID: "Europe/Paris",
			Latitude:       48.8566,
			Longitude:      2.3522,
		},
		{
			ID:             "berlin",
			Name:           "Berlin",
			NameZH:         "柏林",
			NameJA:         "ベルリン",
			Country:        "Germany",
			CountryZH:      "德国",
			CountryJA:      "ドイツ",
			CountryCode:    "DE",
			Continent:      "欧洲",
			IanaTimeZoneID: "Europe/Berlin",
			Latitude:       52.5200,
			Longitude:      13.4050,
		},
		{
			ID:             "sydney",
			Name:           "Sydney",
			NameZH:         "悉尼",
			NameJA:         "シドニー",
			Country:        "Australia",
			CountryZH:      "澳大利亚",
			CountryJA:      "オーストラリア",
			CountryCode:    "AU",
			Continent:      "大洋洲",
			IanaTimeZoneID: "Australia/Sydney",
			Latitude:       -33.8688,
			Longitude:      151.2093,
		},
		{
			ID:             "singapore",
			Name:           "Singapore",
			NameZH:         "新加坡",
			NameJA:         "シンガポール",
			Country:        "Singapore",
			CountryZH:      "新加坡",
			CountryJA:      "シンガポール",
			CountryCode:    "SG",
			Continent:      "亚洲",
			IanaTimeZoneID: "Asia/Singapore",
			Latitude:       1.3521,
			Longitude:      103.8198,
		},
		{
			ID:             "dubai",
			Name:           "Dubai",
			NameZH:         "迪拜",
			NameJA:         "ドバイ",
			Country:        "United Arab Emirates",
			CountryZH:      "阿拉伯联合酋长国",
			CountryJA:      "アラブ首長国連邦",
			CountryCode:    "AE",
			Continent:      "亚洲",
			IanaTimeZoneID: "Asia/Dubai",
			Latitude:       25.2048,
			Longitude:      55.2708,
		},
		{
			ID:             "toronto",
			Name:           "Toronto",
			NameZH:         "多伦多",
			NameJA:         "トロント",
			Country:        "Canada",
			CountryZH:      "加拿大",
			CountryJA:      "カナダ",
			CountryCode:    "CA",
			Continent:      "北美洲",
			IanaTimeZoneID: "America/Toronto",
			Latitude:       43.6532,
			Longitude:      -79.3832,
		},
		{
			ID:             "cairo",
			Name:           "Cairo",
			NameZH:         "开罗",
			NameJA:         "カイロ",
			Country:        "Egypt",
			CountryZH:      "埃及",
			CountryJA:      "エジプト",
			CountryCode:    "EG",
			Continent:      "非洲",
			IanaTimeZoneID: "Africa/Cairo",
			Latitude:       30.0444,
			Longitude:      31.2357,
		},
		{
			ID:             "mumbai",
			Name:           "Mumbai",
			NameZH:         "孟买",
			NameJA:         "ムンバイ",
			Country:        "India",
			CountryZH:      "印度",
			CountryJA:      "インド",
			CountryCode:    "IN",
			Continent:      "亚洲",
			IanaTimeZoneID: "Asia/Kolkata",
			Latitude:       19.0760,
			Longitude:      72.8777,
		},
		{
			ID:             "saopaulo",
			Name:           "São Paulo",
			NameZH:         "圣保罗",
			NameJA:         "サンパウロ",
			Country:        "Brazil",
			CountryZH:      "巴西",
			CountryJA:      "ブラジル",
			CountryCode:    "BR",
			Continent:      "南美洲",
			IanaTimeZoneID: "America/Sao_Paulo",
			Latitude:       -23.5505,
			Longitude:      -46.6333,
		},
		{
			ID:             "seoul",
			Name:           "Seoul",
			NameZH:         "首尔",
			NameJA:         "ソウル",
			Country:        "South Korea",
			CountryZH:      "韩国",
			CountryJA:      "韓国",
			CountryCode:    "KR",
			Continent:      "亚洲",
			IanaTimeZoneID: "Asia/Seoul",
			Latitude:       37.5665,
			Longitude:      126.9780,
		},
		{
			ID:             "auckland",
			Name:           "Auckland",
			NameZH:         "奥克兰",
			NameJA:         "オークランド",
			Country:        "New Zealand",
			CountryZH:      "新西兰",
			CountryJA:      "ニュージーランド",
			CountryCode:    "NZ",
			Continent:      "大洋洲",
			IanaTimeZoneID: "Pacific/Auckland",
			Latitude:       -36.8485,
			Longitude:      174.7633,
		},
		{
			ID:             "shanghai",
			Name:           "Shanghai",
			NameZH:         "上海",
			NameJA:         "上海",
			Country:        "China",
			CountryZH:      "中国",
			CountryJA:      "中国",
			CountryCode:    "CN",
			Continent:      "亚洲",
			IanaTimeZoneID: "Asia/Shanghai",
			Latitude:       31.2304,
			Longitude:      121.4737,
		},
		{
			ID:             "hongkong",
			Name:           "Hong Kong",
			NameZH:         "香港",
			NameJA:         "香港",
			Country:        "China",
			CountryZH:      "中国",
			CountryJA:      "中国",
			CountryCode:    "CN",
			Continent:      "亚洲",
			IanaTimeZoneID: "Asia/Hong_Kong",
			Latitude:       22.3193,
			Longitude:      114.1694,
		},
		{
			ID:             "taipei",
			Name:           "Taipei",
			NameZH:         "台北",
			NameJA:         "台北",
			Country:        "China",
			CountryZH:      "中国",
			CountryJA:      "中国",
			CountryCode:    "CN",
			Continent:      "亚洲",
			IanaTimeZoneID: "Asia/Taipei",
			Latitude:       25.0330,
			Longitude:      121.5654,
		},
		{
			ID:             "osaka",
			Name:           "Osaka",
			NameZH:         "大阪",
			NameJA:         "大阪",
			Country:        "Japan",
			CountryZH:      "日本",
			CountryJA:      "日本",
			CountryCode:    "JP",
			Continent:      "亚洲",
			IanaTimeZoneID: "Asia/Tokyo",
			Latitude:       34.6937,
			Longitude:      135.5023,
		},
		{
			ID:             "sanfrancisco",
			Name:           "San Francisco",
			NameZH:         "旧金山",
			NameJA:         "サンフランシスコ",
			Country:        "United States",
			CountryZH:      "美国",
			CountryJA:      "アメリカ",
			CountryCode:    "US",
			Continent:      "北美洲",
			IanaTimeZoneID: "America/Los_Angeles",
			Latitude:       37.7749,
			Longitude:      -122.4194,
		},
		{
			ID:             "chicago",
			Name:           "Chicago",
			NameZH:         "芝加哥",
			NameJA:         "シカゴ",
			Country:        "United States",
			CountryZH:      "美国",
			CountryJA:      "アメリカ",
			CountryCode:    "US",
			Continent:      "北美洲",
			IanaTimeZoneID: "America/Chicago",
			Latitude:       41.8781,
			Longitude:      -87.6298,
		},
		{
			ID:             "vancouver",
			Name:           "Vancouver",
			NameZH:         "温哥华",
			NameJA:         "バンクーバー",
			Country:        "Canada",
			CountryZH:      "加拿大",
			CountryJA:      "カナダ",
			CountryCode:    "CA",
			Continent:      "北美洲",
			IanaTimeZoneID: "America/Vancouver",
			Latitude:       49.2827,
			Longitude:      -123.1207,
		},
	}
}

var cityAliases = map[string]string{
	"beijing":   "china",
	"tokyo":     "japan",
	"london":    "united-kingdom",
	"paris":     "france",
	"berlin":    "germany",
	"singapore": "singapore",
	"cairo":     "egypt",
	"seoul":     "south-korea",
}

// CanonicalCityID resolves legacy capital city IDs (such as "beijing") to their canonical 193-catalog ID (such as "china").
func CanonicalCityID(id string) string {
	if alias, ok := cityAliases[id]; ok {
		return alias
	}
	return id
}

// SameCityID reports whether two city IDs refer to the same city entry (accounting for legacy aliases).
func SameCityID(a, b string) bool {
	if a == b {
		return true
	}
	return CanonicalCityID(a) == CanonicalCityID(b)
}

// FindCityByID returns a city by its ID from the 193 global countries or builtin cities.
func FindCityByID(id string) (City, bool) {
	// If id is a distinct non-capital builtin city, return its exact metadata first.
	if CanonicalCityID(id) == id {
		for _, c := range BuiltinCities() {
			if c.ID == id {
				return c, true
			}
		}
	}
	targetID := CanonicalCityID(id)
	for _, c := range Global193Countries() {
		if c.ID == targetID || c.ID == id {
			return c, true
		}
	}
	for _, c := range BuiltinCities() {
		if c.ID == id {
			return c, true
		}
	}
	return City{}, false
}
