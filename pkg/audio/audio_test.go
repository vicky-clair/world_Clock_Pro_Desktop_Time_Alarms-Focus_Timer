package audio

import (
	"bytes"
	"testing"
)

func TestSynthesizeAlarmAndAmbientWAVs(t *testing.T) {
	alarms := []string{"Classic Chime", "Gentle Chime", "Digital Bell"}
	for _, name := range alarms {
		wav := GetAlarmWAV(name)
		if len(wav) < 1000 {
			t.Fatalf("expected valid WAV for alarm %s, got len=%d", name, len(wav))
		}
		if !bytes.Equal(wav[0:4], []byte("RIFF")) || !bytes.Equal(wav[8:12], []byte("WAVE")) {
			t.Fatalf("invalid RIFF WAVE header for %s", name)
		}
	}

	ambients := []string{"Gentle Rain", "Ocean Waves", "Forest Stream", "White Noise"}
	for _, name := range ambients {
		wav := GetAmbientWAV(name)
		if len(wav) < 10000 {
			t.Fatalf("expected valid WAV for ambient %s, got len=%d", name, len(wav))
		}
		if !bytes.Equal(wav[0:4], []byte("RIFF")) || !bytes.Equal(wav[8:12], []byte("WAVE")) {
			t.Fatalf("invalid RIFF WAVE header for %s", name)
		}
	}
}

func TestEngineStateSync(t *testing.T) {
	eng := NewEngine()

	// 1. Trigger alarm ringing
	eng.SyncState(true, "Gentle Chime", false, "", 0)
	if eng.ActiveKey() != "alarm:Gentle Chime" {
		t.Fatalf("expected alarm:Gentle Chime, got %s", eng.ActiveKey())
	}

	// 2. Dismiss alarm, start sleep mode
	eng.SyncState(false, "", true, "Ocean Waves", 0.8)
	if eng.ActiveKey() != "ambient:Ocean Waves" {
		t.Fatalf("expected ambient:Ocean Waves, got %s", eng.ActiveKey())
	}

	// 3. Stop sleep mode
	eng.SyncState(false, "", false, "Ocean Waves", 0)
	if eng.ActiveKey() != "" {
		t.Fatalf("expected empty active key after stop, got %s", eng.ActiveKey())
	}

	// 4. Preview sound
	eng.PreviewAlarm("Digital Bell")
	if eng.PreviewKey() != "alarm:Digital Bell" {
		t.Fatalf("expected preview alarm:Digital Bell, got %s", eng.PreviewKey())
	}
}
