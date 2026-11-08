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
	if len(s2.WorldCities) == 0 {
		t.Errorf("expected default WorldCities to be preserved when updating unrelated settings")
	}
	if s2.CloseBehavior != CloseBehaviorMinimize {
		t.Errorf("expected default CloseBehavior %q, got %q", CloseBehaviorMinimize, s2.CloseBehavior)
	}
}

func TestStorageDeepCopyAndBackupRecovery(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "settings.json")

	store := NewStore(path)
	// First update writes settings.json
	if err := store.Update(func(cfg *Settings) {
		cfg.CloseBehavior = CloseBehaviorExit
	}); err != nil {
		t.Fatalf("first update failed: %v", err)
	}
	// Second update rotates first settings.json to settings.json.bak
	if err := store.Update(func(cfg *Settings) {
		cfg.Theme = ThemeNord
	}); err != nil {
		t.Fatalf("second update failed: %v", err)
	}

	// Verify Get() deep-copies WorldCities
	snap := store.Get()
	origFirst := snap.WorldCities[0].CityID
	snap.WorldCities[0].CityID = "mutated-outside-store"
	if store.Get().WorldCities[0].CityID != origFirst {
		t.Fatal("mutating slice returned by Get() should not mutate Store internal state")
	}

	// Simulate corrupted primary settings.json and verify automatic recovery from .bak
	if err := os.WriteFile(path, []byte("{corrupted-json"), 0644); err != nil {
		t.Fatalf("failed to corrupt primary file: %v", err)
	}
	recovered := NewStore(path)
	if recovered.Get().CloseBehavior != CloseBehaviorExit {
		t.Fatalf("expected recovery from .bak to restore CloseBehaviorExit, got %q", recovered.Get().CloseBehavior)
	}
	if len(recovered.Get().WorldCities) == 0 {
		t.Fatal("expected recovered settings to retain WorldCities")
	}
}

func TestStorageZeroDiskWriteWhenUnchanged(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "settings.json")

	store := NewStore(path)
	if err := store.Save(); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	// Corrupt disk file in-memory check: if we call Save() when settings are unchanged,
	// saveLocked() must skip disk I/O completely because bytes.Equal(data, s.lastSaved) is true.
	sentinel := []byte(`{"sentinel":true}`)
	if err := os.WriteFile(path, sentinel, 0644); err != nil {
		t.Fatalf("write sentinel failed: %v", err)
	}
	if err := store.Save(); err != nil {
		t.Fatalf("unchanged save failed: %v", err)
	}
	afterBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after unchanged save failed: %v", err)
	}
	if string(afterBytes) != string(sentinel) {
		t.Fatalf("expected unchanged Save() to perform zero disk writes, but file was overwritten")
	}
}
