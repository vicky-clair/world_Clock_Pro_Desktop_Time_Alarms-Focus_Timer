// Package ui 实现基于 Viva GPU 原生框架的声明式桌面界面、帧重绘调度、主题配色与交互事件路由。
package ui

import (
	"os"
	"time"

	"worldclock/pkg/clock"
	"worldclock/pkg/sleep"
	"worldclock/pkg/storage"
	"worldclock/pkg/timer"
	"worldclock/pkg/wallpaper"

	"github.com/vivaui/viva/app"
	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// NewAppConfig 构建 World Clock Pro 应用的 Viva 窗口配置、绘制回调与生命周期钩子。
// NewAppConfig creates the Viva app configuration for the World Clock application.
func NewAppConfig(state *AppState) app.Config {
	var inv app.Invalidator
	state.Invalidator = &inv

	cfg := app.Config{
		Tree:        &Tree,
		Title:       "World Clock Pro: Desktop Time, Alarms & Focus Timer",
		Width:       840,
		Height:      580,
		Invalidator: &inv,
		TextInput:   flex.Fields(&state.CitySearchField),
	}

	// Paint 是核心声明式绘制函数：每帧根据当前状态声明界面布局，并按需规划下一次唤醒时间。
	cfg.Paint = func(canvas *ui.Canvas, size ui.Size) {
		// 1. 记录当前帧统一时间戳，保证同一帧内所有时钟读数完全一致
		now := time.Now()
		state.Now = now
		state.UTC = now.UTC()

		// 2. 检查壁纸轮播是否到达切换时间点（未到期时为 O(1) 零内存分配检查）
		if state.Wallpaper.AdvanceIfDue(now) {
			state.Persist()
		}

		settings := state.Store.Get()
		theme := BuildTheme(settings.Theme, state.Wallpaper.Config().Enabled, state.Wallpaper.Config().CardOpacity, settings.UIFontColor)

		// 3. 检查当前分钟是否有闹钟应当触发响铃（若处于贪睡静音期内则返回 nil）
		ringing := state.Scheduler.CheckTriggers(now)

		// 4. 无论当前处于哪个标签页，同步推进后台专注计时器、睡眠模式和倒计时器进度
		state.Tracker.Update(now)
		slState, slVol := state.Sleep.Update(now)
		_, slSound, _ := state.Sleep.Info()
		timerRinging := false
		for _, t := range state.TimerMgr.Timers() {
			t.Update(now)
			if t.State == timer.StateExpired {
				timerRinging = true
			}
		}

		// 4b. 同步闹钟响铃、倒计时到期提醒与睡眠环境白噪音的硬件音频输出状态
		alarmActive := ringing != nil || timerRinging
		alarmSound := "Classic Chime"
		if ringing != nil && ringing.Sound != "" {
			alarmSound = ringing.Sound
		} else if timerRinging {
			alarmSound = "Digital Bell"
		}
		sleepPlaying := slState == sleep.StateActive || slState == sleep.StateFading
		if state.Audio != nil {
			state.Audio.SyncState(alarmActive, alarmSound, sleepPlaying, string(slSound), slVol)
		}

		// 4c. 后台非阻塞同步“历史上的今天”在线 API（每天每种语言仅请求一次，完成后自动触发单次重绘）
		if state.News != nil {
			state.News.EnsureFetchedAsync(now, string(state.Lang()), false, state.Invalidate)
		}

		// 5. 判断是否存在需要高频刷新（约 25 FPS）的活动秒表、倒计时或番茄钟
		fastTick := false
		if state.Stopwatch.State() == timer.StateRunning {
			fastTick = true
		}
		for _, t := range state.TimerMgr.Timers() {
			if t.State == timer.StateRunning {
				fastTick = true
				break
			}
		}
		trStatus := state.Tracker.Status()
		if trStatus == "Working" || trStatus == "Short Break" || trStatus == "Big Break" {
			fastTick = true
		}

		// 6. 严格遵守 Viva 框架“空闲零重绘”准则：
		// 绝不运行后台忙碌轮询！仅在需要更新画面的精确时刻预约下一帧重绘：
		if fastTick {
			canvas.RequestFrameIn(40 * time.Millisecond) // 秒表/倒计时运行期间按 40ms (~25 fps) 刷新
		} else {
			delay := clock.NextBoundaryDelay(now, settings.ShowSeconds || sleepPlaying)
			if wpDelay := state.Wallpaper.NextSlideDelay(now); wpDelay > 0 && wpDelay < delay {
				delay = wpDelay
			}
			canvas.RequestFrameIn(delay)
		}

		// 7. 绘制窗口根背景色与自定义壁纸图层
		bounds := ui.Rect{Width: size.Width, Height: size.Height}
		canvas.Fill(bounds, theme.Background)
		if state.Wallpaper.Config().Enabled {
			state.Wallpaper.Draw(canvas, bounds, theme)
		}

		// 8. 绘制声明式 Flex UI 组件树
		state.View.Draw(canvas, theme, bounds, func(f *flex.Frame) {
			if state.IsPopout {
				DrawPopoutClock(f, state, theme)
				return
			}

			f.Column().Fill().Children(func() {
				// 顶部全局导航栏与响铃提醒横幅
				DrawHeader(f, state, theme)

				// 当前激活的功能标签页视图
				f.Column().Fill().Children(func() {
					switch state.ActiveTab {
					case TabClock:
						DrawMainClock(f, state, theme)
					case TabWorld:
						DrawWorldClock(f, state, theme)
					case TabAlarms:
						DrawAlarms(f, state, theme)
					case TabTimers:
						DrawTimers(f, state, theme)
					case TabTracker:
						DrawTracker(f, state, theme)
					case TabSleep:
						DrawSleepMode(f, state, theme)
					case TabNews:
						DrawNews(f, state, theme)
					case TabSettings:
						DrawSettings(f, state, theme)
					case TabPopout:
						DrawPopoutClock(f, state, theme)
					default:
						DrawMainClock(f, state, theme)
					}
				})
			})
		})
	}

	// 鼠标滚轮事件转发给 Flex 滚动容器处理
	cfg.OnScroll = func(event ui.ScrollEvent) ui.EventResult {
		return state.View.Wheel(event)
	}

	// 跨线程唤醒回调：在 UI 主线程安全消费待处理的窗口恢复或壁纸文件加载请求
	cfg.OnWake = func() {
		state.ProcessPendingWake()
	}

	// 窗口关闭按钮（✕）拦截回调：支持“最小化到后台系统托盘”或“直接退出程序”
	cfg.OnCloseRequest = func(allow func()) {
		state.Persist()
		if state.ForceQuit || state.Store.Get().CloseBehavior == storage.CloseBehaviorExit {
			allow()
			if state.Application != nil {
				state.Application.Quit()
			}
			return
		}
		if state.Window != nil {
			state.Window.Hide()
			return
		}
		allow()
	}

	// 窗口销毁前兜底保存最新状态并停止音频输出
	cfg.OnClosed = func() {
		if state.Audio != nil {
			state.Audio.Stop()
		}
		state.Persist()
	}

	// 支持将本地图片文件或整个图片文件夹直接拖拽进窗口设置为壁纸/轮播图
	cfg.OnFileDrop = func(event ui.FileDropEvent) ui.EventResult {
		if event.Phase == ui.FileDropDrop && len(event.Paths) > 0 {
			first := event.Paths[0]
			if info, err := os.Stat(first); err == nil && info.IsDir() {
				if count, err := state.Wallpaper.LoadFolder(first); err == nil && count > 0 {
					state.Persist()
					return ui.EventResult{Handled: true, Repaint: true}
				}
			}
			if len(event.Paths) > 1 {
				if state.Wallpaper.LoadImages(event.Paths) > 0 {
					state.Persist()
					return ui.EventResult{Handled: true, Repaint: true}
				}
			}
			if wallpaper.IsSupportedImageExt(first) {
				_ = state.Wallpaper.LoadImage(first)
				state.Persist()
				return ui.EventResult{Handled: true, Repaint: true}
			}
		}
		if event.Phase == ui.FileDropOver {
			return ui.EventResult{Handled: true}
		}
		return ui.EventResult{}
	}

	return cfg
}

// BuildStatusIcon 生成系统托盘使用的 32x32 像素高清时钟图标（RGBA 格式）。
// BuildStatusIcon generates a crisp 32x32 straight-RGBA clock icon for the system tray.
func BuildStatusIcon() app.StatusIcon {
	const sz = 32
	rgba := make([]byte, sz*sz*4)
	cx, cy := 15.5, 15.5
	rOuter := 13.5
	rInner := 11.0
	for y := range sz {
		for x := range sz {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist2 := dx*dx + dy*dy
			off := (y*sz + x) * 4
			if dist2 <= rOuter*rOuter {
				if dist2 >= rInner*rInner {
					// Olive-gold rim (#8F9779)
					rgba[off+0] = 0x8F
					rgba[off+1] = 0x97
					rgba[off+2] = 0x79
					rgba[off+3] = 0xFF
				} else {
					// Dark clock face (#181F1C)
					rgba[off+0] = 0x18
					rgba[off+1] = 0x1F
					rgba[off+2] = 0x1C
					rgba[off+3] = 0xFF
				}
			}
		}
	}
	// Draw 12-o'clock vertical hand and 3-o'clock horizontal hand in bright white
	for y := 7; y <= 16; y++ {
		for x := 15; x <= 16; x++ {
			off := (y*sz + x) * 4
			rgba[off+0], rgba[off+1], rgba[off+2], rgba[off+3] = 0xFF, 0xFF, 0xFF, 0xFF
		}
	}
	for x := 15; x <= 22; x++ {
		for y := 15; y <= 16; y++ {
			off := (y*sz + x) * 4
			rgba[off+0], rgba[off+1], rgba[off+2], rgba[off+3] = 0xFF, 0xBE, 0x0B, 0xFF
		}
	}
	return app.StatusIcon{
		Size: sz,
		RGBA: rgba,
	}
}
