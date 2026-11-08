package iploc

import (
	"strings"
	"testing"
	"time"
)

func TestMaskIP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"45.129.228.101", "45.129.***.101"},
		{"192.168.1.1", "192.168.***.1"},
		{"10.0.0.1", "10.0.***.1"},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", "2001:0db8:****:7334"},
		{"invalid", "invalid"},
		{"", ""},
	}

	for _, tt := range tests {
		got := MaskIP(tt.input)
		if got != tt.expected {
			t.Errorf("MaskIP(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatBadge(t *testing.T) {
	info := Info{
		IP:          "45.129.228.101",
		Country:     "新加坡",
		CountryCode: "SG",
		City:        "新加坡",
		IsSuccess:   true,
	}

	badgeZH := info.FormatBadge("zh")
	if !strings.Contains(badgeZH, "新加坡") || !strings.Contains(badgeZH, "45.129.228.101") {
		t.Errorf("unexpected badgeZH: %s", badgeZH)
	}

	badgeEN := info.FormatBadge("en")
	if !strings.Contains(badgeEN, "IP Country") {
		t.Errorf("unexpected badgeEN: %s", badgeEN)
	}

	badgeJA := info.FormatBadge("ja")
	if !strings.Contains(badgeJA, "IP の国") {
		t.Errorf("unexpected badgeJA: %s", badgeJA)
	}

	// Unsuccessful info returns empty string
	infoFail := Info{IsSuccess: false}
	if infoFail.FormatBadge("zh") != "" {
		t.Errorf("expected empty badge for failed info")
	}
}

func TestServiceMockFetch(t *testing.T) {
	svc := NewService()
	svc.SetMockFetcher(func(lang string) (Info, error) {
		country := "中国"
		if lang == "en" {
			country = "China"
		}
		return Info{
			IP:          "120.244.11.22",
			Country:     country,
			CountryCode: "CN",
			City:        "上海",
			Latitude:    31.2304,
			Longitude:   121.4737,
			IsSuccess:   true,
		}, nil
	})

	var updated bool
	notifyChan := make(chan struct{})
	info := svc.EnsureFresh("zh", func() {
		updated = true
		close(notifyChan)
	})

	// First synchronous call returns empty before background completes
	if info.IsSuccess {
		t.Errorf("expected initial call to return zero info before completion")
	}

	select {
	case <-notifyChan:
	case <-time.After(2 * time.Second):
		t.Fatal("mock fetch timed out")
	}

	fresh := svc.Info()
	if !fresh.IsSuccess {
		t.Fatal("expected fresh info to be marked success")
	}
	if fresh.Country != "中国" || fresh.City != "上海" {
		t.Fatalf("expected country 中国 city 上海, got %s %s", fresh.Country, fresh.City)
	}
	if !updated {
		t.Fatal("expected onUpdate callback to be invoked")
	}
}
