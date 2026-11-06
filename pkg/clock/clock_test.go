package clock

import (
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	tm := time.Date(2026, 11, 5, 14, 5, 9, 0, time.UTC)

	t24s, p24s := FormatTime(tm, true, true)
	if t24s != "14:05:09" || p24s != "" {
		t.Errorf("expected 14:05:09, got %s, period=%s", t24s, p24s)
	}

	t24, p24 := FormatTime(tm, true, false)
	if t24 != "14:05" || p24 != "" {
		t.Errorf("expected 14:05, got %s, period=%s", t24, p24)
	}

	t12s, p12s := FormatTime(tm, false, true)
	if t12s != "02:05:09" || p12s != "PM" {
		t.Errorf("expected 02:05:09 PM, got %s %s", t12s, p12s)
	}

	t12, p12 := FormatTime(tm, false, false)
	if t12 != "02:05" || p12 != "PM" {
		t.Errorf("expected 02:05 PM, got %s %s", t12, p12)
	}

	// Midnight check
	tmMidnight := time.Date(2026, 11, 5, 0, 15, 0, 0, time.UTC)
	tMid, pMid := FormatTime(tmMidnight, false, false)
	if tMid != "12:15" || pMid != "AM" {
		t.Errorf("expected 12:15 AM, got %s %s", tMid, pMid)
	}
}

func TestFormatDate(t *testing.T) {
	tm := time.Date(2026, 11, 5, 14, 5, 9, 0, time.UTC)
	full := FormatDate(tm, true)
	short := FormatDate(tm, false)

	if full != "Thursday, November 5, 2026" {
		t.Errorf("unexpected full date: %s", full)
	}
	if short != "Thu, Nov 5, 2026" {
		t.Errorf("unexpected short date: %s", short)
	}
}

func TestNextBoundaryDelay(t *testing.T) {
	now := time.Date(2026, 11, 5, 14, 5, 9, 400000000, time.UTC)
	dSec := NextBoundaryDelay(now, true)
	if dSec <= 0 || dSec > time.Second {
		t.Errorf("unexpected delay for second: %v", dSec)
	}

	dMin := NextBoundaryDelay(now, false)
	if dMin <= 0 || dMin > time.Minute {
		t.Errorf("unexpected delay for minute: %v", dMin)
	}
}

func TestComputeAnalogGeometry(t *testing.T) {
	tm := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	geom := ComputeAnalogGeometry(100, 100, 80, tm)

	if geom.CenterX != 100 || geom.CenterY != 100 {
		t.Errorf("unexpected center: (%f, %f)", geom.CenterX, geom.CenterY)
	}

	// At 12:00:00, hands point straight up (Y < cy, X ≈ cx)
	if geom.HourHand.Y >= 100 || geom.MinuteHand.Y >= 100 || geom.SecondHand.Y >= 100 {
		t.Errorf("hands should point up at 12:00:00: hr=%v, min=%v, sec=%v", geom.HourHand, geom.MinuteHand, geom.SecondHand)
	}
}

func TestDetectSystemLocation(t *testing.T) {
	now := time.Now()
	loc := DetectSystemLocation(now)
	if loc.CountryCode == "" || loc.CountryZH == "" || loc.CountryEN == "" || loc.CountryJA == "" {
		t.Fatalf("expected non-empty country names, got %+v", loc)
	}
	if loc.UTCOffset == "" {
		t.Fatal("expected non-empty UTCOffset")
	}

	// Verify resolveLocation with explicit CN/JP/US inputs
	cn := resolveLocation(time.Date(2026, 11, 6, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), "CN", "zh-CN", "China Standard Time")
	if cn.CountryCode != "CN" || cn.CountryZH != "中国" || cn.CityZH != "北京" {
		t.Fatalf("unexpected CN location: %+v", cn)
	}
	jp := resolveLocation(time.Date(2026, 11, 6, 12, 0, 0, 0, time.FixedZone("JST", 9*3600)), "JP", "ja-JP", "Tokyo Standard Time")
	if jp.CountryCode != "JP" || jp.CountryJA != "日本" || jp.CityJA != "東京" {
		t.Fatalf("unexpected JP location: %+v", jp)
	}
}
