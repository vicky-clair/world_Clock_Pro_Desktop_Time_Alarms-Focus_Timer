package clock

import (
	"strings"
	"testing"
	"time"
)

func TestGetHolidays(t *testing.T) {
	// 1. Test 1024 Programmers' Day (Oct 24, 2026)
	t1024 := time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC)
	h1024 := GetHolidays(t1024, "CN")
	found1024 := false
	for _, h := range h1024 {
		if strings.Contains(h.NameZH, "1024 程序员节") {
			found1024 = true
			break
		}
	}
	if !found1024 {
		t.Fatal("expected 1024 Programmers' Day on Oct 24")
	}

	// 2. Test Black Friday (Nov 27, 2026)
	tBF := time.Date(2026, 11, 27, 10, 0, 0, 0, time.UTC)
	hBF := GetHolidays(tBF, "US")
	foundBF := false
	for _, h := range hBF {
		if strings.Contains(h.NameZH, "黑色星期五") {
			foundBF = true
			break
		}
	}
	if !foundBF {
		t.Fatal("expected Black Friday on Nov 27, 2026")
	}

	// 3. Test Cyber Monday (Nov 30, 2026)
	tCM := time.Date(2026, 11, 30, 10, 0, 0, 0, time.UTC)
	hCM := GetHolidays(tCM, "US")
	foundCM := false
	for _, h := range hCM {
		if strings.Contains(h.NameZH, "网络星期一") {
			foundCM = true
			break
		}
	}
	if !foundCM {
		t.Fatal("expected Cyber Monday on Nov 30, 2026")
	}

	// 4. Test Double 11 / Singles' Day (Nov 11)
	t1111 := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	h1111 := GetHolidays(t1111, "CN")
	found1111 := false
	for _, h := range h1111 {
		if strings.Contains(h.NameZH, "双11") {
			found1111 = true
			break
		}
	}
	if !found1111 {
		t.Fatal("expected Double 11 Shopping Festival on Nov 11")
	}

	// 5. Test Start of Winter (立冬) on Nov 7, 2026
	tToday := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	hToday := GetHolidays(tToday, "CN")
	foundLiDong := false
	for _, h := range hToday {
		if strings.Contains(h.NameZH, "立冬") {
			foundLiDong = true
			break
		}
	}
	if !foundLiDong {
		t.Fatal("expected Solar Term 立冬 on Nov 7, 2026")
	}

	// 6. Test FormatHolidayBadge and FormatShortHoliday
	badgeZH := FormatHolidayBadge(hToday, "zh")
	if !strings.Contains(badgeZH, "今日节日") || !strings.Contains(badgeZH, "立冬") {
		t.Fatalf("unexpected badgeZH: %s", badgeZH)
	}

	badgeEN := FormatHolidayBadge(hToday, "en")
	if !strings.Contains(badgeEN, "Today's Holiday") || !strings.Contains(badgeEN, "Start of Winter") {
		t.Fatalf("unexpected badgeEN: %s", badgeEN)
	}

	badgeJA := FormatHolidayBadge(hToday, "ja")
	if !strings.Contains(badgeJA, "祝日") || !strings.Contains(badgeJA, "立冬") {
		t.Fatalf("unexpected badgeJA: %s", badgeJA)
	}

	short := FormatShortHoliday(hToday, "zh")
	if !strings.Contains(short, "立冬") {
		t.Fatalf("unexpected short holiday string: %s", short)
	}

	// 7. Test FormatHolidayBadge when there are no holidays (should display "无")
	emptyZH := FormatHolidayBadge(nil, "zh")
	if emptyZH != "🎉 今日节日：无" {
		t.Fatalf("expected '🎉 今日节日：无', got '%s'", emptyZH)
	}

	emptyEN := FormatHolidayBadge(nil, "en")
	if emptyEN != "🎉 Today's Holiday: None" {
		t.Fatalf("expected '🎉 Today's Holiday: None', got '%s'", emptyEN)
	}

	emptyJA := FormatHolidayBadge(nil, "ja")
	if emptyJA != "🎉 今日の祝日・記念日：なし" {
		t.Fatalf("expected '🎉 今日の祝日・記念日：なし', got '%s'", emptyJA)
	}
}
