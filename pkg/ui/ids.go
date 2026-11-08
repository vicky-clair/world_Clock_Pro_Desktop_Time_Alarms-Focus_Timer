// Package ui 实现小瑞综合时钟桌面 (Xiaoru Integrated Clock Desktop) 的 Viva GPU 原生声明式界面、路由、主题与交互控制。
package ui

import "github.com/vivaui/viva/ui"

// Tree 保存全局唯一的 Viva 控件身份标识树（ui.Tree），用于在跨帧绘制时稳定追踪焦点、悬停与点击命中。
// 遵循 Viva 框架规范：所有命中目标均通过具名 Feature 与稳定 ID 声明，绝不在绘制循环中手写数字编号。
var (
	Tree = ui.Tree{}

	// 功能分区（Features / Areas）
	featNav     = Tree.Feature("nav")      // 顶部导航栏区域
	featHeader  = Tree.Feature("header")   // 顶部闹钟提醒横幅区域
	featClock   = Tree.Feature("clock")    // 主时钟区域
	featWorld   = Tree.Feature("world")    // 世界时钟与选城面板区域
	featAlarms  = Tree.Feature("alarms")   // 闹钟列表与新建面板区域
	featTimers  = Tree.Feature("timers")   // 倒计时与秒表区域
	featPopout  = Tree.Feature("popout")   // 桌面悬浮小窗区域
	featTracker = Tree.Feature("tracker")  // 专注计时器（番茄钟）区域
	featSleep   = Tree.Feature("sleep")    // 睡眠与环境音区域
	featNews    = Tree.Feature("news")     // 历史上的今天资讯阅读区域
	featSetting = Tree.Feature("settings") // 设置与偏好区域
	featDialog  = Tree.Feature("dialog")   // 全局模态对话框区域
	featEdgeBar = Tree.Feature("edgebar")  // 屏幕边缘收缩状态条区域
	featPro     = Tree.Feature("pro")      // Pro 会员订阅与 7 天试用中心区域

	// 顶部全局导航标签页 ID (Navigation Tabs)
	TabClockID         = Tree.ID(featNav, "tab-clock")          // 主时钟标签页
	TabWorldID         = Tree.ID(featNav, "tab-world")          // 世界时钟标签页
	TabAlarmsID        = Tree.ID(featNav, "tab-alarms")         // 闹钟管理标签页
	TabTimersID        = Tree.ID(featNav, "tab-timers")         // 倒计时与秒表标签页
	TabTrackerID       = Tree.ID(featNav, "tab-tracker")        // 专注计时器标签页
	TabSleepID         = Tree.ID(featNav, "tab-sleep")          // 睡眠与白噪音标签页
	TabNewsID          = Tree.ID(featNav, "tab-news")           // 历史上的今天标签页
	TabSettingsID      = Tree.ID(featNav, "tab-settings")       // 设置与偏好标签页
	TabProID           = Tree.ID(featNav, "tab-pro")            // 顶部醒目 Pro 会员与 7 天试用入口按钮
	IDBtnHeaderTopmost = Tree.ID(featNav, "btn-header-topmost") // 顶部导航栏窗口置顶锁定开关按钮
	TabPopoutID        = Tree.ID(featNav, "tab-popout")         // 悬浮时钟切换按钮

	// 顶部闹钟响铃提醒横幅按钮 (Header Alert Controls)
	IDBtnSnooze  = Tree.ID(featHeader, "btn-snooze")  // 贪睡 (+5分钟) 按钮
	IDBtnDismiss = Tree.ID(featHeader, "btn-dismiss") // 关闭响铃按钮

	// 主时钟底部快捷控制栏按钮 (Main Clock Dock Controls)
	IDBtnToggle24H     = Tree.ID(featClock, "btn-toggle-24h")     // 12/24 小时制切换按钮
	IDBtnToggleSeconds = Tree.ID(featClock, "btn-toggle-seconds") // 秒数显示开关按钮
	IDBtnToggleStyle   = Tree.ID(featClock, "btn-toggle-style")   // 数字/指针表盘切换按钮
	IDBtnQuickPopout   = Tree.ID(featClock, "btn-quick-popout")   // 进入悬浮时钟模式按钮
	IDBtnQuickEdgeBar  = Tree.ID(featClock, "btn-quick-edgebar")  // 一键缩为边缘状态条按钮

	// 屏幕边缘收缩状态条按钮 (Edge Bar Controls)
	IDBtnEdgeBarTopmost = Tree.ID(featEdgeBar, "btn-edgebar-topmost") // 边缘状态条置顶锁定切换按钮
	IDBtnRestoreEdgeBar = Tree.ID(featEdgeBar, "btn-restore-edgebar") // 从状态条恢复正常窗口按钮

	// 世界时钟与 193 国目录选择弹窗控件 (World Clock Controls)
	IDBtnAddCityModal         = Tree.ID(featWorld, "btn-add-city-modal")    // 打开“添加城市”弹窗按钮
	IDBtnCloseAddCity         = Tree.ID(featWorld, "btn-close-add-city")    // 关闭“添加城市”弹窗按钮
	IDInputCitySearch         = Tree.ID(featWorld, "input-city-search")     // 城市/国家多语言搜索输入框
	IDBtnClearCitySearch      = Tree.ID(featWorld, "btn-clear-city-search") // 清空搜索词按钮
	IDScrollCityModal         = Tree.ID(featWorld, "scroll-city-modal")     // 193 国目录弹窗滚动容器
	IDScrollWorldList         = Tree.ID(featWorld, "scroll-world-list")     // 已添加世界时钟列表滚动容器
	IDDialogDupCityScope      = Tree.ID(featDialog, "dup-city-scope")       // 重复添加城市提醒弹窗作用域
	IDDialogDupCityBackdrop   = Tree.ID(featDialog, "dup-city-backdrop")    // 重复添加城市提醒弹窗遮罩层
	IDDialogDupCityClose      = Tree.ID(featDialog, "dup-city-close")       // 重复添加城市提醒弹窗右上角关闭按钮
	IDBtnDismissDuplicateCity = Tree.ID(featDialog, "btn-dismiss-dup-city") // 重复添加城市提醒弹窗“知道了”确认按钮

	// 选城弹窗常用国家快捷筛选按钮 (Quick Country Filter Buttons)
	IDQuickFilterCN = Tree.ID(featWorld, "qf-cn") // 快捷筛选：中国
	IDQuickFilterUS = Tree.ID(featWorld, "qf-us") // 快捷筛选：美国
	IDQuickFilterJP = Tree.ID(featWorld, "qf-jp") // 快捷筛选：日本
	IDQuickFilterUK = Tree.ID(featWorld, "qf-uk") // 快捷筛选：英国
	IDQuickFilterFR = Tree.ID(featWorld, "qf-fr") // 快捷筛选：法国
	IDQuickFilterDE = Tree.ID(featWorld, "qf-de") // 快捷筛选：德国

	// 选城弹窗六大洲分类筛选标签 (Continent Filter Tabs)
	IDContAll          = Tree.ID(featWorld, "cont-all")    // 全部大洲 (193 国)
	IDContAsia         = Tree.ID(featWorld, "cont-asia")   // 亚洲
	IDContEurope       = Tree.ID(featWorld, "cont-europe") // 欧洲
	IDContAfrica       = Tree.ID(featWorld, "cont-africa") // 非洲
	IDContNorthAmerica = Tree.ID(featWorld, "cont-na")     // 北美洲
	IDContSouthAmerica = Tree.ID(featWorld, "cont-sa")     // 南美洲
	IDContOceania      = Tree.ID(featWorld, "cont-oc")     // 大洋洲

	// 闹钟管理与新建闹钟面板控件 (Alarm Controls)
	IDBtnNewAlarmModal     = Tree.ID(featAlarms, "btn-new-alarm-modal") // 打开“新建闹钟”面板按钮
	IDBtnCloseNewAlarm     = Tree.ID(featAlarms, "btn-close-new-alarm") // 取消/关闭“新建闹钟”面板按钮
	IDBtnSaveNewAlarm      = Tree.ID(featAlarms, "btn-save-new-alarm")  // 保存新建闹钟按钮
	IDBtnPresetMorning     = Tree.ID(featAlarms, "preset-morning")      // 闹钟预设：早晨 07:30
	IDBtnPresetNoon        = Tree.ID(featAlarms, "preset-noon")         // 闹钟预设：午休 12:30
	IDBtnPresetEvening     = Tree.ID(featAlarms, "preset-evening")      // 闹钟预设：晚间 20:00
	IDBtnAlarmHourDown     = Tree.ID(featAlarms, "alarm-hr-down")       // 闹钟时间微调：-1 小时
	IDBtnAlarmHourUp       = Tree.ID(featAlarms, "alarm-hr-up")         // 闹钟时间微调：+1 小时
	IDBtnAlarmMinDown      = Tree.ID(featAlarms, "alarm-min-down")      // 闹钟时间微调：-5 分钟
	IDBtnAlarmMin1Down     = Tree.ID(featAlarms, "alarm-min1-down")     // 闹钟时间精准微调：-1 分钟（支持跨小时/跨午夜借位）
	IDBtnAlarmMin1Up       = Tree.ID(featAlarms, "alarm-min1-up")       // 闹钟时间精准微调：+1 分钟（支持跨小时/跨午夜进位）
	IDBtnAlarmMinUp        = Tree.ID(featAlarms, "alarm-min-up")        // 闹钟时间微调：+5 分钟
	IDBtnRecurOnce         = Tree.ID(featAlarms, "recur-once")          // 重复周期：仅一次
	IDBtnRecurDaily        = Tree.ID(featAlarms, "recur-daily")         // 重复周期：每天
	IDBtnRecurWeekdays     = Tree.ID(featAlarms, "recur-weekdays")      // 重复周期：工作日 (周一至周五)
	IDBtnAlarmSoundGentle  = Tree.ID(featAlarms, "alarm-snd-gentle")    // 闹钟铃声：柔和风铃
	IDBtnAlarmSoundDigital = Tree.ID(featAlarms, "alarm-snd-digital")   // 闹钟铃声：数字脉冲
	IDBtnAlarmSoundClassic = Tree.ID(featAlarms, "alarm-snd-classic")   // 闹钟铃声：经典钟声
	IDBtnAlarmPreviewSound = Tree.ID(featAlarms, "alarm-snd-preview")   // 试听/停止闹钟铃声按钮

	// 倒计时器与高精度秒表控件 (Timers & Stopwatch Controls)
	IDBtnAddTimerCustom     = Tree.ID(featTimers, "btn-add-timer-custom")  // 快捷添加 +10 分钟倒计时按钮
	IDBtnOpenTimerModal     = Tree.ID(featTimers, "btn-open-timer-modal")  // 打开“自定义任务”倒计时面板按钮
	IDBtnCloseTimerModal    = Tree.ID(featTimers, "btn-close-timer-modal") // 关闭自定义倒计时面板按钮
	IDBtnSaveTimerModal     = Tree.ID(featTimers, "btn-save-timer-modal")  // 保存自定义倒计时任务按钮
	IDInputTimerName        = Tree.ID(featTimers, "input-timer-name")      // 倒计时任务名称输入框
	IDBtnTimerPresetTea     = Tree.ID(featTimers, "timer-preset-tea")      // 倒计时模板：泡茶 3 分钟
	IDBtnTimerPresetNoodle  = Tree.ID(featTimers, "timer-preset-noodle")   // 倒计时模板：煮面 5 分钟
	IDBtnTimerPresetNap     = Tree.ID(featTimers, "timer-preset-nap")      // 倒计时模板：午休 20 分钟
	IDBtnTimerPresetPomo    = Tree.ID(featTimers, "timer-preset-pomo")     // 倒计时模板：番茄专注 25 分钟
	IDBtnTimerPresetRead    = Tree.ID(featTimers, "timer-preset-read")     // 倒计时模板：沉浸阅读 45 分钟
	IDBtnTimerPresetWorkout = Tree.ID(featTimers, "timer-preset-workout")  // 倒计时模板：运动健身 1 小时
	IDBtnTimerHrDown        = Tree.ID(featTimers, "timer-hr-down")         // 倒计时时长调节：-1 小时
	IDBtnTimerHrUp          = Tree.ID(featTimers, "timer-hr-up")           // 倒计时时长调节：+1 小时
	IDBtnTimerMin10Down     = Tree.ID(featTimers, "timer-min10-down")      // 倒计时时长调节：-10 分钟
	IDBtnTimerMinDown       = Tree.ID(featTimers, "timer-min-down")        // 倒计时时长精准调节：-1 分钟（支持跨小时借位）
	IDBtnTimerMinUp         = Tree.ID(featTimers, "timer-min-up")          // 倒计时时长精准调节：+1 分钟（支持跨小时进位）
	IDBtnTimerMin10Up       = Tree.ID(featTimers, "timer-min10-up")        // 倒计时时长调节：+10 分钟
	IDBtnTimerSecDown       = Tree.ID(featTimers, "timer-sec-down")        // 倒计时时长调节：-15 秒
	IDBtnTimerSec1Down      = Tree.ID(featTimers, "timer-sec1-down")       // 倒计时时长精准调节：-1 秒（支持跨分/跨时借位）
	IDBtnTimerSec1Up        = Tree.ID(featTimers, "timer-sec1-up")         // 倒计时时长精准调节：+1 秒（支持跨分/跨时进位）
	IDBtnTimerSecUp         = Tree.ID(featTimers, "timer-sec-up")          // 倒计时时长调节：+15 秒
	IDBtnSwStart            = Tree.ID(featTimers, "btn-sw-start")          // 秒表：开始计时
	IDBtnSwPause            = Tree.ID(featTimers, "btn-sw-pause")          // 秒表：暂停计时
	IDBtnSwReset            = Tree.ID(featTimers, "btn-sw-reset")          // 秒表：清零重置
	IDBtnSwLap              = Tree.ID(featTimers, "btn-sw-lap")            // 秒表：记录分圈 (Lap)

	// 生产力专注计时器（番茄工作法）控件 (Productivity Tracker Controls)
	IDBtnTrackerStart    = Tree.ID(featTracker, "btn-tracker-start")     // 开始专注工作阶段按钮
	IDBtnTrackerPause    = Tree.ID(featTracker, "btn-tracker-pause")     // 暂停当前番茄钟按钮
	IDBtnTrackerResume   = Tree.ID(featTracker, "btn-tracker-resume")    // 继续当前番茄钟按钮
	IDBtnTrackerReset    = Tree.ID(featTracker, "btn-tracker-reset")     // 重置当前番茄钟按钮
	IDBtnTrackerSkip     = Tree.ID(featTracker, "btn-tracker-skip")      // 跳过休息进入下一工作段按钮
	IDBtnTrackerPostpone = Tree.ID(featTracker, "btn-tracker-postpone")  // 推迟休息 (+5 分钟) 按钮
	IDBtnTrackerDND      = Tree.ID(featTracker, "btn-tracker-dnd")       // 专注勿扰模式开关按钮
	IDBtnTrackerMin5Down = Tree.ID(featTracker, "btn-tracker-min5-down") // 番茄工作时长调节：-5 分钟
	IDBtnTrackerMin1Down = Tree.ID(featTracker, "btn-tracker-min1-down") // 番茄工作时长精准调节：-1 分钟
	IDBtnTrackerMin1Up   = Tree.ID(featTracker, "btn-tracker-min1-up")   // 番茄工作时长精准调节：+1 分钟
	IDBtnTrackerMin5Up   = Tree.ID(featTracker, "btn-tracker-min5-up")   // 番茄工作时长调节：+5 分钟

	// 睡眠与环境音模式控件 (Sleep & Ambient Mode Controls)
	IDBtnSleepStart    = Tree.ID(featSleep, "btn-sleep-start")     // 启动睡眠模式与环境音伴眠按钮
	IDBtnSleepStop     = Tree.ID(featSleep, "btn-sleep-stop")      // 退出睡眠模式并停止音频按钮
	IDBtnSleep15m      = Tree.ID(featSleep, "btn-sleep-15m")       // 睡眠时长预设：15 分钟
	IDBtnSleep30m      = Tree.ID(featSleep, "btn-sleep-30m")       // 睡眠时长预设：30 分钟
	IDBtnSleep60m      = Tree.ID(featSleep, "btn-sleep-60m")       // 睡眠时长预设：60 分钟
	IDBtnSleepMin5Down = Tree.ID(featSleep, "btn-sleep-min5-down") // 睡眠时长调节：-5 分钟
	IDBtnSleepMin1Down = Tree.ID(featSleep, "btn-sleep-min1-down") // 睡眠时长精准调节：-1 分钟
	IDBtnSleepMin1Up   = Tree.ID(featSleep, "btn-sleep-min1-up")   // 睡眠时长精准调节：+1 分钟
	IDBtnSleepMin5Up   = Tree.ID(featSleep, "btn-sleep-min5-up")   // 睡眠时长调节：+5 分钟
	IDBtnSleepRain     = Tree.ID(featSleep, "btn-sleep-rain")      // 助眠环境音：柔和夜雨
	IDBtnSleepStream   = Tree.ID(featSleep, "btn-sleep-stream")    // 助眠环境音：林间溪流
	IDBtnSleepWaves    = Tree.ID(featSleep, "btn-sleep-waves")     // 助眠环境音：舒缓海浪
	IDBtnSleepWhite    = Tree.ID(featSleep, "btn-sleep-white")     // 助眠环境音：纯净白噪音
	IDBtnSleepPreview  = Tree.ID(featSleep, "btn-sleep-preview")   // 试听/停止当前助眠环境音按钮

	// 历史上的今天资讯控件 (On This Day in History Controls)
	IDNewsCatAll        = Tree.ID(featNews, "cat-all")             // 分类筛选：全部事件
	IDNewsCatTech       = Tree.ID(featNews, "cat-tech")            // 分类筛选：科技突破
	IDNewsCatWorld      = Tree.ID(featNews, "cat-world")           // 分类筛选：世界历史
	IDNewsCatScience    = Tree.ID(featNews, "cat-science")         // 分类筛选：科学探索
	IDBtnRefreshHistory = Tree.ID(featNews, "btn-refresh-history") // 联网刷新今日百科历史事件按钮

	// 桌面悬浮小窗控件 (Popout Clock Controls)
	IDBtnExitPopout      = Tree.ID(featPopout, "btn-exit-popout")       // 还原回主窗口按钮
	IDBtnPinPopout       = Tree.ID(featPopout, "btn-pin-popout")        // 悬浮窗置顶锁定开关按钮
	IDBtnPopoutPosTop    = Tree.ID(featPopout, "btn-popout-pos-top")    // 悬浮窗时钟对齐：靠顶
	IDBtnPopoutPosCenter = Tree.ID(featPopout, "btn-popout-pos-center") // 悬浮窗时钟对齐：居中
	IDBtnPopoutPosBottom = Tree.ID(featPopout, "btn-popout-pos-bottom") // 悬浮窗时钟对齐：靠底

	// 设置与偏好：视觉主题控件 (Theme Controls)
	IDThemeOlive = Tree.ID(featSetting, "theme-olive") // 经典橄榄绿主题
	IDThemeDark  = Tree.ID(featSetting, "theme-dark")  // 深邃暗黑主题
	IDThemeLight = Tree.ID(featSetting, "theme-light") // 明亮清爽主题
	IDThemeNord  = Tree.ID(featSetting, "theme-nord")  // 极地冰蓝 (Nord) 主题

	// 设置与偏好：时钟、温度单位、窗口置顶、开机自启与窗口行为控件 (Clock & Window Behavior Controls)
	IDOptTopmost                = Tree.ID(featSetting, "opt-topmost")                  // 窗口总在最前（置顶锁定）开关
	IDOptShowSeconds            = Tree.ID(featSetting, "opt-show-seconds")             // 是否显示秒数开关
	IDOptFormat24H              = Tree.ID(featSetting, "opt-format-24h")               // 12/24 小时制切换
	IDOptFullDate               = Tree.ID(featSetting, "opt-full-date")                // 完整/简短日期格式切换
	IDOptShowHolidays           = Tree.ID(featSetting, "opt-show-holidays")            // 是否显示节假日与二十四节气开关
	IDOptStyleDigital           = Tree.ID(featSetting, "opt-style-digital")            // 数字时钟样式按钮
	IDOptStyleAnalog            = Tree.ID(featSetting, "opt-style-analog")             // 模拟指针表盘样式按钮
	IDOptPopoutPosTop           = Tree.ID(featSetting, "opt-popout-pos-top")           // 设置页：悬浮时钟靠顶
	IDOptPopoutPosCenter        = Tree.ID(featSetting, "opt-popout-pos-center")        // 设置页：悬浮时钟居中
	IDOptPopoutPosBottom        = Tree.ID(featSetting, "opt-popout-pos-bottom")        // 设置页：悬浮时钟靠底
	IDOptTempFahrenheit         = Tree.ID(featSetting, "opt-temp-unit")                // 温度计量单位切换（摄氏度 °C / 华氏度 °F）
	IDCloseBehaviorMinimize     = Tree.ID(featSetting, "close-behavior-minimize")      // 点击关闭按钮（✕）时：最小化到系统托盘
	IDCloseBehaviorExit         = Tree.ID(featSetting, "close-behavior-exit")          // 点击关闭按钮（✕）时：直接退出程序
	IDOptAutoStart              = Tree.ID(featSetting, "opt-autostart")                // 开机自动启动状态徽章开关
	IDOptAutoStartEnable        = Tree.ID(featSetting, "opt-autostart-enable")         // 启用开机自动启动按钮
	IDOptAutoStartDisable       = Tree.ID(featSetting, "opt-autostart-disable")        // 不启用开机自动启动按钮
	IDOptEdgeDock               = Tree.ID(featSetting, "opt-edge-dock")                // 靠近屏幕边缘自动缩为状态条徽章开关
	IDOptEdgeDockEnable         = Tree.ID(featSetting, "opt-edge-dock-enable")         // 启用状态条收缩按钮
	IDOptEdgeDockDisable        = Tree.ID(featSetting, "opt-edge-dock-disable")        // 不启用状态条收缩按钮
	IDBtnSettingsEdgeBarPreview = Tree.ID(featSetting, "btn-settings-edgebar-preview") // 立即体验缩为边缘状态条按钮

	// 设置中心顶部五大子模块切换按钮 (Settings Sub-Tabs)
	IDSettingsTabLanguage  = Tree.ID(featSetting, "subtab-language")  // 🌐 语言（含中/英/日切换与字体大小/颜色定制）
	IDSettingsTabWallpaper = Tree.ID(featSetting, "subtab-wallpaper") // 🖼️ 壁纸（含本地壁纸轮播、预设与主题配色）
	IDSettingsTabClock     = Tree.ID(featSetting, "subtab-clock")     // ⏰ 时钟（含边缘状态条、开机自启、显示格式、温度单位与关闭行为）
	IDSettingsTabPro       = Tree.ID(featSetting, "subtab-pro")       // 👑 会员（含 7 天免费试用与 ¥3/月、¥28/年 订阅管理）
	IDSettingsTabAbout     = Tree.ID(featSetting, "subtab-about")     // ℹ️ 关于（含架构说明与版本号）

	// Pro 会员订阅与 7 天试用中心按钮 (Pro Subscription & Trial Center Controls)
	IDBtnCloseProModal  = Tree.ID(featPro, "btn-close-pro-modal")   // 关闭 Pro 订阅面板按钮
	IDBtnProSubMonthly  = Tree.ID(featPro, "btn-pro-sub-monthly")   // 开通按月订阅 (¥3/月) 按钮
	IDBtnProSubYearly   = Tree.ID(featPro, "btn-pro-sub-yearly")    // 开通按年订阅 (¥28/年) 按钮
	IDBtnProRestore     = Tree.ID(featPro, "btn-pro-restore")       // 恢复购买 / 刷新授权按钮
	IDBtnProManage      = Tree.ID(featPro, "btn-pro-manage")        // 管理或取消订阅 (Microsoft Store) 按钮
	IDBtnProSimTrial    = Tree.ID(featPro, "btn-pro-sim-trial")     // 测试切换：重置为 7 天免费试用态
	IDBtnProSimFullTest = Tree.ID(featPro, "btn-pro-sim-full-test") // 测试切换：切换为永久免付完整测试版

	// 界面语言切换按钮 (Language Controls)
	IDLangZH = Tree.ID(featSetting, "lang-zh") // 切换为简体中文
	IDLangEN = Tree.ID(featSetting, "lang-en") // 切换为英文 (English)
	IDLangJA = Tree.ID(featSetting, "lang-ja") // 切换为日文 (日本語)

	// 壁纸与轮播图控制按钮 (Wallpaper Controls)
	IDWallpaperToggle          = Tree.ID(featSetting, "wp-toggle")           // 启用/关闭自定义背景壁纸开关
	IDBtnBrowseWallpaper       = Tree.ID(featSetting, "wp-browse")           // 选择本地图片（支持多选）按钮
	IDBtnBrowseFolder          = Tree.ID(featSetting, "wp-browse-folder")    // 选择图片文件夹轮播按钮
	IDBtnSampleWallpaper       = Tree.ID(featSetting, "wp-sample")           // 加载示例高清壁纸按钮
	IDBtnSampleSlideshow       = Tree.ID(featSetting, "wp-sample-slideshow") // 加载 4 张示例轮播图集按钮
	IDBtnClearWallpaper        = Tree.ID(featSetting, "wp-clear")            // 清除壁纸列表按钮
	IDWallpaperSlideshowToggle = Tree.ID(featSetting, "wp-slideshow-toggle") // 自动轮播开关按钮
	IDWallpaperSlidePrev       = Tree.ID(featSetting, "wp-slide-prev")       // 上一张壁纸按钮
	IDWallpaperSlideNext       = Tree.ID(featSetting, "wp-slide-next")       // 下一张壁纸按钮
	IDWallpaperInt1m           = Tree.ID(featSetting, "wp-int-1m")           // 轮播间隔：1 分钟
	IDWallpaperInt5m           = Tree.ID(featSetting, "wp-int-5m")           // 轮播间隔：5 分钟
	IDWallpaperInt10m          = Tree.ID(featSetting, "wp-int-10m")          // 轮播间隔：10 分钟
	IDWallpaperInt15m          = Tree.ID(featSetting, "wp-int-15m")          // 轮播间隔：15 分钟
	IDWallpaperOrderShuffle    = Tree.ID(featSetting, "wp-order-shuffle")    // 轮播顺序：随机不重复洗牌
	IDWallpaperOrderSeq        = Tree.ID(featSetting, "wp-order-seq")        // 轮播顺序：按文件名顺序
	IDPresetGrid               = Tree.ID(featSetting, "wp-preset-grid")      // 内置艺术壁纸：赛博网格
	IDPresetAurora             = Tree.ID(featSetting, "wp-preset-aurora")    // 内置艺术壁纸：极光星夜
	IDPresetForest             = Tree.ID(featSetting, "wp-preset-forest")    // 内置艺术壁纸：迷雾森林
	IDPresetSunset             = Tree.ID(featSetting, "wp-preset-sunset")    // 内置艺术壁纸：暮色地平线
	IDWallpaperFitFill         = Tree.ID(featSetting, "wp-fit-fill")         // 画面缩放：填充 (Fill)
	IDWallpaperFitFit          = Tree.ID(featSetting, "wp-fit-fit")          // 画面缩放：适应 (Fit)
	IDWallpaperFitStretch      = Tree.ID(featSetting, "wp-fit-stretch")      // 画面缩放：拉伸 (Stretch)
	IDWallpaperFitTile         = Tree.ID(featSetting, "wp-fit-tile")         // 画面缩放：平铺 (Tile)
	IDWallpaperFitCenter       = Tree.ID(featSetting, "wp-fit-center")       // 画面缩放：居中 (Center)
	IDWallpaperOpacityDown     = Tree.ID(featSetting, "wp-opacity-down")     // 降低背景图片不透明度 (-10%)
	IDWallpaperOpacityUp       = Tree.ID(featSetting, "wp-opacity-up")       // 提高背景图片不透明度 (+10%)
	IDCardGlassDown            = Tree.ID(featSetting, "card-glass-down")     // 降低前景卡片毛玻璃浓度 (-10%)
	IDCardGlassUp              = Tree.ID(featSetting, "card-glass-up")       // 提高前景卡片毛玻璃浓度 (+10%)

	// 滚动区域容器 ID (ScrollArea IDs)
	IDScrollSettings = Tree.ID(featSetting, "scroll-settings") // 设置页滚动容器
	IDScrollAlarms   = Tree.ID(featAlarms, "scroll-alarms")    // 闹钟列表滚动容器
	IDScrollTimers   = Tree.ID(featTimers, "scroll-timers")    // 倒计时与秒表页滚动容器
	IDScrollNews     = Tree.ID(featNews, "scroll-news")        // 历史上的今天页面滚动容器
	IDScrollTracker  = Tree.ID(featTracker, "scroll-tracker")  // 专注计时页面滚动容器

	// 字体大小、系统字体选择、时钟数字颜色、界面文字颜色与全局缩放按钮 (Font & Color Customization Controls)
	IDBtnFontDefault          = Tree.ID(featSetting, "font-default")         // 恢复系统默认字体按钮
	IDBtnToggleSystemFontList = Tree.ID(featSetting, "btn-toggle-sys-fonts") // 展开/收起系统完整字体库列表按钮
	IDBtnSysFontPagePrev      = Tree.ID(featSetting, "btn-sys-font-prev")    // 系统字体库列表上一页按钮
	IDBtnSysFontPageNext      = Tree.ID(featSetting, "btn-sys-font-next")    // 系统字体库列表下一页按钮
	IDFontSize48              = Tree.ID(featSetting, "fs-48")                // 时钟字号预设：紧凑 48px
	IDFontSize64              = Tree.ID(featSetting, "fs-64")                // 时钟字号预设：标准 64px
	IDFontSize76              = Tree.ID(featSetting, "fs-76")                // 时钟字号预设：大号 76px
	IDFontSize88              = Tree.ID(featSetting, "fs-88")                // 时钟字号预设：特大 88px
	IDFontSize96              = Tree.ID(featSetting, "fs-96")                // 时钟字号预设：超大 96px
	IDFontSizeCustomDown      = Tree.ID(featSetting, "fs-custom-down")       // 时钟字号步进微调：-2px
	IDFontSizeCustomUp        = Tree.ID(featSetting, "fs-custom-up")         // 时钟字号步进微调：+2px

	IDFontColorWhite  = Tree.ID(featSetting, "fc-white")  // 时钟数字颜色：纯白
	IDFontColorGold   = Tree.ID(featSetting, "fc-gold")   // 时钟数字颜色：琥珀金
	IDFontColorOlive  = Tree.ID(featSetting, "fc-olive")  // 时钟数字颜色：橄榄翠绿
	IDFontColorCyan   = Tree.ID(featSetting, "fc-cyan")   // 时钟数字颜色：冰川青
	IDFontColorPurple = Tree.ID(featSetting, "fc-purple") // 时钟数字颜色：霓虹紫
	IDFontColorCoral  = Tree.ID(featSetting, "fc-coral")  // 时钟数字颜色：珊瑚橙红

	IDUIColorDefault = Tree.ID(featSetting, "uic-default") // 全局界面文字颜色：跟随主题默认
	IDUIColorWhite   = Tree.ID(featSetting, "uic-white")   // 全局界面文字颜色：纯白（适合深色壁纸）
	IDUIColorBlack   = Tree.ID(featSetting, "uic-black")   // 全局界面文字颜色：深黑（适合浅色壁纸）
	IDUIColorGold    = Tree.ID(featSetting, "uic-gold")    // 全局界面文字颜色：暖金
	IDUIColorCyan    = Tree.ID(featSetting, "uic-cyan")    // 全局界面文字颜色：亮青
	IDUIColorPurple  = Tree.ID(featSetting, "uic-purple")  // 全局界面文字颜色：柔紫
	IDUIColorOlive   = Tree.ID(featSetting, "uic-olive")   // 全局界面文字颜色：翠绿

	IDFontScale90  = Tree.ID(featSetting, "fscale-90")  // 全局字号缩放：90%
	IDFontScale100 = Tree.ID(featSetting, "fscale-100") // 全局字号缩放：100%（默认）
	IDFontScale110 = Tree.ID(featSetting, "fscale-110") // 全局字号缩放：110%
	IDFontScale120 = Tree.ID(featSetting, "fscale-120") // 全局字号缩放：120%
)

// CityToggleFavID 返回世界时钟城市列表中指定城市的收藏星标按钮 ID。
func CityToggleFavID(cityID string) ui.ID {
	return Tree.ID(featWorld, "fav-"+cityID)
}

// CityRemoveID 返回世界时钟城市列表中指定城市的删除按钮 ID。
func CityRemoveID(cityID string) ui.ID {
	return Tree.ID(featWorld, "rm-"+cityID)
}

// CityMoveUpID 返回世界时钟城市列表中指定城市的上移排序按钮 ID。
func CityMoveUpID(cityID string) ui.ID {
	return Tree.ID(featWorld, "up-"+cityID)
}

// CityMoveDownID 返回世界时钟城市列表中指定城市的下移排序按钮 ID。
func CityMoveDownID(cityID string) ui.ID {
	return Tree.ID(featWorld, "down-"+cityID)
}

// CityAddCatalogID 返回全球 193 国目录中指定城市的添加按钮 ID。
func CityAddCatalogID(cityID string) ui.ID {
	return Tree.ID(featWorld, "add-cat-"+cityID)
}

// AlarmToggleID 返回指定闹钟的启用/停用开关按钮 ID。
func AlarmToggleID(alarmID string) ui.ID {
	return Tree.ID(featAlarms, "tog-"+alarmID)
}

// AlarmDeleteID 返回指定闹钟的删除按钮 ID。
func AlarmDeleteID(alarmID string) ui.ID {
	return Tree.ID(featAlarms, "del-"+alarmID)
}

// AlarmPreviewID 返回指定闹钟的铃声试听按钮 ID。
func AlarmPreviewID(alarmID string) ui.ID {
	return Tree.ID(featAlarms, "prev-"+alarmID)
}

// TimerStartID 返回指定倒计时任务的开始/继续按钮 ID。
func TimerStartID(timerID string) ui.ID {
	return Tree.ID(featTimers, "start-"+timerID)
}

// TimerPauseID 返回指定倒计时任务的暂停按钮 ID。
func TimerPauseID(timerID string) ui.ID {
	return Tree.ID(featTimers, "pause-"+timerID)
}

// TimerResetID 返回指定倒计时任务的重置按钮 ID。
func TimerResetID(timerID string) ui.ID {
	return Tree.ID(featTimers, "reset-"+timerID)
}

// TimerDeleteID 返回指定倒计时任务的删除按钮 ID。
func TimerDeleteID(timerID string) ui.ID {
	return Tree.ID(featTimers, "del-"+timerID)
}

// TimerEditID 返回指定倒计时任务的编辑修改按钮 ID。
func TimerEditID(timerID string) ui.ID {
	return Tree.ID(featTimers, "edit-"+timerID)
}

// CuratedFontID 返回“语言与字体”设置页中精选系统字体快捷按钮的稳定 ID。
func CuratedFontID(family string) ui.ID {
	if family == "" {
		return IDBtnFontDefault
	}
	return Tree.ID(featSetting, "curated-font-"+family)
}

// SystemFontItemID 返回“语言与字体”设置页中系统完整字体库列表项按钮的稳定 ID。
func SystemFontItemID(family string) ui.ID {
	return Tree.ID(featSetting, "sys-font-"+family)
}
