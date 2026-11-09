// Package main 是 小瑞综合时钟桌面 (Xiaoru Integrated Clock Desktop) 桌面应用程序的入口包。
// 负责初始化持久化配置、建立跨进程单实例互斥锁、创建 Viva GPU 原生窗口以及挂载系统托盘菜单。
package main

import (
	"log"
	"os"
	"path/filepath"

	"worldclock/pkg/storage"
	"worldclock/pkg/ui"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/app/singleinstance"
	vui "github.com/vivaui/viva/ui"
)

func main() {
	// 0. 规范工作目录：当被系统开机自启（如 Windows Run 注册表）拉起时，CWD 默认为 C:\Windows\System32。
	// 将工作目录切换至当前可执行文件所在目录，确保相对路径资源加载与本地文件操作安全可靠。
	if exe, err := os.Executable(); err == nil && exe != "" {
		_ = os.Chdir(filepath.Dir(exe))
	}

	// 1. 初始化 JSON 持久化存储与全局应用运行时状态
	store := storage.NewStore("")
	state := ui.NewAppState(store)
	if store.Get().AutoStart {
		// 若已开启开机自动启动，启动时同步刷新当前可执行文件绝对路径到系统注册表启动项
		_ = ui.SyncSystemAutostart(true)
	}

	// 2. 严禁程序多开：通过 Viva 的 singleinstance 建立命名互斥锁与进程间通信 (IPC)。
	// Disallow running multiple instances: if another copy is already running,
	// notify the primary instance to show/focus its window and exit cleanly.
	claim, primary, err := singleinstance.SingleInstance("com.xiaoru.clockdesktop.v1", os.Args[1:], func(args []string) {
		// 当用户尝试启动第二个实例时，唤醒当前主实例并恢复显示主窗口
		state.RequestShowWindow()
	})
	if err == nil && !primary {
		// 当前进程为重复启动的副实例，通知主实例后立即退出
		return
	}
	defer claim.Close()

	// 3. 构建 Viva 窗口配置与绘制管线
	var canvas vui.Canvas
	cfg := ui.NewAppConfig(state)

	// 4. 启动桌面应用程序事件循环、打开主窗口并注册系统托盘图标
	if err := app.RunApplication(func(application *app.Application) error {
		state.Application = application
		win, err := application.OpenWindow(cfg, &canvas)
		if err != nil {
			return err
		}
		state.Window = win
		state.SyncTopmost()

		// 响应系统级重新打开事件（如任务栏或 Dock 激活）
		application.OnReopen(func() {
			win.Show()
		})

		// 注册系统托盘图标与右键菜单（支持点击图标显示主窗口、从托盘彻底退出程序）
		_, _ = application.AddStatusItem(app.StatusItemConfig{
			Tooltip: "小瑞综合时钟桌面 (Xiaoru Integrated Clock Desktop)",
			Icon:    ui.BuildStatusIcon(),
			OnClick: func(app.StatusClick) {
				win.Show()
			},
			Menu: func() []app.MenuItem {
				return []app.MenuItem{
					{
						Title: state.T("tray.show"),
						OnSelect: func() {
							win.Show()
						},
					},
					{
						Separator: true,
					},
					{
						Title: state.T("tray.exit"),
						OnSelect: func() {
							// 标记为强制退出，跳过“最小化到后台”拦截，保存配置并退出进程
							state.ForceQuit = true
							state.Persist()
							win.Close()
							application.Quit()
						},
					},
				}
			},
		})
		return nil
	}); err != nil {
		log.Fatalf("failed to run Xiaoru Integrated Clock Desktop: %v", err)
	}
}
