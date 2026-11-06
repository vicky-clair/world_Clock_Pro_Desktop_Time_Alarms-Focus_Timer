package ui

import (
	"worldclock/pkg/storage"

	"github.com/vivaui/viva/ui"
)

// BuildTheme 根据用户选择的主题类型、壁纸开启状态、卡片毛玻璃透明度与全局字体配色构建 Viva ui.Theme。
// BuildTheme constructs a Viva ui.Theme based on the theme setting.
func BuildTheme(themeType storage.ThemeType, wallpaperEnabled bool, cardOpacity float32, uiFontColor string) ui.Theme {
	var t ui.Theme
	switch themeType {
	case storage.ThemeLight:
		t = ui.LightTheme()
		t.Card = ui.RGB24(0xFFFFFF)
		t.Foreground = ui.RGB24(0x212529)
		t.Primary = ui.RGB24(0x3B7A57) // Amazon / Olive green
		t.PrimaryForeground = ui.RGB24(0xFFFFFF)
		t.Border = ui.RGB24(0xDEE2E6)
		t.Muted = ui.RGB24(0xEDF2F7)
		t.MutedForeground = ui.RGB24(0x6C757D)
		t.Accent = ui.RGB24(0x4A7C59)
		t.Danger = ui.RGB24(0xDC3545)
		// Light theme

	case storage.ThemeDark:
		t = ui.DarkTheme()
		t.Background = ui.RGB24(0x121417)
		t.Card = ui.RGB24(0x1A1D24)
		t.Foreground = ui.RGB24(0xF1F3F5)
		t.Primary = ui.RGB24(0x3A86FF)
		t.PrimaryForeground = ui.RGB24(0xFFFFFF)
		t.Border = ui.RGB24(0x2E3440)
		t.Muted = ui.RGB24(0x242831)
		t.MutedForeground = ui.RGB24(0x8892B0)
		t.Accent = ui.RGB24(0x4CC9F0)
		t.Danger = ui.RGB24(0xEF476F)
		// Dark theme

	case storage.ThemeNord:
		t = ui.DarkTheme()
		t.Background = ui.RGB24(0x2E3440)
		t.Card = ui.RGB24(0x3B4252)
		t.Foreground = ui.RGB24(0xECEFF4)
		t.Primary = ui.RGB24(0x88C0D0)
		t.PrimaryForeground = ui.RGB24(0x2E3440)
		t.Border = ui.RGB24(0x4C566A)
		t.Muted = ui.RGB24(0x434C5E)
		t.MutedForeground = ui.RGB24(0xD8DEE9)
		t.Accent = ui.RGB24(0x81A1C1)
		t.Danger = ui.RGB24(0xBF616A)
		// Nord theme

	case storage.ThemeOlive:
		fallthrough
	default:
		// Olive signature theme
		t = ui.DarkTheme()
		t.Background = ui.RGB24(0x181F1C)
		t.Card = ui.RGB24(0x222C28)
		t.Foreground = ui.RGB24(0xEEF2EF)
		t.Primary = ui.RGB24(0x708238) // Classic Olive
		t.PrimaryForeground = ui.RGB24(0xFFFFFF)
		t.Border = ui.RGB24(0x314039)
		t.Muted = ui.RGB24(0x283530)
		t.MutedForeground = ui.RGB24(0x9EAFA6)
		t.Accent = ui.RGB24(0x8F9779)
		t.Danger = ui.RGB24(0xCF5C53)
		// Olive theme
	}

	if wallpaperEnabled {
		if cardOpacity <= 0 {
			cardOpacity = 0.55
		}
		// Adaptive frosted glass card tint based on font color
		if uiFontColor == "black" {
			t.Card = ui.RGB24(0xFFFFFF).WithAlpha(cardOpacity)
			t.Border = ui.RGB24(0xCCCCCC).WithAlpha(0.60)
		} else {
			t.Card = t.Card.WithAlpha(cardOpacity)
			t.Border = t.Border.WithAlpha(0.65)
		}
	}

	// Apply custom UI font color to foreground across all views
	switch uiFontColor {
	case "black":
		t.Foreground = ui.RGB24(0x111315)
		t.MutedForeground = ui.RGB24(0x4A5568)
		t.PrimaryForeground = ui.RGB24(0x111315)
	case "white":
		t.Foreground = ui.RGB24(0xF8F9FA)
		t.MutedForeground = ui.RGB24(0xA0AEC0)
		t.PrimaryForeground = ui.RGB24(0xF8F9FA)
	case "gold":
		t.Foreground = ui.RGB24(0xFFD166)
		t.MutedForeground = ui.RGB24(0xD4A340)
		t.PrimaryForeground = ui.RGB24(0xFFD166)
	case "cyan":
		t.Foreground = ui.RGB24(0x70D6FF)
		t.MutedForeground = ui.RGB24(0x4EA8CC)
		t.PrimaryForeground = ui.RGB24(0x70D6FF)
	case "purple":
		t.Foreground = ui.RGB24(0xE0AAFF)
		t.MutedForeground = ui.RGB24(0xAD7AC9)
		t.PrimaryForeground = ui.RGB24(0xE0AAFF)
	case "olive":
		t.Foreground = ui.RGB24(0xA3B18A)
		t.MutedForeground = ui.RGB24(0x6E7B58)
		t.PrimaryForeground = ui.RGB24(0xA3B18A)
	}

	return t
}
