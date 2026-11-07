// Package clock 提供公历（阳历）与中国传统农历（阴历、天干地支、生肖）的转换算法。
package clock

import (
	"fmt"
	"time"
)

// LunarDate 结构体保存农历年、月、日、闰月标识、天干地支纪年与生肖信息。
type LunarDate struct {
	Year        int    // 农历年份（如 2026）
	Month       int    // 农历月份（1 ~ 12）
	Day         int    // 农历日期（1 ~ 30）
	IsLeap      bool   // 当前月是否为闰月
	GanZhiYear  string // 天干地支纪年（如 "丙午年"）
	Zodiac      string // 十二生肖（如 "马年"）
	MonthNameZH string // 农历月份中文名称（如 "九月"、"闰六月"）
	DayNameZH   string // 农历日期中文名称（如 "廿九"、"初一"）
}

// FormatZH 输出标准中文农历格式，如 "农历 丙午年(马年) 九月廿九"。
func (l LunarDate) FormatZH() string {
	return fmt.Sprintf("农历 %s(%s) %s%s", l.GanZhiYear, l.Zodiac, l.MonthNameZH, l.DayNameZH)
}

// FormatEN 输出英文本地化农历描述。
func (l LunarDate) FormatEN() string {
	leapStr := ""
	if l.IsLeap {
		leapStr = "Leap "
	}
	return fmt.Sprintf("Lunar: Year of the %s (%s) • %sMonth %d, Day %d", l.ZodiacEN(), l.GanZhiYear, leapStr, l.Month, l.Day)
}

// FormatJA 输出日文本地化旧历描述。
func (l LunarDate) FormatJA() string {
	leapStr := ""
	if l.IsLeap {
		leapStr = "閏"
	}
	return fmt.Sprintf("旧暦：%s(%s) %s%d月%d日", l.GanZhiYear, l.ZodiacJA(), leapStr, l.Month, l.Day)
}

// ZodiacEN 返回当前农历生肖动物的英文名称（如 "Horse", "Dragon"）。
func (l LunarDate) ZodiacEN() string {
	idx := (l.Year - 4) % 12
	if idx < 0 {
		idx += 12
	}
	enAnimals := []string{"Rat", "Ox", "Tiger", "Rabbit", "Dragon", "Snake", "Horse", "Goat", "Monkey", "Rooster", "Dog", "Pig"}
	return enAnimals[idx]
}

// ZodiacJA 返回当前农历生肖动物的日文名称（如 "午・うま"）。
func (l LunarDate) ZodiacJA() string {
	idx := (l.Year - 4) % 12
	if idx < 0 {
		idx += 12
	}
	jaAnimals := []string{"子・ねずみ", "丑・うし", "寅・とら", "卯・うさぎ", "辰・たつ", "巳・へび", "午・うま", "未・ひつじ", "申・さる", "酉・とり", "戌・いぬ", "亥・いのしし"}
	return jaAnimals[idx]
}

// lunarInfo 保存 1900 年至 2050 年的农历天文历法数据（20位十六进制编码）。
// 编码规则：
// - 低 4 位 (data & 0x0F)：当年闰月月份（0 表示无闰月，1~12 表示闰几月）
// - 中间 12 位 ((data >> 4) & 0xFFF)：12 个正常月份大小（1 为 30 天，0 为 29 天）
// - 高 4 位 ((data >> 16) & 0x01)：闰月大小（1 为 30 天，0 为 29 天）
var lunarInfo = []int{
	0x04bd8, 0x04ae0, 0x0a570, 0x054d5, 0x0d260, 0x0d950, 0x16554, 0x056a0, 0x09ad0, 0x055d2, // 1900-1909
	0x04ae0, 0x0a5b6, 0x0a4d0, 0x0d250, 0x1d255, 0x0b540, 0x0d6a0, 0x0ada2, 0x095b0, 0x14977, // 1910-1919
	0x04970, 0x0a4b0, 0x0b4b5, 0x06a50, 0x06d40, 0x1ab54, 0x02b60, 0x09570, 0x052f2, 0x04970, // 1920-1929
	0x06566, 0x0d4a0, 0x0ea50, 0x06e95, 0x05ad0, 0x02b60, 0x186e3, 0x092e0, 0x1c8d7, 0x0c950, // 1930-1939
	0x0d4a0, 0x1d8a6, 0x0b550, 0x056a0, 0x1a5b4, 0x025d0, 0x092d0, 0x0d2b2, 0x0a950, 0x0b557, // 1940-1949
	0x06ca0, 0x0b550, 0x15355, 0x04da0, 0x0a5d0, 0x14573, 0x052d0, 0x0a9a8, 0x0e950, 0x06aa0, // 1950-1959
	0x0aea6, 0x0ab50, 0x04b60, 0x0aae4, 0x0a570, 0x05260, 0x0f263, 0x0d950, 0x05b57, 0x056a0, // 1960-1969
	0x096d0, 0x04dd5, 0x04ad0, 0x0a4d0, 0x0d4d4, 0x0d250, 0x0d558, 0x0b540, 0x0b5a0, 0x195a6, // 1970-1979
	0x095b0, 0x049b0, 0x0a974, 0x0a4b0, 0x0b27a, 0x06a50, 0x06d40, 0x0af46, 0x0ab60, 0x09570, // 1980-1989
	0x04af5, 0x04970, 0x064b0, 0x074a3, 0x0ea50, 0x06b58, 0x055c0, 0x0ab60, 0x096d5, 0x092e0, // 1990-1999
	0x0c960, 0x0d954, 0x0d4a0, 0x0da50, 0x07552, 0x056a0, 0x0abb7, 0x025d0, 0x092d0, 0x0cab5, // 2000-2009
	0x0a950, 0x0b4a0, 0x0baa4, 0x0ad50, 0x055d9, 0x04ba0, 0x0a5b0, 0x15176, 0x052b0, 0x0a930, // 2010-2019
	0x07954, 0x06aa0, 0x0ad50, 0x05b52, 0x04b60, 0x0a6e6, 0x0a4e0, 0x0d260, 0x0ea65, 0x0d530, // 2020-2029
	0x05aa0, 0x076a3, 0x096d0, 0x04bd7, 0x04ad0, 0x0a4d0, 0x1d0b6, 0x0d250, 0x0d520, 0x0dd45, // 2030-2039
	0x0b5a0, 0x056d0, 0x055b2, 0x049b0, 0x0a577, 0x0a4b0, 0x0aa50, 0x1b255, 0x06d20, 0x0ada0, // 2040-2049
	0x14b63, // 2050
}

var (
	ganNames   = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	zhiNames   = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	shengXiao  = []string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}
	monthNames = []string{"正月", "二月", "三月", "四月", "五月", "六月", "七月", "八月", "九月", "十月", "冬月", "腊月"}
	dayNames   = []string{
		"初一", "初二", "初三", "初四", "初五", "初六", "初七", "初八", "初九", "初十",
		"十一", "十二", "十三", "十四", "十五", "十六", "十七", "十八", "十九", "二十",
		"廿一", "廿二", "廿三", "廿四", "廿五", "廿六", "廿七", "廿八", "廿九", "三十",
	}
)

func lunarYearDays(y int) int {
	idx := y - 1900
	if idx < 0 || idx >= len(lunarInfo) {
		return 365
	}
	info := lunarInfo[idx]
	i := 0x8000
	total := 348
	for i > 0x8 {
		if (info & i) != 0 {
			total++
		}
		i >>= 1
	}
	leapM := info & 0x0F
	if leapM > 0 {
		if (info & 0x10000) != 0 {
			total += 30
		} else {
			total += 29
		}
	}
	return total
}

// ConvertToLunar 将公历时间 t 转换为对应的中国农历日期结构体。
func ConvertToLunar(t time.Time) LunarDate {
	// 公历 1900 年 1 月 31 日为农历庚子年正月初一
	base := time.Date(1900, 1, 31, 0, 0, 0, 0, t.Location())
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	offsetDays := int(target.Sub(base).Hours() / 24)
	if offsetDays < 0 {
		offsetDays = 0
	}

	y := 1900
	maxYear := 1900 + len(lunarInfo) - 1
	for y < maxYear && offsetDays > 0 {
		days := lunarYearDays(y)
		if offsetDays < days {
			break
		}
		offsetDays -= days
		y++
	}

	info := 0x04bd8
	if y-1900 < len(lunarInfo) {
		info = lunarInfo[y-1900]
	}

	leapM := info & 0x0F
	isLeap := false
	m := 1
	for m <= 12 && offsetDays > 0 {
		days := 29
		if leapM > 0 && m == (leapM+1) && !isLeap {
			isLeap = true
			m--
			if (info & 0x10000) != 0 {
				days = 30
			}
		} else {
			if (info & (0x8000 >> (m - 1))) != 0 {
				days = 30
			}
		}
		if offsetDays < days {
			break
		}
		offsetDays -= days
		if isLeap {
			isLeap = false
		}
		m++
	}

	d := offsetDays + 1
	if d > 30 {
		d = 30
	}

	// 计算天干地支与生肖（以农历年计算）
	ganIdx := (y - 4) % 10
	if ganIdx < 0 {
		ganIdx += 10
	}
	zhiIdx := (y - 4) % 12
	if zhiIdx < 0 {
		zhiIdx += 12
	}

	gzYear := ganNames[ganIdx] + zhiNames[zhiIdx] + "年"
	zodiac := shengXiao[zhiIdx] + "年"

	mName := "正月"
	if m >= 1 && m <= 12 {
		mName = monthNames[m-1]
	}
	if isLeap {
		mName = "闰" + mName
	}

	dName := "初一"
	if d >= 1 && d <= 30 {
		dName = dayNames[d-1]
	}

	return LunarDate{
		Year:        y,
		Month:       m,
		Day:         d,
		IsLeap:      isLeap,
		GanZhiYear:  gzYear,
		Zodiac:      zodiac,
		MonthNameZH: mName,
		DayNameZH:   dName,
	}
}

// FormatDualCalendar 返回指定语言下的公历（阳历）与农历（阴历）格式化字符串。
func FormatDualCalendar(t time.Time, lang string) (solarStr string, lunarStr string) {
	lunar := ConvertToLunar(t)

	switch lang {
	case "zh":
		weekdayCN := []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}[int(t.Weekday())%7]
		solarStr = fmt.Sprintf("📅 公历：%d年%02d月%02d日 %s", t.Year(), int(t.Month()), t.Day(), weekdayCN)
		lunarStr = fmt.Sprintf("🌙 %s", lunar.FormatZH())
	case "ja":
		weekdayJA := []string{"日", "月", "火", "水", "木", "金", "土"}[int(t.Weekday())%7]
		solarStr = fmt.Sprintf("📅 太陽暦：%d年%02d月%02d日 (%s曜日)", t.Year(), int(t.Month()), t.Day(), weekdayJA)
		lunarStr = fmt.Sprintf("🌙 %s", lunar.FormatJA())
	default:
		solarStr = fmt.Sprintf("📅 Solar: %s, %s %d, %d", t.Weekday().String(), t.Month().String(), t.Day(), t.Year())
		lunarStr = fmt.Sprintf("🌙 %s", lunar.FormatEN())
	}
	return solarStr, lunarStr
}
