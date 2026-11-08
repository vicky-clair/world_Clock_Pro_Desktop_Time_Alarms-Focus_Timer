package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/vivaui/viva/app"
)

const (
	// AutostartAppID 是注册到操作系统登录启动项（Windows Run 注册表 / macOS LaunchAgent / Linux XDG Autostart）的唯一标识符。
	AutostartAppID = "com.xiaoru.clockdesktop"
	// AutostartAppName 是操作系统启动项管理器中展示的应用名称。
	AutostartAppName = "Xiaoru Integrated Clock Desktop"
)

// isTestBinary 判断当前进程是否为 go test 编译出的测试二进制，避免单元测试污染宿主机的真实开机启动注册表。
func isTestBinary() bool {
	if len(os.Args) == 0 {
		return false
	}
	exe := strings.ToLower(filepath.Base(os.Args[0]))
	if strings.HasSuffix(exe, ".test") || strings.HasSuffix(exe, ".test.exe") {
		return true
	}
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}

// SyncSystemAutostart 将“开机自动启动”开关状态同步写入操作系统当前用户的登录自启配置。
// 在单元测试环境下会自动跳过系统注册表写入，保证测试隔离安全。
func SyncSystemAutostart(enabled bool) error {
	if isTestBinary() {
		return nil
	}
	return app.SetAutostart(app.Autostart{
		ID:   AutostartAppID,
		Name: AutostartAppName,
	}, enabled)
}
