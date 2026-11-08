//go:build !windows

package ui

import "github.com/vivaui/viva/ui"

// platformGetMonitorWorkArea 在非 Windows 平台返回默认屏幕工作区矩形。
func platformGetMonitorWorkArea(_ ui.Rect) ui.Rect {
	return DefaultScreenWorkArea()
}

// platformShrinkWindowToEdgeBar 在非 Windows 平台为空实现（界面层仍渲染精简状态条视图）。
func platformShrinkWindowToEdgeBar(_ string, _ ui.Rect) {}

// platformRestoreWindowFromEdgeBar 在非 Windows 平台为空实现（界面层恢复完整视图）。
func platformRestoreWindowFromEdgeBar(_ ui.Rect) {}

// platformSetWindowTopmost 在非 Windows 平台为空实现。
func platformSetWindowTopmost(_ bool) {}
