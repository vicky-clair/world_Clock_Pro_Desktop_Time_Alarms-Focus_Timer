//go:build !windows

package clock

import "os"

// platformDetectOSRegion 非 Windows 平台下通过 POSIX 环境变量（LC_ALL / LANG）识别系统区域语言。
func platformDetectOSRegion() (geoCode string, localeName string, winTZName string) {
	localeName = os.Getenv("LC_ALL")
	if localeName == "" {
		localeName = os.Getenv("LANG")
	}
	return "", localeName, ""
}
