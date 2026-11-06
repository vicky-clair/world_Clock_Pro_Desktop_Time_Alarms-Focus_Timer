package wallpaper

import (
	"image"
	"image/color"
	"math"
)

// PresetID 标识内置的程序化艺术壁纸预设类型。
// PresetID identifies a built-in wallpaper preset.
type PresetID string

const (
	PresetGrid   PresetID = "grid"   // 几何高对比网格预设
	PresetAurora PresetID = "aurora" // 极光星空预设
	PresetForest PresetID = "forest" // 幽静森林树冠预设
	PresetSunset PresetID = "sunset" // 落日余晖群山预设
)

// GeneratePresetImage 根据指定预设 ID 实时生成高对比度程序化艺术壁纸图像（无需外部图片文件）。
// GeneratePresetImage generates a colorful, high-contrast procedural image for the given preset.
func GeneratePresetImage(preset PresetID) image.Image {
	switch preset {
	case PresetGrid:
		return generateGridImage()
	case PresetForest:
		return generateForestImage()
	case PresetSunset:
		return generateSunsetImage()
	case PresetAurora:
		fallthrough
	default:
		return generateAuroraImage()
	}
}

// generateGridImage generates a 160x160 high-contrast grid pattern.
// Makes Tile vs Cover vs Fit vs Center instantly and unmistakably recognizable.
func generateGridImage() *image.RGBA {
	w, h := 160, 160
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := range h {
		for x := range w {
			r, g, b := uint8(22), uint8(28), uint8(36)

			isBorder := x < 2 || x >= w-2 || y < 2 || y >= h-2
			isCrossX := (x >= 78 && x <= 82 && y >= 65 && y <= 95)
			isCrossY := (y >= 78 && y <= 82 && x >= 65 && x <= 95)
			isDiag := (x == y || x == (h-1-y)) && (x%10 < 5)

			if isBorder {
				r, g, b = 76, 201, 240 // Bright cyan border
			} else if isCrossX || isCrossY {
				r, g, b = 255, 190, 11 // Amber gold center cross
			} else if isDiag {
				r, g, b = 112, 130, 56 // Muted olive
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

// generateAuroraImage creates a 500x350 starry aurora sky.
func generateAuroraImage() *image.RGBA {
	w, h := 500, 350
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := range h {
		fy := float64(y) / float64(h)
		for x := range w {
			fx := float64(x) / float64(w)

			r := 15.0 + 20.0*fy
			g := 20.0 + 35.0*fy
			b := 35.0 + 50.0*fy

			wave := math.Sin(fx*6.0+fy*3.0)*0.5 + 0.5
			wave2 := math.Cos(fx*4.0-fy*2.0)*0.5 + 0.5
			aurora := wave * wave2 * (1.0 - math.Abs(fy-0.45)*2.2)
			if aurora > 0 {
				r += 40.0 * aurora
				g += 150.0 * aurora
				b += 130.0 * aurora
			}

			seed := (x*374761393 + y*668265263) ^ 0x5bf03635
			if (seed % 140) == 0 {
				r += 160
				g += 180
				b += 200
			}

			if r > 255 {
				r = 255
			}
			if g > 255 {
				g = 255
			}
			if b > 255 {
				b = 255
			}

			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255})
		}
	}
	return img
}

// generateForestImage creates a 500x350 deep emerald nature canopy.
func generateForestImage() *image.RGBA {
	w, h := 500, 350
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := range h {
		fy := float64(y) / float64(h)
		for x := range w {
			fx := float64(x) / float64(w)

			canopy := math.Sin(fx*8.0)*math.Cos(fy*8.0)*0.25 + 0.5
			sunbeam := math.Max(0, 1.0-math.Abs(fx-0.3-fy*0.4)*3.5)

			r := 20.0 + 30.0*canopy + 90.0*sunbeam
			g := 45.0 + 75.0*canopy + 110.0*sunbeam
			b := 30.0 + 35.0*canopy + 40.0*sunbeam

			if r > 255 {
				r = 255
			}
			if g > 255 {
				g = 255
			}
			if b > 255 {
				b = 255
			}

			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255})
		}
	}
	return img
}

// generateSunsetImage creates a 500x350 warm amber and purple mountain sunset.
func generateSunsetImage() *image.RGBA {
	w, h := 500, 350
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := range h {
		fy := float64(y) / float64(h)
		for x := range w {
			fx := float64(x) / float64(w)

			r := 40.0 + 190.0*fy
			g := 25.0 + 110.0*fy
			b := 60.0 + 30.0*(1.0-fy)

			mountain := 0.65 + 0.12*math.Sin(fx*9.0) + 0.08*math.Cos(fx*17.0)
			if fy > mountain {
				r, g, b = 20.0, 18.0, 26.0
			}

			if r > 255 {
				r = 255
			}
			if g > 255 {
				g = 255
			}
			if b > 255 {
				b = 255
			}

			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255})
		}
	}
	return img
}
