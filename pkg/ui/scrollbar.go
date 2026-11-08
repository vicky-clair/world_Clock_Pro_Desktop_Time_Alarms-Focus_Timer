package ui

import (
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawScrollbar 为可滚动区域 ScrollArea 绘制现代圆角轨道与自适应比例胶囊滑块。
// DrawScrollbar paints a modern, sleek vertical scrollbar track and pill thumb for a ScrollArea.
func DrawScrollbar(canvas *ui.Canvas, areaBox flex.Box, ss *flex.ScrollState, theme ui.Theme) {
	if ss == nil || ss.Content <= ss.Viewport || ss.Viewport <= 0 {
		return
	}
	r := areaBox.Rect()
	if r.Width <= 0 || r.Height <= 0 {
		return
	}

	trackW := float32(6)
	trackX := r.X + r.Width - trackW - 2
	trackY := r.Y + 4
	trackH := r.Height - 8
	if trackH <= 0 {
		return
	}

	// 绘制半透明圆角背景轨道
	canvas.Rounded(ui.Rect{
		X:      trackX,
		Y:      trackY,
		Width:  trackW,
		Height: trackH,
	}, 3, theme.Border.WithAlpha(0.18))

	// 根据视口高度与内容总高度计算自适应滑块高度（限制最小高度为 28px）
	thumbH := (ss.Viewport / ss.Content) * trackH
	if thumbH < 28 {
		thumbH = 28
	}
	if thumbH > trackH {
		thumbH = trackH
	}

	// 计算当前滚动偏移比例并映射为滑块的垂直坐标
	maxScroll := ss.Max()
	ratio := float32(0)
	if maxScroll > 0 {
		ratio = ss.Offset / maxScroll
	}
	if ratio < 0 {
		ratio = 0
	} else if ratio > 1 {
		ratio = 1
	}

	thumbY := trackY + ratio*(trackH-thumbH)

	// 绘制现代胶囊形圆角滑块
	canvas.Rounded(ui.Rect{
		X:      trackX,
		Y:      thumbY,
		Width:  trackW,
		Height: thumbH,
	}, 3, theme.MutedForeground.WithAlpha(0.60))
}
