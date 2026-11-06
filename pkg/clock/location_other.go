//go:build !windows

package clock

import "os"

func platformDetectOSRegion() (geoCode string, localeName string, winTZName string) {
	localeName = os.Getenv("LC_ALL")
	if localeName == "" {
		localeName = os.Getenv("LANG")
	}
	return "", localeName, ""
}
