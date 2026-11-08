//go:build windows

package ui

import (
	"sync"
	"syscall"
	"unsafe"

	"github.com/vivaui/viva/ui"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procGetActiveWindow  = user32.NewProc("GetActiveWindow")
	procFindWindowW      = user32.NewProc("FindWindowW")
	procIsWindow         = user32.NewProc("IsWindow")
	procGetWindowRect    = user32.NewProc("GetWindowRect")
	procSetWindowPos     = user32.NewProc("SetWindowPos")
	procMonitorFromPoint = user32.NewProc("MonitorFromPoint")
	procGetMonitorInfoW  = user32.NewProc("GetMonitorInfoW")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procGetDpiForWindow  = user32.NewProc("GetDpiForWindow")
	winRectMu            sync.Mutex
	cachedAppHWND        uintptr
	savedOuterWinRect    win32Rect
	hasSavedOuterWinRect bool
)

// win32Rect 对应 Win32 API 的 RECT 结构体（左上右下物理像素坐标）。
type win32Rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// win32MonitorInfo 对应 Win32 API 的 MONITORINFO 结构体，包含显示器完整区域与排除任务栏后的工作区。
type win32MonitorInfo struct {
	CbSize    uint32
	RcMonitor win32Rect
	RcWork    win32Rect
	DwFlags   uint32
}

const (
	monitorDefaultToNearest = 0x00000002  // 若点不在任何显示器内，返回距离最近的显示器句柄
	swpNoSize               = 0x0001      // 保持当前窗口宽高尺寸不变
	swpNoMove               = 0x0002      // 保持当前窗口屏幕坐标不变
	swpNoZOrder             = 0x0004      // 保持当前窗口 Z 序不变
	swpNoActivate           = 0x0010      // 调整尺寸/位置/置顶时不额外抢占焦点
	swpShowWindow           = 0x0040      // 确保窗口保持可见
	hwndTopmost             = ^uintptr(0) // HWND_TOPMOST (-1)：将窗口置于所有非置顶窗口之上并保持置顶
	hwndNoTopmost           = ^uintptr(1) // HWND_NOTOPMOST (-2)：取消窗口置顶状态
	smCxScreen              = 0           // 主显示器屏幕宽度索引
	smCyScreen              = 1           // 主显示器屏幕高度索引
)

// resolveAppHWND 获取并缓存当前应用程序主窗口的 Win32 句柄 (HWND)。
// 优先复用已验证有效的缓存句柄；若尚未缓存，则依次通过 GetActiveWindow 与窗口标题 FindWindowW 解析。
func resolveAppHWND() uintptr {
	winRectMu.Lock()
	cached := cachedAppHWND
	winRectMu.Unlock()

	if cached != 0 && procIsWindow.Find() == nil {
		if ok, _, _ := procIsWindow.Call(cached); ok != 0 {
			return cached
		}
	}

	if procGetActiveWindow.Find() == nil {
		if hwnd, _, _ := procGetActiveWindow.Call(); hwnd != 0 {
			winRectMu.Lock()
			cachedAppHWND = hwnd
			winRectMu.Unlock()
			return hwnd
		}
	}

	if procFindWindowW.Find() == nil {
		if titlePtr, err := syscall.UTF16PtrFromString("小瑞综合时钟桌面 | Xiaoru Integrated Clock Desktop"); err == nil {
			if hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr))); hwnd != 0 {
				winRectMu.Lock()
				cachedAppHWND = hwnd
				winRectMu.Unlock()
				return hwnd
			}
		}
	}
	return 0
}

// platformSetWindowTopmost 调用 Win32 SetWindowPos 将主窗口设置为系统级“总在最前 (HWND_TOPMOST)”或取消置顶“(HWND_NOTOPMOST)”。
func platformSetWindowTopmost(topmost bool) {
	if isTestBinary() || procSetWindowPos.Find() != nil {
		return
	}
	hwnd := resolveAppHWND()
	if hwnd == 0 {
		return
	}
	insertAfter := hwndNoTopmost
	if topmost {
		insertAfter = hwndTopmost
	}
	_, _, _ = procSetWindowPos.Call(
		hwnd,
		insertAfter,
		0,
		0,
		0,
		0,
		swpNoMove|swpNoSize|swpNoActivate,
	)
}

// platformGetMonitorWorkArea 获取当前窗口所在显示器的可用工作区矩形（排除系统任务栏后的屏幕坐标）。
func platformGetMonitorWorkArea(frame ui.Rect) ui.Rect {
	ptX := int32(frame.X + frame.Width/2)
	ptY := int32(frame.Y + frame.Height/2)
	ptPacked := uintptr(uint32(ptX)) | (uintptr(uint32(ptY)) << 32)

	if err := procMonitorFromPoint.Find(); err == nil && procGetMonitorInfoW.Find() == nil {
		hMon, _, _ := procMonitorFromPoint.Call(ptPacked, monitorDefaultToNearest)
		if hMon != 0 {
			var mi win32MonitorInfo
			mi.CbSize = uint32(unsafe.Sizeof(mi))
			ret, _, _ := procGetMonitorInfoW.Call(hMon, uintptr(unsafe.Pointer(&mi)))
			if ret != 0 {
				w := float32(mi.RcWork.Right - mi.RcWork.Left)
				h := float32(mi.RcWork.Bottom - mi.RcWork.Top)
				if w > 0 && h > 0 {
					return ui.Rect{
						X:      float32(mi.RcWork.Left),
						Y:      float32(mi.RcWork.Top),
						Width:  w,
						Height: h,
					}
				}
			}
		}
	}

	if err := procGetSystemMetrics.Find(); err == nil {
		cx, _, _ := procGetSystemMetrics.Call(smCxScreen)
		cy, _, _ := procGetSystemMetrics.Call(smCyScreen)
		if cx > 0 && cy > 0 {
			return ui.Rect{X: 0, Y: 0, Width: float32(cx), Height: float32(cy)}
		}
	}
	return DefaultScreenWorkArea()
}

// getWindowScaleFactor 查询指定窗口句柄的 DPI 缩放倍率（96 DPI 为 1.0 倍）。
func getWindowScaleFactor(hwnd uintptr) float32 {
	if hwnd != 0 && procGetDpiForWindow.Find() == nil {
		dpi, _, _ := procGetDpiForWindow.Call(hwnd)
		if dpi >= 96 && dpi <= 480 {
			return float32(dpi) / 96.0
		}
	}
	return 1.0
}

// platformShrinkWindowToEdgeBar 将当前主线程窗口缩小为贴合屏幕边缘的紧凑横条尺寸。
func platformShrinkWindowToEdgeBar(edge string, normalFrame ui.Rect) {
	if isTestBinary() {
		return
	}
	hwnd := resolveAppHWND()
	if hwnd == 0 {
		return
	}

	var cur win32Rect
	if procGetWindowRect.Find() == nil {
		ret, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&cur)))
		if ret != 0 && (cur.Bottom-cur.Top) > 220 {
			winRectMu.Lock()
			savedOuterWinRect = cur
			hasSavedOuterWinRect = true
			winRectMu.Unlock()
		}
	}

	scale := getWindowScaleFactor(hwnd)
	barW := int32(float32(560) * scale)
	barH := int32(float32(118) * scale)

	work := platformGetMonitorWorkArea(normalFrame)
	workLeft := int32(work.X)
	workTop := int32(work.Y)
	workRight := int32(work.X + work.Width)
	workBottom := int32(work.Y + work.Height)

	targetX := cur.Left
	targetY := cur.Top
	switch edge {
	case "top":
		targetY = workTop + 4
	case "bottom":
		targetY = workBottom - barH - 4
	case "left":
		targetX = workLeft + 4
	case "right":
		targetX = workRight - barW - 4
	default:
		targetY = workTop + 4
	}

	if targetX < workLeft+4 {
		targetX = workLeft + 4
	}
	if targetX+barW > workRight-4 {
		targetX = workRight - barW - 4
	}
	if targetY < workTop+4 {
		targetY = workTop + 4
	}
	if targetY+barH > workBottom-4 {
		targetY = workBottom - barH - 4
	}

	if procSetWindowPos.Find() == nil {
		_, _, _ = procSetWindowPos.Call(
			hwnd,
			0,
			uintptr(targetX),
			uintptr(targetY),
			uintptr(barW),
			uintptr(barH),
			swpNoZOrder|swpNoActivate|swpShowWindow,
		)
	}
}

// platformRestoreWindowFromEdgeBar 将状态条窗口恢复为进入状态条前的正常尺寸，并向屏幕内部偏移以防再次触发边缘吸附。
func platformRestoreWindowFromEdgeBar(savedFrame ui.Rect) {
	if isTestBinary() {
		return
	}
	hwnd := resolveAppHWND()
	if hwnd == 0 {
		return
	}

	scale := getWindowScaleFactor(hwnd)
	restoreW := int32(float32(856) * scale)
	restoreH := int32(float32(618) * scale)
	restoreX := int32(float32(120) * scale)
	restoreY := int32(float32(90) * scale)

	winRectMu.Lock()
	if hasSavedOuterWinRect {
		w := savedOuterWinRect.Right - savedOuterWinRect.Left
		h := savedOuterWinRect.Bottom - savedOuterWinRect.Top
		if w >= 400 && h >= 300 {
			restoreW = w
			restoreH = h
		}
		restoreX = savedOuterWinRect.Left
		restoreY = savedOuterWinRect.Top
	}
	winRectMu.Unlock()

	work := platformGetMonitorWorkArea(savedFrame)
	margin := int32(float32(64) * scale)
	workLeft := int32(work.X)
	workTop := int32(work.Y)
	workRight := int32(work.X + work.Width)
	workBottom := int32(work.Y + work.Height)

	if restoreX < workLeft+margin {
		restoreX = workLeft + margin
	}
	if restoreY < workTop+margin {
		restoreY = workTop + margin
	}
	if restoreX+restoreW > workRight-margin {
		restoreX = workRight - restoreW - margin
	}
	if restoreY+restoreH > workBottom-margin {
		restoreY = workBottom - restoreH - margin
	}

	if procSetWindowPos.Find() == nil {
		_, _, _ = procSetWindowPos.Call(
			hwnd,
			0,
			uintptr(restoreX),
			uintptr(restoreY),
			uintptr(restoreW),
			uintptr(restoreH),
			swpNoZOrder|swpNoActivate|swpShowWindow,
		)
	}
}
