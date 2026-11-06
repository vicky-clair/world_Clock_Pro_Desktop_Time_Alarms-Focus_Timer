package action

import (
	"testing"
)

func TestAlarmActionValidation(t *testing.T) {
	validURL := AlarmAction{
		Type:   TypeOpenURL,
		Target: "https://example.com/daily-briefing",
	}
	if err := validURL.Validate(); err != nil {
		t.Fatalf("expected valid URL, got: %v", err)
	}

	invalidScheme := AlarmAction{
		Type:   TypeOpenURL,
		Target: "ftp://example.com/resource",
	}
	if err := invalidScheme.Validate(); err == nil {
		t.Fatal("expected error for non-http scheme")
	}

	loopback := AlarmAction{
		Type:   TypeOpenURL,
		Target: "http://localhost:8080/status",
	}
	if err := loopback.Validate(); err == nil {
		t.Fatal("expected error for loopback target")
	}

	emptyTarget := AlarmAction{
		Type:   TypeOpenURL,
		Target: "   ",
	}
	if err := emptyTarget.Validate(); err == nil {
		t.Fatal("expected error for empty target")
	}
}
