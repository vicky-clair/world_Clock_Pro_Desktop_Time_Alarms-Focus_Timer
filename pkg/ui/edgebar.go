package ui

import "github.com/vivaui/viva/ui"

const (
	// EdgeSnapThresholdTop 定义窗口内容区顶部靠近屏幕工作区顶部的触发阈值（含系统标题栏高度补偿）。
	EdgeSnapThresholdTop float32 = 42
	// EdgeSnapThresholdSide 定义窗口内容区靠近屏幕左、右、下边缘的触发阈值（逻辑像素）。
	EdgeSnapThresholdSide float32 = 20
	// EdgeReleaseMargin 定义窗口移出边缘防抖区域的回差阈值。
	EdgeReleaseMargin float32 = 56
)

// DefaultScreenWorkArea 返回在无显示器信息（如无头单元测试环境）下的默认参考屏幕工作区 (1920x1080)。
func DefaultScreenWorkArea() ui.Rect {
	return ui.Rect{X: 0, Y: 0, Width: 1920, Height: 1080}
}

// DetectScreenEdge 检测给定的窗口内容矩形 frame 是否靠近屏幕工作区 workArea 的边缘。
// 若靠近边缘则返回对应方向 ("top", "left", "right", "bottom")，否则返回 ""。
func DetectScreenEdge(frame ui.Rect, workArea ui.Rect) string {
	if frame.Width <= 0 || frame.Height <= 0 {
		return ""
	}
	if workArea.Width <= 0 || workArea.Height <= 0 {
		workArea = DefaultScreenWorkArea()
	}

	// 若窗口本身几乎占满整个屏幕（例如最大化或全屏尺寸），不视为边缘停靠小条触发
	if frame.Width >= workArea.Width-32 && frame.Height >= workArea.Height-64 {
		return ""
	}

	if frame.Y <= workArea.Y+EdgeSnapThresholdTop {
		return "top"
	}
	if frame.X <= workArea.X+EdgeSnapThresholdSide {
		return "left"
	}
	if frame.X+frame.Width >= workArea.X+workArea.Width-EdgeSnapThresholdSide {
		return "right"
	}
	if frame.Y+frame.Height >= workArea.Y+workArea.Height-EdgeSnapThresholdSide {
		return "bottom"
	}
	return ""
}

// IsClearlyInsideScreen 判断窗口是否已明显离开屏幕四边（用于解除“恢复正常”后的防抖锁定）。
func IsClearlyInsideScreen(frame ui.Rect, workArea ui.Rect) bool {
	if frame.Width <= 0 || frame.Height <= 0 {
		return false
	}
	if workArea.Width <= 0 || workArea.Height <= 0 {
		workArea = DefaultScreenWorkArea()
	}
	return frame.Y > workArea.Y+EdgeReleaseMargin &&
		frame.X > workArea.X+EdgeReleaseMargin &&
		frame.X+frame.Width < workArea.X+workArea.Width-EdgeReleaseMargin &&
		frame.Y+frame.Height < workArea.Y+workArea.Height-EdgeReleaseMargin
}
