// Package clock 提供世界各国传统节日、法定假日、二十四节气以及现代网络特色节日（如 1024 程序员节、黑五打折、双11等）的识别与格式化功能。
package clock

import (
	"fmt"
	"strings"
	"time"
)

// HolidayKind 定义节日类型分类。
type HolidayKind string

const (
	KindTraditional   HolidayKind = "traditional" // 传统民俗与法定公休节日（如春节、中秋节、国庆节、元旦等）
	KindSolarTerm     HolidayKind = "solar_term"  // 中国二十四节气（立冬、冬至、春分、夏至等）
	KindInternet      HolidayKind = "internet"    // 现代科技与网络流行节日（1024程序员节、520表白日、圆周率日等）
	KindShopping      HolidayKind = "shopping"    // 全球知名电商与打折促销购物节（黑色星期五、网络星期一、双11、双12、618等）
	KindInternational HolidayKind = "intl"        // 国际公认与联合国节日（妇女节、儿童节、地球日、和平日等）
)

// Holiday 存储单个节日的本地化名称、图标与分类。
type Holiday struct {
	NameZH string      // 中文名称
	NameEN string      // 英文名称
	NameJA string      // 日文名称
	Kind   HolidayKind // 节日分类
	Icon   string      // 醒目 Emoji 图标（如 💻, 🛍️, ❄️, 🧧）
}

// LocalizedName 根据语言代码返回节日的本地化名称。
func (h Holiday) LocalizedName(lang string) string {
	switch lang {
	case "zh":
		return h.NameZH
	case "ja":
		if h.NameJA != "" {
			return h.NameJA
		}
		return h.NameZH
	default:
		if h.NameEN != "" {
			return h.NameEN
		}
		return h.NameZH
	}
}

// 21 世纪（2000-2099）二十四节气天文计算常数
var solarTermCoeffs = []struct {
	month int
	c     float64
	zh    string
	en    string
	ja    string
	icon  string
}{
	{1, 5.4055, "小寒", "Minor Cold", "小寒", "❄️"},
	{1, 20.12, "大寒", "Major Cold", "大寒", "❄️"},
	{2, 3.87, "立春", "Start of Spring", "立春", "🌱"},
	{2, 18.73, "雨水", "Rain Water", "雨水", "🌧️"},
	{3, 5.63, "惊蛰", "Awakening of Insects", "啓蟄", "⚡"},
	{3, 20.646, "春分", "Spring Equinox", "春分", "🌸"},
	{4, 4.81, "清明", "Pure Brightness", "清明", "🌿"},
	{4, 20.1, "谷雨", "Grain Rain", "穀雨", "🌧️"},
	{5, 5.52, "立夏", "Start of Summer", "立夏", "☀️"},
	{5, 21.04, "小满", "Grain Buds", "小満", "🌾"},
	{6, 5.678, "芒种", "Grain in Ear", "芒種", "🌾"},
	{6, 21.37, "夏至", "Summer Solstice", "夏至", "☀️"},
	{7, 7.108, "小暑", "Minor Heat", "小暑", "🌡️"},
	{7, 22.83, "大暑", "Major Heat", "大暑", "🔥"},
	{8, 7.5, "立秋", "Start of Autumn", "立秋", "🍂"},
	{8, 23.13, "处暑", "End of Heat", "処暑", "🌾"},
	{9, 7.646, "白露", "White Dew", "白露", "💧"},
	{9, 23.042, "秋分", "Autumn Equinox", "秋分", "🍁"},
	{10, 8.318, "寒露", "Cold Dew", "寒露", "🍂"},
	{10, 23.438, "霜降", "Frost's Descent", "霜降", "❄️"},
	{11, 7.438, "立冬", "Start of Winter", "立冬", "❄️"},
	{11, 22.36, "小雪", "Minor Snow", "小雪", "🌨️"},
	{12, 7.18, "大雪", "Major Snow", "大雪", "🌨️"},
	{12, 21.94, "冬至", "Winter Solstice", "冬至", "🥟"},
}

// computeSolarTermDay 采用百年高精度天文经验公式计算指定年份对应节气的准确公历日期。
func computeSolarTermDay(year int, month int, c float64) int {
	y := year % 100
	d := int(float64(y)*0.2422+c) - int((y-1)/4)
	return d
}

// thanksgivingDay 返回指定年份 11 月第四个星期四（感恩节）的日期号。
func thanksgivingDay(year int) int {
	firstDay := time.Date(year, time.November, 1, 0, 0, 0, 0, time.UTC)
	firstWeekday := int(firstDay.Weekday())
	offset := (4 - firstWeekday + 7) % 7
	firstThursday := 1 + offset
	return firstThursday + 21
}

// GetHolidays 检测给定日期在指定国家代码（如 "CN", "US", "JP" 等）下的全部当日节日与网络纪念日。
func GetHolidays(t time.Time, countryCode string) []Holiday {
	var list []Holiday
	year := t.Year()
	month := int(t.Month())
	day := t.Day()
	cc := strings.ToUpper(strings.TrimSpace(countryCode))

	// 1. 浮动节日检测：感恩节、黑色星期五、网络星期一、母亲节、父亲节、国际程序员日
	tgDay := thanksgivingDay(year)
	tgDate := time.Date(year, time.November, tgDay, 0, 0, 0, 0, time.UTC)
	bfDate := tgDate.AddDate(0, 0, 1) // 感恩节次日 = 黑色星期五 (Black Friday)
	cmDate := tgDate.AddDate(0, 0, 4) // 感恩节下周一 = 网络星期一 (Cyber Monday)

	// 黑色星期五打折季 (Black Friday)
	if month == int(bfDate.Month()) && day == bfDate.Day() {
		list = append(list, Holiday{
			NameZH: "黑色星期五打折狂欢节 (Black Friday)",
			NameEN: "Black Friday Sale",
			NameJA: "ブラックフライデー セール",
			Kind:   KindShopping,
			Icon:   "🛍️",
		})
	}

	// 网络星期一大促 (Cyber Monday)
	if month == int(cmDate.Month()) && day == cmDate.Day() {
		list = append(list, Holiday{
			NameZH: "网络星期一大促 (Cyber Monday)",
			NameEN: "Cyber Monday Tech Deals",
			NameJA: "サイバーマンデー 大セール",
			Kind:   KindShopping,
			Icon:   "🛒",
		})
	}

	// 美国感恩节 (Thanksgiving)
	if month == 11 && day == tgDay {
		list = append(list, Holiday{
			NameZH: "感恩节 (Thanksgiving)",
			NameEN: "Thanksgiving Day",
			NameJA: "感謝祭 (Thanksgiving)",
			Kind:   KindTraditional,
			Icon:   "🦃",
		})
	}

	// 国际程序员日（每年的第 256 天，2^8）
	if t.YearDay() == 256 {
		list = append(list, Holiday{
			NameZH: "国际程序员日 (第256天)",
			NameEN: "International Programmers' Day",
			NameJA: "国際プログラマーの日",
			Kind:   KindInternet,
			Icon:   "👨‍💻",
		})
	}

	// 母亲节（5月第二个星期日）
	if month == 5 && t.Weekday() == time.Sunday && (day-1)/7 == 1 {
		list = append(list, Holiday{
			NameZH: "母亲节",
			NameEN: "Mother's Day",
			NameJA: "母の日",
			Kind:   KindTraditional,
			Icon:   "💐",
		})
	}

	// 父亲节（6月第三个星期日）
	if month == 6 && t.Weekday() == time.Sunday && (day-1)/7 == 2 {
		list = append(list, Holiday{
			NameZH: "父亲节",
			NameEN: "Father's Day",
			NameJA: "父の日",
			Kind:   KindTraditional,
			Icon:   "👔",
		})
	}

	// 2. 现代科技、网络特色节日与电商大促（用户重点需求：1024程序员节、双11、618、520等）
	switch {
	case month == 10 && day == 24:
		list = append(list, Holiday{
			NameZH: "1024 程序员节",
			NameEN: "1024 Programmers' Day",
			NameJA: "1024 プログラマーの日",
			Kind:   KindInternet,
			Icon:   "💻",
		})
	case month == 11 && day == 11:
		list = append(list, Holiday{
			NameZH: "双11全球购物狂欢节 / 光棍节",
			NameEN: "Double 11 Global Shopping Gala / Singles' Day",
			NameJA: "ダブル11 ショッピング祭り / 独身の日",
			Kind:   KindShopping,
			Icon:   "🛍️",
		})
	case month == 12 && day == 12:
		list = append(list, Holiday{
			NameZH: "双12年终购物盛典",
			NameEN: "Double 12 Year-End Sale",
			NameJA: "ダブル12 年末セール",
			Kind:   KindShopping,
			Icon:   "🎁",
		})
	case month == 6 && day == 18:
		list = append(list, Holiday{
			NameZH: "618 年中狂欢大促",
			NameEN: "618 Mid-Year Shopping Gala",
			NameJA: "618 年中大感謝祭",
			Kind:   KindShopping,
			Icon:   "🔥",
		})
	case month == 5 && day == 20:
		list = append(list, Holiday{
			NameZH: "520 网络情人节 (表白日)",
			NameEN: "520 Internet Valentine's Day",
			NameJA: "520 インターネット告白デー",
			Kind:   KindInternet,
			Icon:   "💖",
		})
	case month == 5 && day == 21:
		list = append(list, Holiday{
			NameZH: "521 甜蜜告白日",
			NameEN: "521 Confession Day",
			NameJA: "521 告白の日",
			Kind:   KindInternet,
			Icon:   "💌",
		})
	case month == 3 && day == 14:
		list = append(list, Holiday{
			NameZH: "圆周率日 (π Day) • 白色情人节",
			NameEN: "Pi Day (π Day) & White Day",
			NameJA: "円周率の日 (π Day)・ホワイトデー",
			Kind:   KindInternet,
			Icon:   "🥧",
		})
	case month == 5 && day == 4:
		list = append(list, Holiday{
			NameZH: "星球大战日 (May the 4th)",
			NameEN: "Star Wars Day (\"May the 4th Be With You\")",
			NameJA: "スター・ウォーズの日",
			Kind:   KindInternet,
			Icon:   "⚔️",
		})
	case month == 4 && day == 1:
		list = append(list, Holiday{
			NameZH: "愚人节",
			NameEN: "April Fools' Day",
			NameJA: "エイプリルフール",
			Kind:   KindInternet,
			Icon:   "🃏",
		})
	case month == 7 && day == 17:
		list = append(list, Holiday{
			NameZH: "世界表情符号日 (World Emoji Day)",
			NameEN: "World Emoji Day",
			NameJA: "世界絵文字デー",
			Kind:   KindInternet,
			Icon:   "🥳",
		})
	case month == 9 && day == 9:
		list = append(list, Holiday{
			NameZH: "99 公益日 • 拒绝加班倡议日",
			NameEN: "99 Charity Day",
			NameJA: "99 チャリティーデー",
			Kind:   KindInternet,
			Icon:   "🌱",
		})
	case month == 8 && day == 18:
		list = append(list, Holiday{
			NameZH: "818 电商发烧狂欢节",
			NameEN: "818 Tech Shopping Festival",
			NameJA: "818 ショッピングフェス",
			Kind:   KindShopping,
			Icon:   "⚡",
		})
	case month == 8 && day == 8:
		list = append(list, Holiday{
			NameZH: "全民健身日 • 爸爸节",
			NameEN: "National Fitness Day",
			NameJA: "全民フィットネスデー",
			Kind:   KindInternet,
			Icon:   "🏃",
		})
	}

	// 3. 中国二十四节气识别（支持 21 世纪全部年份天文匹配）
	for _, st := range solarTermCoeffs {
		if st.month == month {
			termDay := computeSolarTermDay(year, st.month, st.c)
			if day == termDay {
				list = append(list, Holiday{
					NameZH: "节气：" + st.zh,
					NameEN: "Solar Term: " + st.en,
					NameJA: "節気：" + st.ja,
					Kind:   KindSolarTerm,
					Icon:   st.icon,
				})
				break
			}
		}
	}

	// 4. 农历传统节日匹配（通过高精度 lunar.go 离线换算）
	lunar := ConvertToLunar(t)
	if !lunar.IsLeap {
		switch {
		case lunar.Month == 1 && lunar.Day == 1:
			list = append(list, Holiday{
				NameZH: "春节 (农历正月初一)",
				NameEN: "Chinese New Year (Spring Festival)",
				NameJA: "春節 (旧正月)",
				Kind:   KindTraditional,
				Icon:   "🧧",
			})
		case lunar.Month == 1 && lunar.Day == 15:
			list = append(list, Holiday{
				NameZH: "元宵节 (上元灯会)",
				NameEN: "Lantern Festival",
				NameJA: "元宵節 (ランタン祭り)",
				Kind:   KindTraditional,
				Icon:   "🏮",
			})
		case lunar.Month == 2 && lunar.Day == 2:
			list = append(list, Holiday{
				NameZH: "龙抬头 (春耕节)",
				NameEN: "Dragon Head-Raising Day",
				NameJA: "竜抬頭 (春耕祭り)",
				Kind:   KindTraditional,
				Icon:   "🐉",
			})
		case lunar.Month == 5 && lunar.Day == 5:
			list = append(list, Holiday{
				NameZH: "端午节 (粽香龙舟)",
				NameEN: "Dragon Boat Festival",
				NameJA: "端午節 (ちまき・ドラゴンボート)",
				Kind:   KindTraditional,
				Icon:   "🛶",
			})
		case lunar.Month == 7 && lunar.Day == 7:
			list = append(list, Holiday{
				NameZH: "七夕节 (传统中国情人节)",
				NameEN: "Qixi Festival (Chinese Valentine's Day)",
				NameJA: "七夕祭り",
				Kind:   KindTraditional,
				Icon:   "🎋",
			})
		case lunar.Month == 7 && lunar.Day == 15:
			list = append(list, Holiday{
				NameZH: "中元节 (盂兰盆节)",
				NameEN: "Ghost Festival (Zhongyuan)",
				NameJA: "中元節 (お盆)",
				Kind:   KindTraditional,
				Icon:   "🪔",
			})
		case lunar.Month == 8 && lunar.Day == 15:
			list = append(list, Holiday{
				NameZH: "中秋节 (月圆人团圆)",
				NameEN: "Mid-Autumn Festival",
				NameJA: "中秋節 (月見・お月見)",
				Kind:   KindTraditional,
				Icon:   "🥮",
			})
		case lunar.Month == 9 && lunar.Day == 9:
			list = append(list, Holiday{
				NameZH: "重阳节 (登高祈福)",
				NameEN: "Double Ninth Festival",
				NameJA: "重陽の節句",
				Kind:   KindTraditional,
				Icon:   "⛰️",
			})
		case lunar.Month == 12 && lunar.Day == 8:
			list = append(list, Holiday{
				NameZH: "腊八节 (喝腊八粥)",
				NameEN: "Laba Rice Porridge Festival",
				NameJA: "臘八節 (ろうはちせつ)",
				Kind:   KindTraditional,
				Icon:   "🥣",
			})
		case lunar.Month == 12 && (lunar.Day == 23 || lunar.Day == 24):
			list = append(list, Holiday{
				NameZH: "小年 (祭灶迎春)",
				NameEN: "Little New Year (Kitchen God Day)",
				NameJA: "小年 (かまどの神の日)",
				Kind:   KindTraditional,
				Icon:   "🧹",
			})
		}

		// 检查明日是否为正月初一，是则今日为除夕（大年三十/廿九）
		tomorrow := t.AddDate(0, 0, 1)
		lunarTom := ConvertToLunar(tomorrow)
		if lunarTom.Month == 1 && lunarTom.Day == 1 {
			list = append(list, Holiday{
				NameZH: "除夕 (岁除守岁 • 大年夜)",
				NameEN: "Chinese New Year's Eve",
				NameJA: "大晦日 (旧正月イブ)",
				Kind:   KindTraditional,
				Icon:   "🎆",
			})
		}
	}

	// 5. 全球与国际公认节日（元旦、情人节、妇女节、劳动节、儿童节、圣诞等）
	switch {
	case month == 1 && day == 1:
		list = append(list, Holiday{
			NameZH: "元旦新年",
			NameEN: "New Year's Day",
			NameJA: "元日 (お正月)",
			Kind:   KindTraditional,
			Icon:   "🎆",
		})
	case month == 2 && day == 14:
		list = append(list, Holiday{
			NameZH: "西方情人节",
			NameEN: "Valentine's Day",
			NameJA: "バレンタインデー",
			Kind:   KindTraditional,
			Icon:   "🌹",
		})
	case month == 3 && day == 8:
		list = append(list, Holiday{
			NameZH: "国际劳动妇女节",
			NameEN: "International Women's Day",
			NameJA: "国際女性デー",
			Kind:   KindInternational,
			Icon:   "🌸",
		})
	case month == 4 && day == 22:
		list = append(list, Holiday{
			NameZH: "世界地球日",
			NameEN: "Earth Day",
			NameJA: "アースデイ (地球の日)",
			Kind:   KindInternational,
			Icon:   "🌍",
		})
	case month == 4 && day == 23:
		list = append(list, Holiday{
			NameZH: "世界读书日",
			NameEN: "World Book Day",
			NameJA: "世界図書の日",
			Kind:   KindInternational,
			Icon:   "📖",
		})
	case month == 5 && day == 1:
		list = append(list, Holiday{
			NameZH: "国际劳动节",
			NameEN: "International Workers' Day",
			NameJA: "メーデー (労働者の日)",
			Kind:   KindTraditional,
			Icon:   "🛠️",
		})
	case month == 6 && day == 1:
		list = append(list, Holiday{
			NameZH: "国际儿童节",
			NameEN: "International Children's Day",
			NameJA: "国際こどもの日",
			Kind:   KindInternational,
			Icon:   "🎈",
		})
	case month == 9 && day == 21:
		list = append(list, Holiday{
			NameZH: "国际和平日",
			NameEN: "International Day of Peace",
			NameJA: "国際平和デー",
			Kind:   KindInternational,
			Icon:   "🕊️",
		})
	case month == 10 && day == 24:
		list = append(list, Holiday{
			NameZH: "联合国日",
			NameEN: "United Nations Day",
			NameJA: "国際連合デー",
			Kind:   KindInternational,
			Icon:   "🇺🇳",
		})
	case month == 10 && day == 31:
		list = append(list, Holiday{
			NameZH: "万圣夜 (Halloween)",
			NameEN: "Halloween",
			NameJA: "ハロウィン (Halloween)",
			Kind:   KindTraditional,
			Icon:   "🎃",
		})
	case month == 11 && day == 1:
		list = append(list, Holiday{
			NameZH: "万圣节 (诸圣瞻礼)",
			NameEN: "All Saints' Day",
			NameJA: "万聖節",
			Kind:   KindTraditional,
			Icon:   "🕯️",
		})
	case month == 12 && day == 24:
		list = append(list, Holiday{
			NameZH: "平安夜 (Christmas Eve)",
			NameEN: "Christmas Eve",
			NameJA: "クリスマスイブ",
			Kind:   KindTraditional,
			Icon:   "🎄",
		})
	case month == 12 && day == 25:
		list = append(list, Holiday{
			NameZH: "圣诞节 (Christmas)",
			NameEN: "Christmas Day",
			NameJA: "クリスマス",
			Kind:   KindTraditional,
			Icon:   "🎅",
		})
	case month == 12 && day == 31:
		list = append(list, Holiday{
			NameZH: "跨年夜 (岁末迎新)",
			NameEN: "New Year's Eve",
			NameJA: "大晦日 (New Year's Eve)",
			Kind:   KindTraditional,
			Icon:   "🥂",
		})
	}

	// 6. 各国专属国庆与法定特色节日（当定位匹配对应国家时呈现）
	if cc == "CN" || cc == "CHINA" || cc == "" {
		switch {
		case month == 5 && day == 4:
			list = append(list, Holiday{
				NameZH: "五四青年节",
				NameEN: "Youth Day",
				NameJA: "五四青年節",
				Kind:   KindTraditional,
				Icon:   "⭐",
			})
		case month == 7 && day == 1:
			list = append(list, Holiday{
				NameZH: "建党纪念日 • 香港回归纪念日",
				NameEN: "CPC Founding Day & HK Return Day",
				NameJA: "建党記念日・香港復帰記念日",
				Kind:   KindTraditional,
				Icon:   "⭐",
			})
		case month == 8 && day == 1:
			list = append(list, Holiday{
				NameZH: "建军节 (八一建军节)",
				NameEN: "Army Day",
				NameJA: "建軍節",
				Kind:   KindTraditional,
				Icon:   "🎖️",
			})
		case month == 9 && day == 10:
			list = append(list, Holiday{
				NameZH: "中国教师节",
				NameEN: "Teachers' Day",
				NameJA: "教師の日",
				Kind:   KindTraditional,
				Icon:   "🧑‍🏫",
			})
		case month == 10 && day == 1:
			list = append(list, Holiday{
				NameZH: "中华人民共和国国庆节",
				NameEN: "National Day of China",
				NameJA: "国慶節 (中国建国記念日)",
				Kind:   KindTraditional,
				Icon:   "🇨🇳",
			})
		case month == 11 && day == 8:
			list = append(list, Holiday{
				NameZH: "中国记者节",
				NameEN: "Journalists' Day",
				NameJA: "記者節",
				Kind:   KindTraditional,
				Icon:   "📰",
			})
		}
	}

	if cc == "US" || cc == "USA" {
		switch {
		case month == 7 && day == 4:
			list = append(list, Holiday{
				NameZH: "美国独立日",
				NameEN: "Independence Day (4th of July)",
				NameJA: "独立記念日",
				Kind:   KindTraditional,
				Icon:   "🇺🇸",
			})
		case month == 11 && day == 11:
			list = append(list, Holiday{
				NameZH: "美国退伍军人节",
				NameEN: "Veterans Day",
				NameJA: "退役軍人の日",
				Kind:   KindTraditional,
				Icon:   "🎖️",
			})
		}
	}

	if cc == "JP" || cc == "JAPAN" {
		switch {
		case month == 2 && day == 11:
			list = append(list, Holiday{
				NameZH: "日本建国纪念日",
				NameEN: "National Foundation Day",
				NameJA: "建国記念の日",
				Kind:   KindTraditional,
				Icon:   "🎌",
			})
		case month == 2 && day == 23:
			list = append(list, Holiday{
				NameZH: "日本天皇诞生日",
				NameEN: "Emperor's Birthday",
				NameJA: "天皇誕生日",
				Kind:   KindTraditional,
				Icon:   "🎌",
			})
		case month == 4 && day == 29:
			list = append(list, Holiday{
				NameZH: "昭和之日",
				NameEN: "Showa Day",
				NameJA: "昭和の日",
				Kind:   KindTraditional,
				Icon:   "🎌",
			})
		case month == 5 && day == 3:
			list = append(list, Holiday{
				NameZH: "宪法纪念日",
				NameEN: "Constitution Memorial Day",
				NameJA: "憲法記念日",
				Kind:   KindTraditional,
				Icon:   "🎌",
			})
		case month == 5 && day == 5:
			list = append(list, Holiday{
				NameZH: "日本儿童节",
				NameEN: "Children's Day",
				NameJA: "こどもの日",
				Kind:   KindTraditional,
				Icon:   "🎏",
			})
		case month == 8 && day == 11:
			list = append(list, Holiday{
				NameZH: "山之日",
				NameEN: "Mountain Day",
				NameJA: "山の日",
				Kind:   KindTraditional,
				Icon:   "⛰️",
			})
		case month == 11 && day == 3:
			list = append(list, Holiday{
				NameZH: "日本文化日",
				NameEN: "Culture Day",
				NameJA: "文化の日",
				Kind:   KindTraditional,
				Icon:   "🎌",
			})
		case month == 11 && day == 23:
			list = append(list, Holiday{
				NameZH: "勤劳感谢日",
				NameEN: "Labor Thanksgiving Day",
				NameJA: "勤労感謝の日",
				Kind:   KindTraditional,
				Icon:   "🎌",
			})
		}
	}

	if cc == "GB" || cc == "UK" {
		if month == 12 && day == 26 {
			list = append(list, Holiday{
				NameZH: "英国节礼日",
				NameEN: "Boxing Day",
				NameJA: "ボクシング・デー",
				Kind:   KindTraditional,
				Icon:   "🇬🇧",
			})
		}
	}

	if cc == "FR" {
		if month == 7 && day == 14 {
			list = append(list, Holiday{
				NameZH: "法国国庆日 (巴士底日)",
				NameEN: "Bastille Day (French National Day)",
				NameJA: "パリ祭 (フランス革命記念日)",
				Kind:   KindTraditional,
				Icon:   "🇫🇷",
			})
		}
	}

	return list
}

// FormatHolidayBadge 返回用于主时钟与悬浮时钟展示的今日节日完整徽章文本。若当日无节日则明确标注“无”。
func FormatHolidayBadge(holidays []Holiday, lang string) string {
	if len(holidays) == 0 {
		switch lang {
		case "ja":
			return "🎉 今日の祝日・記念日：なし"
		case "en":
			return "🎉 Today's Holiday: None"
		default:
			return "🎉 今日节日：无"
		}
	}

	var parts []string
	prefix := "🎉 今日节日："
	switch lang {
	case "ja":
		prefix = "🎉 今日の祝日・記念日："
	case "en":
		prefix = "🎉 Today's Holiday: "
	}

	for _, h := range holidays {
		parts = append(parts, fmt.Sprintf("%s %s", h.Icon, h.LocalizedName(lang)))
	}

	return prefix + strings.Join(parts, " • ")
}

// FormatShortHoliday 返回用于世界时钟城市卡片展示的简明节日标签。
func FormatShortHoliday(holidays []Holiday, lang string) string {
	if len(holidays) == 0 {
		return ""
	}
	h := holidays[0]
	return fmt.Sprintf("%s %s", h.Icon, h.LocalizedName(lang))
}
