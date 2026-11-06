package main

import (
	"path/filepath"
	"testing"

	"worldclock/pkg/storage"
	"worldclock/pkg/ui"

	"github.com/vivaui/viva/app"
	vui "github.com/vivaui/viva/ui"
)

func TestAppDriverSmoke(t *testing.T) {
	tmpDir := t.TempDir()
	store := storage.NewStore(filepath.Join(tmpDir, "settings.json"))
	state := ui.NewAppState(store)

	var canvas vui.Canvas
	cfg := ui.NewAppConfig(state)

	driver, err := app.NewDriver(cfg, &canvas)
	if err != nil {
		t.Fatalf("failed to create driver: %v", err)
	}
	defer driver.Close()

	snap := driver.Snapshot()
	if len(snap.Hits) == 0 {
		t.Fatal("expected interactive hit controls on screen")
	}

	// Verify header and clock are painted
	hasClock := false
	for _, l := range snap.Labels {
		if l.Value == "时钟" || l.Value == "Clock" {
			hasClock = true
			break
		}
	}
	if !hasClock {
		t.Fatal("expected '时钟' or 'Clock' tab in snapshot")
	}
}
