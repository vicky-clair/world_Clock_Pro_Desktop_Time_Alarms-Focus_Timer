//go:build !windows

package ui

import "github.com/vivaui/viva/app"

// syncSystemAutostartPlatform 在非 Windows 平台（macOS LaunchAgent / Linux XDG Autostart）继续交由 Viva 标准库处理。
func syncSystemAutostartPlatform(enabled bool) error {
	return app.SetAutostart(app.Autostart{
		ID:   AutostartAppID,
		Name: AutostartAppName,
	}, enabled)
}
