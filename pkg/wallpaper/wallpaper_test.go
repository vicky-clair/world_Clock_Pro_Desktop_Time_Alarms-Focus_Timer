package wallpaper

import (
	"image/color"
	"testing"
	"time"

	"github.com/vivaui/viva/ui"
)

func TestImageToBitmap(t *testing.T) {
	img := CreateTestImage(64, 64, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	bmp := ImageToBitmap(img, 42)

	if bmp.Width != 64 || bmp.Height != 64 {
		t.Fatalf("expected 64x64, got %dx%d", bmp.Width, bmp.Height)
	}
	if bmp.Stride != 64*4 {
		t.Fatalf("expected stride %d, got %d", 64*4, bmp.Stride)
	}
	// Check BGRA ordering
	b := bmp.Pix[0]
	g := bmp.Pix[1]
	r := bmp.Pix[2]
	a := bmp.Pix[3]

	if b != 50 || g != 100 || r != 200 || a != 255 {
		t.Errorf("unexpected pixel values: B=%d, G=%d, R=%d, A=%d", b, g, r, a)
	}
}

func TestWallpaperLifecycleAndFileLoading(t *testing.T) {
	imgPath, err := EnsureSampleLocalFile()
	if err != nil {
		t.Fatal(err)
	}

	mgr := NewManager(DefaultConfig())
	mgr.SetEnabled(true)
	mgr.SetFitMode(FitCover)
	mgr.SetOpacity(0.8)

	if err := mgr.LoadImage(imgPath); err != nil {
		t.Fatalf("failed to load image: %v", err)
	}

	if mgr.Config().ImagePath != imgPath || !mgr.Config().Enabled {
		t.Fatal("expected image path to be set and enabled")
	}

	// Test drawing into canvas with different fit modes
	var canvas ui.Canvas
	canvas.BeginFrame()
	theme := ui.DarkTheme()
	bounds := ui.Rect{X: 0, Y: 0, Width: 800, Height: 600}

	// Test all fit modes
	modes := []FitMode{FitCover, FitContain, FitStretch, FitTile, FitCenter}
	for _, mode := range modes {
		mgr.SetFitMode(mode)
		mgr.Draw(&canvas, bounds, theme)
	}
	canvas.EndFrame()
}

func TestSlideshowFolderAndIntervalsAndShuffle(t *testing.T) {
	mgr := NewManager(DefaultConfig())
	uris := SampleSlideshowURIs()
	count := mgr.LoadImages(uris)
	if count != 4 {
		t.Fatalf("expected 4 images indexed, got %d", count)
	}
	cfg := mgr.Config()
	if !cfg.Enabled || !cfg.SlideshowEnabled || len(cfg.ImagePaths) != 4 {
		t.Fatalf("unexpected slideshow config after LoadImages: %+v", cfg)
	}

	// Test Sequential Mode + 10-minute interval
	mgr.SetSlideshowOrder(OrderSequential)
	mgr.SetIntervalMinutes(10)
	if mgr.Config().IntervalMinutes != 10 {
		t.Fatalf("expected 10m interval, got %d", mgr.Config().IntervalMinutes)
	}

	startIdx := mgr.Config().CurrentIndex
	baseNow := time.Now()
	// Not due yet after 5 minutes
	if mgr.AdvanceIfDue(baseNow.Add(5 * time.Minute)) {
		t.Fatal("should not advance at 5m when interval is 10m")
	}
	// Due at 10 minutes + 1s
	t1 := baseNow.Add(10*time.Minute + time.Second)
	if !mgr.AdvanceIfDue(t1) {
		t.Fatal("expected slideshow to advance at 10m deadline")
	}
	if mgr.Config().CurrentIndex != (startIdx+1)%4 {
		t.Fatalf("expected sequential index %d, got %d", (startIdx+1)%4, mgr.Config().CurrentIndex)
	}

	// Test 15-minute and 1-minute intervals
	mgr.SetIntervalMinutes(15)
	if mgr.Config().IntervalMinutes != 15 {
		t.Fatalf("expected 15m interval, got %d", mgr.Config().IntervalMinutes)
	}
	mgr.SetIntervalMinutes(1)
	if mgr.Config().IntervalMinutes != 1 {
		t.Fatalf("expected 1m interval, got %d", mgr.Config().IntervalMinutes)
	}

	// Test manual NextSlide / PrevSlide
	idxBefore := mgr.Config().CurrentIndex
	if !mgr.NextSlide(t1) {
		t.Fatal("NextSlide returned false")
	}
	if mgr.Config().CurrentIndex != (idxBefore+1)%4 {
		t.Fatalf("expected NextSlide index %d, got %d", (idxBefore+1)%4, mgr.Config().CurrentIndex)
	}
	if !mgr.PrevSlide(t1) {
		t.Fatal("PrevSlide returned false")
	}
	if mgr.Config().CurrentIndex != idxBefore {
		t.Fatalf("expected PrevSlide index %d, got %d", idxBefore, mgr.Config().CurrentIndex)
	}

	// Test Shuffle Mode visits all 4 photos without immediate repeat
	mgr.SetSlideshowOrder(OrderShuffle)
	visited := make(map[int]bool)
	for range 4 {
		mgr.NextSlide(t1)
		visited[mgr.Config().CurrentIndex] = true
	}
	if len(visited) != 4 {
		t.Fatalf("expected Fisher-Yates shuffle to visit all 4 photos in 1 cycle, visited %d", len(visited))
	}
}

func TestDownsampleImage4K(t *testing.T) {
	// 3840x2160 4K image should be downsampled to 1920x1080
	big := CreateTestImage(3840, 2160, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	small := DownsampleImage(big, 1920)
	if small.Bounds().Dx() != 1920 || small.Bounds().Dy() != 1080 {
		t.Fatalf("expected 1920x1080 downsampled image, got %dx%d", small.Bounds().Dx(), small.Bounds().Dy())
	}
}

func BenchmarkAdvanceIfDueIdle(b *testing.B) {
	mgr := NewManager(DefaultConfig())
	mgr.SetEnabled(true)
	mgr.SetSlideshowEnabled(true)
	now := time.Now()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = mgr.AdvanceIfDue(now)
	}
}

func BenchmarkImageToBitmapFastPath(b *testing.B) {
	img := CreateTestImage(1280, 720, color.RGBA{R: 60, G: 120, B: 180, A: 255})
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		_ = ImageToBitmapWithOpacity(img, 0.75, uint64(i+1))
	}
}
