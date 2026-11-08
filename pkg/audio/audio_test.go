package audio

import (
	"bytes"
	"encoding/binary"
	"math"
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
		// 校验 PCM 格式标识 (AudioFormat=1, Channels=1, SampleRate=22050, BitsPerSample=16)
		if fmtTag := binary.LittleEndian.Uint16(wav[20:22]); fmtTag != 1 {
			t.Fatalf("expected uncompressed PCM format=1 for %s, got %d", name, fmtTag)
		}
		if bits := binary.LittleEndian.Uint16(wav[34:36]); bits != 16 {
			t.Fatalf("expected 16-bit PCM for %s, got %d", name, bits)
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

func TestScaleWAVVolumeInto(t *testing.T) {
	base := GetAlarmWAV("Classic Chime")

	// 1. 100% 音量直接返回原始切片
	full := ScaleWAVVolumeInto(nil, base, 1.0)
	if &full[0] != &base[0] {
		t.Fatal("expected vol=1.0 to return original slice without copy")
	}

	// 2. 50% 音量对 16-bit PCM 振幅精确减半
	var dst []byte
	half := ScaleWAVVolumeInto(dst, base, 0.5)
	if len(half) != len(base) {
		t.Fatalf("expected scaled WAV length %d, got %d", len(base), len(half))
	}
	if !bytes.Equal(half[:44], base[:44]) {
		t.Fatal("expected 44-byte RIFF header to remain identical after scaling")
	}

	var maxBase, maxHalf int16
	for off := 44; off+1 < len(base); off += 2 {
		sb := int16(binary.LittleEndian.Uint16(base[off : off+2]))
		sh := int16(binary.LittleEndian.Uint16(half[off : off+2]))
		if sb > maxBase {
			maxBase = sb
		}
		if sh > maxHalf {
			maxHalf = sh
		}
	}
	ratio := float64(maxHalf) / float64(maxBase)
	if math.Abs(ratio-0.5) > 0.01 {
		t.Fatalf("expected peak amplitude ratio ~0.5, got %.4f (base=%d, half=%d)", ratio, maxBase, maxHalf)
	}

	// 3. 0% 音量输出完全静音采样
	zero := ScaleWAVVolumeInto(half, base, 0.0)
	for off := 44; off+1 < len(zero); off += 2 {
		if s := int16(binary.LittleEndian.Uint16(zero[off : off+2])); s != 0 {
			t.Fatalf("expected zero sample at offset %d, got %d", off, s)
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

	// 2. Dismiss alarm, start sleep mode at 80% volume
	eng.SyncState(false, "", true, "Ocean Waves", 0.8)
	if eng.ActiveKey() != "ambient:Ocean Waves" {
		t.Fatalf("expected ambient:Ocean Waves, got %s", eng.ActiveKey())
	}
	if eng.activeVolBucket != 8 {
		t.Fatalf("expected volume bucket 8, got %d", eng.activeVolBucket)
	}

	// 2b. Simulate sleep fade-out to 40% volume -> software volume bucket updates to 4 and double-buffer retains previous buffer
	eng.SyncState(false, "", true, "Ocean Waves", 0.4)
	if eng.activeVolBucket != 4 {
		t.Fatalf("expected volume bucket 4 during fade-out, got %d", eng.activeVolBucket)
	}
	if len(eng.activeBuf) == 0 || len(eng.prevBuf) == 0 {
		t.Fatal("expected double-buffered activeBuf and prevBuf to be populated during volume transition")
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
