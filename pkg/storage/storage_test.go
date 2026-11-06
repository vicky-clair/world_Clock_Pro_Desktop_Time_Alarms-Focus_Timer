package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoragePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_settings.json")

	store := NewStore(path)
	s := store.Get()
	if !s.TimeFormat24H {
		t.Errorf("expected default 24H to be true")
	}

	err := store.Update(func(cfg *Settings) {
		cfg.TimeFormat24H = false
		cfg.ClockStyle = "analog"
		cfg.Theme = ThemeDark
	})
	if err != nil {
		t.Fatalf("failed to update store: %v", err)
	}

	// Verify file was written
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("settings file was not created: %v", err)
	}

	// Load in a fresh store instance
	store2 := NewStore(path)
	s2 := store2.Get()
	if s2.TimeFormat24H != false {
		t.Errorf("expected TimeFormat24H to be false after reload")
	}
	if s2.ClockStyle != "analog" {
		t.Errorf("expected ClockStyle to be analog after reload")
	}
	if s2.Theme != ThemeDark {
		t.Errorf("expected Theme to be dark after reload")
	}
}
