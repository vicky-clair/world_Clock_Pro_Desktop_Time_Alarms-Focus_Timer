package clock

import (
	"strings"
	"testing"
	"time"
)

func TestConvertToLunar(t *testing.T) {
	// 2026-11-07 corresponds to lunar year 2026 (丙午年 马年), 9th month, 29th day
	refDate := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	lunar := ConvertToLunar(refDate)

	if lunar.Year != 2026 {
		t.Fatalf("expected year 2026, got %d", lunar.Year)
	}
	if lunar.Month != 9 {
		t.Fatalf("expected month 9, got %d", lunar.Month)
	}
	if lunar.Day != 29 {
		t.Fatalf("expected day 29, got %d", lunar.Day)
	}
	if lunar.GanZhiYear != "丙午年" {
		t.Fatalf("expected 丙午年, got %s", lunar.GanZhiYear)
	}
	if lunar.Zodiac != "马年" {
		t.Fatalf("expected 马年, got %s", lunar.Zodiac)
	}
	if lunar.MonthNameZH != "九月" {
		t.Fatalf("expected 九月, got %s", lunar.MonthNameZH)
	}
	if lunar.DayNameZH != "廿九" {
		t.Fatalf("expected 廿九, got %s", lunar.DayNameZH)
	}
}

func TestFormatDualCalendar(t *testing.T) {
	refDate := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)

	// Chinese
	solarZH, lunarZH := FormatDualCalendar(refDate, "zh")
	if !strings.Contains(solarZH, "公历") || !strings.Contains(solarZH, "2026年11月07日") {
		t.Fatalf("unexpected solarZH: %s", solarZH)
	}
	if !strings.Contains(lunarZH, "农历") || !strings.Contains(lunarZH, "丙午年(马年) 九月廿九") {
		t.Fatalf("unexpected lunarZH: %s", lunarZH)
	}

	// English
	solarEN, lunarEN := FormatDualCalendar(refDate, "en")
	if !strings.Contains(solarEN, "Solar") || !strings.Contains(solarEN, "2026") {
		t.Fatalf("unexpected solarEN: %s", solarEN)
	}
	if !strings.Contains(lunarEN, "Lunar") || !strings.Contains(lunarEN, "Horse") {
		t.Fatalf("unexpected lunarEN: %s", lunarEN)
	}

	// Japanese
	solarJA, lunarJA := FormatDualCalendar(refDate, "ja")
	if !strings.Contains(solarJA, "太陽暦") {
		t.Fatalf("unexpected solarJA: %s", solarJA)
	}
	if !strings.Contains(lunarJA, "旧暦") {
		t.Fatalf("unexpected lunarJA: %s", lunarJA)
	}
}
