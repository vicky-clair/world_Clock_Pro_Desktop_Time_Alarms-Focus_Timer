package ui

import (
	"path/filepath"
	"testing"

	"worldclock/pkg/storage"

	"github.com/vivaui/viva/app"
	vui "github.com/vivaui/viva/ui"
)

func BenchmarkPaintMainClock(b *testing.B) {
	tmpDir := b.TempDir()
	store := storage.NewStore(filepath.Join(tmpDir, "bench_settings.json"))
	state := NewAppState(store)

	var canvas vui.Canvas
	cfg := NewAppConfig(state)
	cfg.Width = 900
	cfg.Height = 800

	driver, err := app.NewDriver(cfg, &canvas)
	if err != nil {
		b.Fatalf("failed to create driver: %v", err)
	}
	defer driver.Close()

	// Initial warmup paint
	_ = driver.Resize(900, 800)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		// Drive repaint via resize
		_ = driver.Resize(900, 800)
	}
}
