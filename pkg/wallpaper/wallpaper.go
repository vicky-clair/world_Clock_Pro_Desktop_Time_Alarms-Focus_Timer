// Package wallpaper 实现零常驻后台开销的桌面背景与壁纸轮播引擎。
// 采用“仅索引路径 + 按需单张解码 + 自动 1080p 降采样 + Fisher-Yates 不重复洗牌”架构，
// 确保即使导入数千张高分辨率照片轮播也不会影响电脑内存与 CPU/GPU 性能。
package wallpaper

import (
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vivaui/viva/ui"
)

// FitMode 定义背景壁纸在窗口中的缩放与对齐模式。
// FitMode describes how the background image is positioned and scaled.
type FitMode string

const (
	FitCover   FitMode = "cover"   // 填充 / 铺满 (Fill preserving aspect ratio)
	FitContain FitMode = "contain" // 适应 / 保持比例 (Fit inside bounds)
	FitStretch FitMode = "stretch" // 拉伸 (Stretch to fill bounds)
	FitTile    FitMode = "tile"    // 平铺 (Repeat across bounds)
	FitCenter  FitMode = "center"  // 居中 (Original size centered)
)

// SlideshowOrder 定义多张壁纸轮播时的切换顺序模式。
// SlideshowOrder describes how multiple wallpaper photos are rotated.
type SlideshowOrder string

const (
	OrderShuffle    SlideshowOrder = "shuffle"    // 随机播放 (Fisher-Yates non-repeating shuffle)
	OrderSequential SlideshowOrder = "sequential" // 顺序播放 (In-order loop)
)

// maxWallpaperDim 限制解码驻留内存的壁纸最大边长为 1920 像素（1080p），避免 4K/8K 原图挤占内存与显存带宽。
// maxWallpaperDim caps decoded in-memory wallpaper dimensions so 4K/8K photos
// never bloat RAM or GPU upload bandwidth.
const maxWallpaperDim = 1920

// Config 表示壁纸与自动轮播的持久化配置参数。
// Config represents wallpaper and slideshow settings.
type Config struct {
	Enabled          bool           `json:"enabled"`              // 是否启用自定义/预设壁纸背景
	ImagePath        string         `json:"imagePath"`            // 当前显示的图片路径（为空则使用内置艺术预设）
	Preset           PresetID       `json:"preset"`               // 内置预设 ID ("grid", "aurora", "forest", "sunset")
	FitMode          FitMode        `json:"fitMode"`              // 适应模式
	Opacity          float32        `json:"opacity"`              // 背景壁纸不透明度 (0.0 ~ 1.0)
	CardOpacity      float32        `json:"cardOpacity"`          // 卡片毛玻璃不透明度 (0.2 ~ 1.0)
	SlideshowEnabled bool           `json:"slideshowEnabled"`     // 是否开启多图轮播
	IntervalMinutes  int            `json:"intervalMinutes"`      // 轮播间隔分钟数 (1, 5, 10, 15)
	SlideshowOrder   SlideshowOrder `json:"slideshowOrder"`       // 播放顺序 ("shuffle" 或 "sequential")
	ImagePaths       []string       `json:"imagePaths,omitempty"` // 轮播图路径索引列表
	FolderPath       string         `json:"folderPath,omitempty"` // 轮播图所属文件夹路径
	CurrentIndex     int            `json:"currentIndex"`         // 当前播放到的图片下标
}

// DefaultConfig 返回默认的壁纸与轮播配置。
// DefaultConfig returns clean default wallpaper configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:          false,
		ImagePath:        "",
		Preset:           PresetGrid,
		FitMode:          FitCover,
		Opacity:          0.75,
		CardOpacity:      0.45,
		SlideshowEnabled: false,
		IntervalMinutes:  5,
		SlideshowOrder:   OrderShuffle,
	}
}

func normalizeConfig(cfg Config) Config {
	if cfg.Preset == "" {
		cfg.Preset = PresetGrid
	}
	if cfg.FitMode == "" {
		cfg.FitMode = FitCover
	}
	if cfg.Opacity <= 0 {
		cfg.Opacity = 0.75
	}
	if cfg.CardOpacity <= 0 {
		cfg.CardOpacity = 0.45
	}
	if cfg.IntervalMinutes != 1 && cfg.IntervalMinutes != 5 && cfg.IntervalMinutes != 10 && cfg.IntervalMinutes != 15 {
		cfg.IntervalMinutes = 5
	}
	if cfg.SlideshowOrder != OrderSequential && cfg.SlideshowOrder != OrderShuffle {
		cfg.SlideshowOrder = OrderShuffle
	}
	return cfg
}

// IsSupportedImageExt reports whether filePath has a supported wallpaper image extension or built-in preset URI.
func IsSupportedImageExt(filePath string) bool {
	if strings.HasPrefix(filePath, "@preset:") {
		return true
	}
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".bmp"
}

// ScanFolderImages scans a directory using os.ReadDir without decoding any images,
// returning only file paths of supported image formats.
func ScanFolderImages(folderPath string) ([]string, error) {
	cleanDir := filepath.Clean(strings.Trim(folderPath, "\"' "))
	entries, err := os.ReadDir(cleanDir)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if IsSupportedImageExt(name) {
			paths = append(paths, filepath.Join(cleanDir, name))
		}
	}
	return paths, nil
}

var bitmapIDCounter uint64 = 1000

// Manager handles caching, on-demand single-frame decoding, downsampling,
// deadline-based slideshow rotation, and rendering of background wallpapers.
type Manager struct {
	mu          sync.RWMutex
	cfg         Config
	rawImage    image.Image // single active decoded & downsampled image
	cachedPath  string
	hasBitmap   bool
	bitmap      ui.Bitmap
	nextSlideAt time.Time
	shufflePerm []int
	shufflePos  int
}

// NewManager creates a wallpaper manager.
func NewManager(cfg Config) *Manager {
	cfg = normalizeConfig(cfg)
	m := &Manager{
		cfg: cfg,
	}

	// If a folder path was persisted and ImagePaths is empty, re-index paths lazily (fast os.ReadDir).
	if cfg.FolderPath != "" && len(cfg.ImagePaths) == 0 {
		if paths, err := ScanFolderImages(cfg.FolderPath); err == nil && len(paths) > 0 {
			m.cfg.ImagePaths = paths
		}
	}

	if len(m.cfg.ImagePaths) > 0 {
		if m.cfg.CurrentIndex < 0 || m.cfg.CurrentIndex >= len(m.cfg.ImagePaths) {
			m.cfg.CurrentIndex = 0
		}
		m.rebuildShuffleLocked()
		if m.cfg.ImagePath == "" {
			m.cfg.ImagePath = m.cfg.ImagePaths[m.cfg.CurrentIndex]
		}
	}

	if m.cfg.ImagePath != "" {
		m.loadImageLocked(m.cfg.ImagePath)
	}
	if !m.hasBitmap {
		m.rawImage = GeneratePresetImage(m.cfg.Preset)
		m.rebuildBitmapLocked()
	}
	m.resetDeadlineLocked(time.Now())
	return m
}

func (m *Manager) intervalDurationLocked() time.Duration {
	mins := m.cfg.IntervalMinutes
	if mins != 1 && mins != 5 && mins != 10 && mins != 15 {
		mins = 5
	}
	return time.Duration(mins) * time.Minute
}

func (m *Manager) resetDeadlineLocked(now time.Time) {
	if m.cfg.Enabled && m.cfg.SlideshowEnabled && len(m.cfg.ImagePaths) > 1 {
		m.nextSlideAt = now.Add(m.intervalDurationLocked())
	} else {
		m.nextSlideAt = time.Time{}
	}
}

func (m *Manager) rebuildShuffleLocked() {
	n := len(m.cfg.ImagePaths)
	if n <= 1 {
		m.shufflePerm = nil
		m.shufflePos = 0
		return
	}
	perm := rand.Perm(n)
	// Ensure the first item in a new shuffle isn't identical to the current active index.
	if perm[0] == m.cfg.CurrentIndex {
		perm[0], perm[n-1] = perm[n-1], perm[0]
	}
	m.shufflePerm = perm
	m.shufflePos = -1
}

// Config returns a copy of wallpaper settings.
func (m *Manager) Config() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cfgCopy := m.cfg
	if len(m.cfg.ImagePaths) > 0 {
		cfgCopy.ImagePaths = make([]string, len(m.cfg.ImagePaths))
		copy(cfgCopy.ImagePaths, m.cfg.ImagePaths)
	}
	return cfgCopy
}

// RenderParams O(1) 零内存分配读取壁纸启用状态与卡片毛玻璃不透明度，专供每帧 Paint 高频调用，避免拷贝 ImagePaths 切片。
func (m *Manager) RenderParams() (enabled bool, cardOpacity float32) {
	m.mu.RLock()
	enabled = m.cfg.Enabled
	cardOpacity = m.cfg.CardOpacity
	m.mu.RUnlock()
	return enabled, cardOpacity
}

// SetConfig updates wallpaper configuration.
func (m *Manager) SetConfig(cfg Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg = normalizeConfig(cfg)
	m.rebuildShuffleLocked()
	if m.cfg.ImagePath != "" {
		m.loadImageLocked(m.cfg.ImagePath)
	} else {
		m.rawImage = GeneratePresetImage(m.cfg.Preset)
		m.rebuildBitmapLocked()
	}
	m.resetDeadlineLocked(time.Now())
}

// SetFitMode updates the fit mode.
func (m *Manager) SetFitMode(mode FitMode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg.FitMode = mode
}

// SetOpacity updates background image opacity (0.0 to 1.0) and rebuilds bitmap pixels with exact premultiplied alpha.
func (m *Manager) SetOpacity(opacity float32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if opacity < 0.05 {
		opacity = 0.05
	} else if opacity > 1.0 {
		opacity = 1.0
	}
	m.cfg.Opacity = opacity
	m.rebuildBitmapLocked()
}

// SetCardOpacity updates card glass opacity (0.2 to 1.0).
func (m *Manager) SetCardOpacity(op float32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if op < 0.20 {
		op = 0.20
	} else if op > 1.0 {
		op = 1.0
	}
	m.cfg.CardOpacity = op
}

// SetEnabled enables or disables the wallpaper.
func (m *Manager) SetEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg.Enabled = enabled
	m.resetDeadlineLocked(time.Now())
}

// SetSlideshowEnabled enables or disables automatic wallpaper carousel rotation.
func (m *Manager) SetSlideshowEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg.SlideshowEnabled = enabled
	m.resetDeadlineLocked(time.Now())
}

// SetIntervalMinutes sets the rotation interval (1, 5, 10, or 15 minutes).
func (m *Manager) SetIntervalMinutes(minutes int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if minutes != 1 && minutes != 5 && minutes != 10 && minutes != 15 {
		minutes = 5
	}
	m.cfg.IntervalMinutes = minutes
	m.resetDeadlineLocked(time.Now())
}

// SetSlideshowOrder sets the playback order ("shuffle" or "sequential").
func (m *Manager) SetSlideshowOrder(order SlideshowOrder) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if order != OrderSequential && order != OrderShuffle {
		order = OrderShuffle
	}
	m.cfg.SlideshowOrder = order
	if order == OrderShuffle {
		m.rebuildShuffleLocked()
	}
}

// SetPreset switches to a built-in procedural preset wallpaper.
func (m *Manager) SetPreset(preset PresetID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg.Preset = preset
	m.cfg.ImagePath = ""
	m.cfg.ImagePaths = nil
	m.cfg.FolderPath = ""
	m.cfg.SlideshowEnabled = false
	m.cfg.Enabled = true
	m.nextSlideAt = time.Time{}
	m.rawImage = GeneratePresetImage(preset)
	m.rebuildBitmapLocked()
}

// ClearCustomImage reverts to preset and clears custom slideshow list.
func (m *Manager) ClearCustomImage() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg.ImagePath = ""
	m.cfg.ImagePaths = nil
	m.cfg.FolderPath = ""
	m.cfg.CurrentIndex = 0
	m.cfg.SlideshowEnabled = false
	m.nextSlideAt = time.Time{}
	m.rawImage = GeneratePresetImage(m.cfg.Preset)
	m.rebuildBitmapLocked()
}

// DownsampleImage scales an oversized image down so neither width nor height exceeds maxDim.
// Images already within maxDim are returned as a compact *image.RGBA.
func DownsampleImage(src image.Image, maxDim int) *image.RGBA {
	b := src.Bounds()
	srcW, srcH := b.Dx(), b.Dy()
	if srcW <= 0 || srcH <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	dstW, dstH := srcW, srcH
	if srcW > maxDim || srcH > maxDim {
		if srcW >= srcH {
			dstW = maxDim
			dstH = max(1, (srcH*maxDim)/srcW)
		} else {
			dstH = max(1, maxDim)
			dstW = max(1, (srcW*maxDim)/srcH)
		}
	}
	if dstW == srcW && dstH == srcH {
		if rgba, ok := src.(*image.RGBA); ok && b.Min.X == 0 && b.Min.Y == 0 {
			return rgba
		}
	}
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	dstPix := dst.Pix
	dstStride := dst.Stride

	// Fast nearest/box sampling into compact RGBA
	for y := range dstH {
		srcY := b.Min.Y + (y*srcH)/dstH
		rowOff := y * dstStride
		for x := range dstW {
			srcX := b.Min.X + (x*srcW)/dstW
			r, g, bVal, a := src.At(srcX, srcY).RGBA()
			off := rowOff + x*4
			dstPix[off+0] = byte(r >> 8)
			dstPix[off+1] = byte(g >> 8)
			dstPix[off+2] = byte(bVal >> 8)
			dstPix[off+3] = byte(a >> 8)
		}
	}
	return dst
}

func (m *Manager) loadImageLocked(filePath string) bool {
	trimmed := strings.Trim(filePath, "\"' ")
	if after, ok := strings.CutPrefix(trimmed, "@preset:"); ok {
		presetID := PresetID(after)
		m.rawImage = GeneratePresetImage(presetID)
		m.cachedPath = trimmed
		m.cfg.ImagePath = trimmed
		m.rebuildBitmapLocked()
		return true
	}

	cleanPath := filepath.Clean(trimmed)

	f, err := os.Open(cleanPath)
	if err != nil {
		// Fallback to preset if no image is currently loaded
		if m.rawImage == nil {
			m.rawImage = GeneratePresetImage(m.cfg.Preset)
			m.rebuildBitmapLocked()
		}
		return false
	}
	defer f.Close()

	// 先读取图片头部尺寸元数据，拒绝超过 16384x16384 的异常超大文件以防范内存暴涨
	if imgCfg, _, err := image.DecodeConfig(f); err != nil || imgCfg.Width <= 0 || imgCfg.Height <= 0 || imgCfg.Width > 16384 || imgCfg.Height > 16384 {
		if m.rawImage == nil {
			m.rawImage = GeneratePresetImage(m.cfg.Preset)
			m.rebuildBitmapLocked()
		}
		return false
	}
	if _, err := f.Seek(0, 0); err != nil {
		return false
	}

	img, _, err := image.Decode(f)
	if err != nil {
		if m.rawImage == nil {
			m.rawImage = GeneratePresetImage(m.cfg.Preset)
			m.rebuildBitmapLocked()
		}
		return false
	}

	// Downsample oversized 4K/8K photos immediately so only compact pixels stay in memory
	m.rawImage = DownsampleImage(img, maxWallpaperDim)
	m.cachedPath = cleanPath
	m.cfg.ImagePath = cleanPath
	m.rebuildBitmapLocked()
	return true
}

// LoadImage loads a single local PNG, JPEG, or BMP file (or @preset: URI) into the wallpaper engine.
func (m *Manager) LoadImage(filePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg.Enabled = true
	m.cfg.FolderPath = ""
	trimmed := strings.Trim(filePath, "\"' ")
	if !strings.HasPrefix(trimmed, "@preset:") {
		trimmed = filepath.Clean(trimmed)
	}
	m.cfg.ImagePaths = []string{trimmed}
	m.cfg.CurrentIndex = 0
	m.cfg.SlideshowEnabled = false
	m.nextSlideAt = time.Time{}
	m.loadImageLocked(trimmed)
	return nil
}

// LoadImages loads a list of local image paths into the wallpaper carousel.
// Only the first/active image is decoded immediately; remaining paths stay indexed on disk.
func (m *Manager) LoadImages(paths []string) int {
	valid := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		clean := strings.Trim(p, "\"' ")
		if !strings.HasPrefix(clean, "@preset:") {
			clean = filepath.Clean(clean)
		}
		if !IsSupportedImageExt(clean) {
			continue
		}
		if _, exists := seen[clean]; exists {
			continue
		}
		seen[clean] = struct{}{}
		valid = append(valid, clean)
	}
	if len(valid) == 0 {
		return 0
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg.Enabled = true
	m.cfg.ImagePaths = valid
	m.cfg.CurrentIndex = 0
	m.rebuildShuffleLocked()
	if len(valid) > 1 {
		m.cfg.SlideshowEnabled = true
		if m.cfg.SlideshowOrder == OrderShuffle && len(m.shufflePerm) > 0 {
			m.shufflePos = 0
			m.cfg.CurrentIndex = m.shufflePerm[0]
		}
	} else {
		m.cfg.SlideshowEnabled = false
	}
	m.loadImageLocked(valid[m.cfg.CurrentIndex])
	m.resetDeadlineLocked(time.Now())
	return len(valid)
}

// LoadFolder indexes all supported images in folderPath and activates the carousel.
func (m *Manager) LoadFolder(folderPath string) (int, error) {
	paths, err := ScanFolderImages(folderPath)
	if err != nil {
		return 0, err
	}
	if len(paths) == 0 {
		return 0, nil
	}
	count := m.LoadImages(paths)
	m.mu.Lock()
	m.cfg.FolderPath = filepath.Clean(strings.Trim(folderPath, "\"' "))
	m.mu.Unlock()
	return count, nil
}

// NextSlide advances to the next wallpaper in the playlist immediately.
func (m *Manager) NextSlide(now time.Time) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stepSlideLocked(1, now)
}

// PrevSlide steps back to the previous wallpaper in the playlist immediately.
func (m *Manager) PrevSlide(now time.Time) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stepSlideLocked(-1, now)
}

func (m *Manager) stepSlideLocked(dir int, now time.Time) bool {
	n := len(m.cfg.ImagePaths)
	if n <= 1 {
		return false
	}

	// Try up to n paths in case any file on disk was moved/deleted
	for range n {
		var nextIdx int
		if m.cfg.SlideshowOrder == OrderShuffle {
			if len(m.shufflePerm) != n {
				m.rebuildShuffleLocked()
			}
			if dir >= 0 {
				m.shufflePos++
				if m.shufflePos >= n {
					m.rebuildShuffleLocked()
					m.shufflePos = 0
				}
			} else {
				if m.shufflePos <= 0 {
					m.shufflePos = n - 1
				} else {
					m.shufflePos--
				}
			}
			nextIdx = m.shufflePerm[m.shufflePos]
		} else {
			if dir >= 0 {
				nextIdx = (m.cfg.CurrentIndex + 1) % n
			} else {
				nextIdx = (m.cfg.CurrentIndex - 1 + n) % n
			}
		}

		m.cfg.CurrentIndex = nextIdx
		if m.loadImageLocked(m.cfg.ImagePaths[nextIdx]) {
			m.resetDeadlineLocked(now)
			return true
		}
	}
	m.resetDeadlineLocked(now)
	return false
}

// AdvanceIfDue checks whether the slideshow deadline has arrived.
// Fast path is O(1) and zero-allocation under RLock.
func (m *Manager) AdvanceIfDue(now time.Time) bool {
	m.mu.RLock()
	due := m.cfg.Enabled &&
		m.cfg.SlideshowEnabled &&
		len(m.cfg.ImagePaths) > 1 &&
		!m.nextSlideAt.IsZero() &&
		!now.Before(m.nextSlideAt)
	m.mu.RUnlock()

	if !due {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	// Re-check under write lock
	if !m.cfg.Enabled || !m.cfg.SlideshowEnabled || len(m.cfg.ImagePaths) <= 1 || m.nextSlideAt.IsZero() || now.Before(m.nextSlideAt) {
		return false
	}
	return m.stepSlideLocked(1, now)
}

// NextSlideDelay returns the time remaining until the next scheduled slide transition,
// or 0 if slideshow rotation is not currently active.
func (m *Manager) NextSlideDelay(now time.Time) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.cfg.Enabled || !m.cfg.SlideshowEnabled || len(m.cfg.ImagePaths) <= 1 || m.nextSlideAt.IsZero() {
		return 0
	}
	d := m.nextSlideAt.Sub(now)
	if d < 100*time.Millisecond {
		return 100 * time.Millisecond
	}
	return d
}

func (m *Manager) rebuildBitmapLocked() {
	if m.rawImage == nil {
		m.hasBitmap = false
		return
	}
	id := atomic.AddUint64(&bitmapIDCounter, 1)
	m.bitmap = ImageToBitmapWithOpacity(m.rawImage, m.cfg.Opacity, id)
	m.hasBitmap = true
}

// Draw renders the background wallpaper onto canvas according to FitMode and Opacity.
func (m *Manager) Draw(canvas *ui.Canvas, bounds ui.Rect, theme ui.Theme) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.cfg.Enabled || !m.hasBitmap {
		return
	}

	bmp := m.bitmap
	bw := float32(bmp.Width)
	bh := float32(bmp.Height)
	if bw <= 0 || bh <= 0 {
		return
	}

	switch m.cfg.FitMode {
	case FitCover:
		scaleX := bounds.Width / bw
		scaleY := bounds.Height / bh
		scale := float32(math.Max(float64(scaleX), float64(scaleY)))

		drawW := bw * scale
		drawH := bh * scale
		drawX := bounds.X + (bounds.Width-drawW)/2
		drawY := bounds.Y + (bounds.Height-drawH)/2

		canvas.DrawBitmap(bmp, ui.Rect{
			X:      drawX,
			Y:      drawY,
			Width:  drawW,
			Height: drawH,
		})

	case FitContain:
		scaleX := bounds.Width / bw
		scaleY := bounds.Height / bh
		scale := float32(math.Min(float64(scaleX), float64(scaleY)))

		drawW := bw * scale
		drawH := bh * scale
		drawX := bounds.X + (bounds.Width-drawW)/2
		drawY := bounds.Y + (bounds.Height-drawH)/2

		canvas.DrawBitmap(bmp, ui.Rect{
			X:      drawX,
			Y:      drawY,
			Width:  drawW,
			Height: drawH,
		})

	case FitStretch:
		canvas.DrawBitmap(bmp, bounds)

	case FitTile:
		stepX := bw
		stepY := bh
		for y := bounds.Y; y < bounds.Y+bounds.Height; y += stepY {
			for x := bounds.X; x < bounds.X+bounds.Width; x += stepX {
				canvas.DrawBitmap(bmp, ui.Rect{
					X:      x,
					Y:      y,
					Width:  stepX,
					Height: stepY,
				})
			}
		}

	case FitCenter:
		drawX := bounds.X + (bounds.Width-bw)/2
		drawY := bounds.Y + (bounds.Height-bh)/2
		canvas.DrawBitmap(bmp, ui.Rect{
			X:      drawX,
			Y:      drawY,
			Width:  bw,
			Height: bh,
		})

	default:
		canvas.DrawBitmap(bmp, bounds)
	}
}

// ImageToBitmapWithOpacity converts Go image.Image to premultiplied BGRA ui.Bitmap scaled by opacity.
func ImageToBitmapWithOpacity(img image.Image, opacity float32, id uint64) ui.Bitmap {
	if opacity < 0 {
		opacity = 0
	} else if opacity > 1 {
		opacity = 1
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	stride := w * 4
	pix := make([]byte, h*stride)

	// Fast path for *image.RGBA (avoids 2,000,000 interface calls per 1080p frame)
	if rgba, ok := img.(*image.RGBA); ok {
		op256 := uint32(opacity*256.0 + 0.5)
		srcPix := rgba.Pix
		srcStride := rgba.Stride
		for y := range h {
			srcRow := (b.Min.Y - rgba.Rect.Min.Y + y) * srcStride
			dstRow := y * stride
			for x := range w {
				sOff := srcRow + (b.Min.X-rgba.Rect.Min.X+x)*4
				dOff := dstRow + x*4
				r := uint32(srcPix[sOff+0])
				g := uint32(srcPix[sOff+1])
				bVal := uint32(srcPix[sOff+2])
				a := uint32(srcPix[sOff+3])

				pix[dOff+0] = byte((bVal * op256) >> 8)
				pix[dOff+1] = byte((g * op256) >> 8)
				pix[dOff+2] = byte((r * op256) >> 8)
				pix[dOff+3] = byte((a * op256) >> 8)
			}
		}
		return ui.Bitmap{
			Pix:     pix,
			Width:   w,
			Height:  h,
			Stride:  stride,
			ID:      id,
			Version: 1,
		}
	}

	for y := range h {
		for x := range w {
			c := img.At(b.Min.X+x, b.Min.Y+y)
			r, g, bVal, a := c.RGBA() // 0-0xffff premultiplied

			// Direct premultiplied alpha scaling
			rf := float32(r>>8) * opacity
			gf := float32(g>>8) * opacity
			bf := float32(bVal>>8) * opacity
			af := float32(a>>8) * opacity

			offset := y*stride + x*4
			pix[offset+0] = byte(bf)
			pix[offset+1] = byte(gf)
			pix[offset+2] = byte(rf)
			pix[offset+3] = byte(af)
		}
	}

	return ui.Bitmap{
		Pix:     pix,
		Width:   w,
		Height:  h,
		Stride:  stride,
		ID:      id,
		Version: 1,
	}
}

// ImageToBitmap converts Go image.Image to premultiplied BGRA ui.Bitmap at 100% opacity.
func ImageToBitmap(img image.Image, id uint64) ui.Bitmap {
	return ImageToBitmapWithOpacity(img, 1.0, id)
}

// CreateTestImage creates a test image for testing.
func CreateTestImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	return img
}

// EnsureSampleLocalFile returns an in-memory sample wallpaper URI without writing temporary files to disk.
func EnsureSampleLocalFile() (string, error) {
	return "@preset:sunset", nil
}

// SampleSlideshowURIs returns 4 built-in preset wallpaper URIs for zero-disk-IO carousel previewing.
func SampleSlideshowURIs() []string {
	return []string{
		"@preset:aurora",
		"@preset:forest",
		"@preset:sunset",
		"@preset:grid",
	}
}
