//go:build windows

package clock

import (
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procGetUserDefaultGeoName    = kernel32.NewProc("GetUserDefaultGeoName")
	procGetUserDefaultLocaleName = kernel32.NewProc("GetUserDefaultLocaleName")
	procGetTimeZoneInformation   = kernel32.NewProc("GetTimeZoneInformation")
)

// timeZoneInformation 对应 Win32 TIME_ZONE_INFORMATION 结构体。
type timeZoneInformation struct {
	Bias         int32
	StandardName [32]uint16
	StandardDate [8]uint16
	StandardBias int32
	DaylightName [32]uint16
	DaylightDate [8]uint16
	DaylightBias int32
}

// platformDetectOSRegion 调用 Windows 原生 Kernel32 API 获取系统设置的“国家或地区”代码、区域语言及系统时区名称。
func platformDetectOSRegion() (geoCode string, localeName string, winTZName string) {
	// 1. 读取 Windows 10/11 系统设置 -> 时间和语言 -> 区域 -> “国家或地区” (ISO-3166 两字母代码，如 "CN", "US", "JP")
	var geoBuf [16]uint16
	if err := procGetUserDefaultGeoName.Find(); err == nil {
		ret, _, _ := procGetUserDefaultGeoName.Call(uintptr(unsafe.Pointer(&geoBuf[0])), uintptr(len(geoBuf)))
		if ret > 0 {
			geoCode = strings.TrimSpace(syscall.UTF16ToString(geoBuf[:]))
		}
	}

	// 2. 读取用户默认区域格式名称 (如 "zh-CN", "en-US", "ja-JP")
	var locBuf [85]uint16
	if err := procGetUserDefaultLocaleName.Find(); err == nil {
		ret, _, _ := procGetUserDefaultLocaleName.Call(uintptr(unsafe.Pointer(&locBuf[0])), uintptr(len(locBuf)))
		if ret > 0 {
			localeName = strings.TrimSpace(syscall.UTF16ToString(locBuf[:]))
		}
	}

	// 3. 读取 Windows 系统时区标准名称 (如 "China Standard Time", "Tokyo Standard Time")
	var tzi timeZoneInformation
	if err := procGetTimeZoneInformation.Find(); err == nil {
		ret, _, _ := procGetTimeZoneInformation.Call(uintptr(unsafe.Pointer(&tzi)))
		if ret != 0xFFFFFFFF {
			winTZName = strings.TrimSpace(syscall.UTF16ToString(tzi.StandardName[:]))
		}
	}

	return geoCode, localeName, winTZName
}
