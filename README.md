<div align="center">

# 🌍 小瑞综合时钟桌面 | Xiaoru Integrated Clock Desktop | 小瑞総合時計デスクトップ

**Go-First, GPU-Native (Direct3D 11 / DirectWrite) Desktop Clock, Global Timezone Suite, Smart Alarms & Pomodoro Focus Station**

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![UI Framework](https://img.shields.io/badge/UI-Viva%20v0.2.4%20(GPU%20Native)-4E9F3D?style=flat-square)](https://github.com/vivaui/viva)
[![Platform](https://img.shields.io/badge/Platform-Windows%2010%2F11%20(x64)-0078D6?style=flat-square&logo=windows)](https://www.microsoft.com/windows)
[![Idle CPU](https://img.shields.io/badge/Idle%20CPU-0.0%25%20(Zero%20Polling)-success?style=flat-square)](#)
[![Languages](https://img.shields.io/badge/i18n-简体中文%20%7C%20English%20%7C%20日本語-orange?style=flat-square)](#)

<p align="center">
  <a href="#-简体中文文档-chinese"><b>🇨🇳 简体中文</b></a> •
  <a href="#-english-documentation"><b>🇺🇸 English</b></a> •
  <a href="#-日本語ドキュメント-japanese"><b>🇯🇵 日本語</b></a>
</p>

</div>

---

# 🇨🇳 简体中文文档 (Chinese)

## ✨ 项目简介

**小瑞综合时钟桌面（Xiaoru Integrated Clock Desktop / 小瑞総合時計デスクトップ）** 是一款基于 **Go 语言** 与 **Viva GPU 原生桌面框架**（Direct3D 11 + DirectWrite 硬件加速）打造的高性能全能桌面时钟与专注力效率工具。

- **无 WebView / 无 Electron 臃肿依赖**：单文件原生编译，启动毫秒级响应，内存占用极低。
- **零轮询空闲架构 (Zero-Polling Idle)**：应用在空闲时**绝不运行后台死循环轮询**，仅在整秒/整分边界或壁纸切换时刻精准唤醒重绘，平时 CPU 与 GPU 占用率保持 `0%`。
- **完整三语无缝切换**：内置 **简体中文 / English / 日本語** 完整本地化，所有开关状态（`是`/`否`、`ON`/`OFF`、`はい`/`いいえ`）、城市名称、天气、资讯与设置项实时同步。

---

## 🚀 核心功能一览

| 模块 | 功能亮点 |
| :--- | :--- |
| 🕒 **主时钟与双历超大字呈现** | 自动根据电脑操作系统区域与系统时区显示**所在国家/地区**（如 `📍 所在国家/地区：中国 · 北京 (UTC+08:00)`）及本地时间；**字号整体放大近两倍**（115-120px 时间大字、28px 阳历、24px 农历、20px 节日徽章），彻底杜绝地点文字重复；**阳历（公历）与中国传统农历（阴历）高精度双历并排**（含天干地支、生肖、月日，如 `🌙 农历 丙午年(马年) 九月廿九`）；支持 **世界各国当日节日、二十四节气（立冬等）与网络特色节日**（1024程序员节、黑色星期五打折季、双11、双12、618、520等；无节日显示“🎉 今日节日：无”）；支持 **数字时钟 / 经典指针表盘** 一键切换、12/24 小时制、秒数开关与实时本地天气（`°C` / `°F`）。 |
| 🌍 **全球世界时钟与各国当日节日** | 覆盖全球 **193 个联合国成员国首都 + 全球核心金融科技城市**。支持**中/英/日/时区实时搜索**、七大洲筛选、**重复添加弹窗提醒防重**、置顶收藏、上下排序与崩溃安全的双阶段持久化保护。每座城市卡片均根据当地时间与所属国家**实时显示该国正在庆祝的节日与二十四节气**。 |
| ⏰ **智能闹钟与真实合成铃声** | 内置纯 Go 程序化合成音频引擎（`经典钟声` / `柔和风铃` / `数字脉冲`），支持一键 **`🔊 试听`**、早晨/午间/晚间快捷预设、**`-1小时 / +1小时 / -5分钟 / -1分钟 / +1分钟 / +5分钟` 精确时间微调**（自动处理跨小时及 `23:59 ↔ 00:00` 跨午夜进位借位）、一次性/每天/工作日重复周期、响铃横幅提醒与 **贪睡 (+5分钟)**。 |
| ⏱️ **倒计时任务自定义与秒表** | **支持自定义创建、修改与删除倒计时任务**（支持任意自定义任务名与 **`-1小时/+1小时`、`-10分/-1分/+1分/+10分`、`-15秒/-1秒/+1秒/+15秒` 精确进位借位微调**，内置泡茶/煮面/午休/番茄/阅读/健身 6 种快捷模板，**修改与删除永久持久化落盘，重启绝不恢复已删除任务**，到期自动播放提示铃声）+ **毫秒级高精度秒表**（支持无限分段计圈 `Lap` 与累计耗时对比）。 |
| 🎯 **番茄专注追踪** | 内置 Pomodoro 工作流（专注工作 → 短休息 → 长休息自动循环），待机状态支持 **`-5分钟 / -1分钟 / +1分钟 / +5分钟` 精准调节专注时长**，提供 **🔕 勿扰模式 (`是`/`否`)**、跳过/推迟休息以及累计完成周期与专注时长统计。 |
| 🌙 **睡眠与白噪音合成器** | 极暗护眼夜间时钟界面，支持 `15 / 30 / 60 分钟` 快捷预设及 **`-5分钟 / -1分钟 / +1分钟 / +5分钟` 精确微调**，内置 **🌧️ 雨声 / 🌲 溪流 / 🌊 海浪 / 📻 白噪音** 真实无缝循环音频合成与试听，并在结束前 5 分钟自动线性渐弱音量（Fade-out）。 |
| 📜 **历史上的今天 (API 同步)** | 自动调用 **百度百科 / Wikipedia 历史上的今天公开 API** 获取当日真实历史大事记、科技探索与人文纪事，支持分类筛选、年份徽章展示、手动刷新及离线精选库兜底。 |
| 🖼️ **高性能壁纸轮播** | 支持单张/多张本地照片、**整个照片文件夹导入**或直接拖拽文件/目录到窗口；支持 **`1 / 5 / 10 / 15 分钟`** 轮播间隔、**随机洗牌 (Fisher-Yates) / 顺序播放**、5 种适应模式及毛玻璃卡片透光度调节；**后台自动轮播期间零硬盘写入**，充分保护 SSD 寿命。 |
| 📌 **桌面悬浮时钟与位置切换** | 优化紧凑型悬浮时钟窗口（`↗ 悬浮时钟`），**88px 超大字号**，界面内置 **[⬆ 顶部]**、**[⏺ 居中]**、**[⬇ 底部]** 垂直显示位置切换按钮并持久化记忆；支持 **`📌 置顶: 是 / 否`**、双历全信息、防多开单实例唤醒以及关闭按钮**最小化到系统托盘或直接退出**设置。 |
| ➖ **屏幕边缘自动缩小为状态条** | 当软件窗口靠近屏幕**顶部、左侧、右侧或底部边缘**时，自动收缩为 `560×118` 精简横条，简洁展示当前时间与公历/农历日期，点击右侧 **`⤢ 恢复正常`** 按钮即可一键恢复原窗口尺寸并防抖移离边缘；在 **设置 → ⏰ 时钟** 最顶部提供醒目的 **`✅ 启用状态条收缩`** 与 **`🚫 不启用`** 双按钮及 **`➖ 立即缩为状态条`** 快捷按钮。 |
| 🚀 **开机自动启动设置** | 在 **设置 → ⏰ 时钟** 中内置 **`✅ 启用开机自启`** 与 **`🚫 不启用`** 双按钮，一键注册或移除系统当前用户的登录开机自启项（Windows 注册表 `Run` 键 / macOS LaunchAgent / Linux XDG Autostart）。 |
| 👑 **Pro 会员与 7 天免费试用** | 顶部导航栏与 **设置 → 👑 会员** 提供醒目入口：新用户享 **7 天全功能免费试用**，支持 **按月订阅（`¥3 / 月`）** 与 **按年订阅（`¥28 / 年`，立省 22%）**，并内置完整测试版免付费一键切换台。 |

---

## 📖 详细使用指南 (User Guide)

### 1. 主时钟、悬浮时钟与屏幕边缘状态条
- **切换表盘与时间格式**：在主时钟底部控制栏可一键切换 **`24小时制 / 12小时制`**、**`显示秒数: 是/否`**、**`样式: 数字 / 指针`**。
- **进入桌面悬浮时钟**：点击顶部导航栏 **`↗ 悬浮时钟`** 或主时钟底部的 **`悬浮时钟模式 ↗`**。在悬浮小窗顶部可通过 **`[⬆ 顶部]`**、**`[⏺ 居中]`**、**`[⬇ 底部]`** 调整时钟在窗口内的垂直位置，点击 **`📌 置顶`** 可锁定在桌面最前端，点击 **`还原窗口 ↙`** 返回完整界面。
- **屏幕边缘自动缩小为状态条**：
  - **自动触发**：将窗口拖动靠近屏幕顶部或左右边缘，窗口会自动缩小为紧凑型时间日期状态条。
  - **手动触发**：点击主时钟底部控制栏的 **`➖ 缩为状态条`**（或 **设置 → ⏰ 时钟** 中的 **`➖ 立即缩为状态条`**）。
  - **恢复正常**：点击状态条右侧的 **`⤢ 恢复正常`** 按钮，即可恢复正常窗口大小并自动离开屏幕边缘。
  - **开启 / 关闭该功能**：进入 **设置 → ⏰ 时钟**，在最上方第一张卡片点击 **`✅ 启用状态条收缩`** 或 **`🚫 不启用`** 即可随时开启或禁用。

### 2. 开机自动启动设置
- 进入 **设置 → ⏰ 时钟**，在上方 **“开机自动启动”** 卡片中点击 **`✅ 启用开机自启`** 即可在每次登录电脑时自动运行本软件；若不需要开机自启，点击 **`🚫 不启用`** 即可立即取消。

### 3. 世界时钟管理（全球 193 国与核心都市）
- 点击 **`世界时钟`** 标签页 → 点击右上角 **`+ 添加城市`**。
- 在搜索框输入**中文、英文、日文、国家代码或时区名称**（如 `纽约`、`东京`、`伦敦`、`UTC`），或点击七大洲及国家快捷按钮过滤，点击 **`+ 添加`** 即可加入列表。
- 列表支持 **`↑` / `↓`** 调整顺序、**`★ 收藏`** 置顶、**`删除`** 移除，且自动拦截重复添加并弹出提醒。

### 4. 倒计时任务自定义（新建 / 修改 / 删除）与高精度秒表
- 点击 **`计时器`** 标签页：
  - **新建自定义任务**：点击 **`+ 新建自定义任务`**，输入任务名称（或点击 🍵泡茶、🍜煮面、😴午休、🍅番茄专注、📖深度学习、🏃运动健身 6 大快捷模板），通过 **`-1小时/+1小时`**、**`-10分/-1分/+1分/+10分`**、**`-15秒/-1秒/+1秒/+15秒`** 按钮精准调整时长（自动跨分/跨时进位借位）后点击 **`保存新建任务`**。
  - **修改原有任务**：点击任意倒计时卡片右侧的 **`✎ 编辑`**，修改名称或时长后保存。
  - **删除任务**：点击卡片右侧的 **`✕`** 即可永久删除（实时写入配置文件，退出重开绝不会再出现）。
  - **高精度秒表**：页面下方提供百分之一秒精度秒表，运行中点击 **`计圈`** 可记录无限分段耗时。

### 5. 智能闹钟、番茄专注与睡眠白噪音
- **闹钟**：在 **`闹钟`** 页点击 **`+ 新建闹钟`**，选择早晨/午间/晚间预设，或通过 **`-1小时 / +1小时 / -5分钟 / -1分钟 / +1分钟 / +5分钟`** 按钮将时间精确调整到任意分钟，选择重复周期与铃声（支持 **`🔊 试听`**）。到点时顶部弹出醒目红条并播放合成铃声，支持 **`贪睡 (+5分钟)`**。
- **时间追踪（番茄钟）**：在 **`时间追踪`** 页待机时可通过 **`-5分钟 / -1分钟 / +1分钟 / +5分钟`** 自定义专注时长，点击 **`开始专注时段`** 后自动按“专注工作 → 短休息 → 4周期后长休息”流转，支持 **`🔕 勿扰`** 与今日专注统计。
- **睡眠模式**：在 **`睡眠模式`** 页选择 `15 / 30 / 60 分钟` 预设或点击 **`-5分钟 / -1分钟 / +1分钟 / +5分钟`** 微调伴眠时长，选择环境音（雨声/溪流/海浪/白噪音，支持 **`🔊 试听环境音`**），结束前 5 分钟音量会自动柔和淡出至静音。

### 6. 设置与偏好中心（五大模块详解）
- **`🌐 语言`（语言切换 + 字体排版与配色）**：
  - 一键切换 **`🇨🇳 简体中文`**、**`🇺🇸 English`**、**`🇯🇵 日本語`**；
  - 实时预览并调节时钟数字大小（`48px–96px` 预设及 `±2px` 步进微调，范围 `32px–120px`）、时钟数字高亮色（纯白/琥珀金/橄榄绿/冰川青/霓虹紫/珊瑚红）、界面全域文字颜色（支持深色/浅色壁纸一键反色适配）与全局字号缩放（`90% / 100% / 110% / 120%`）。
- **`🖼️ 壁纸`（本地轮播图 + 毛玻璃 + 主题）**：
  - 开启背景壁纸后，可直接将多张 PNG/JPG 图片或整个照片文件夹**拖入窗口**，支持 `1 / 5 / 10 / 15 分钟` 随机或顺序轮播、5 种画面适配、背景透明度与卡片毛玻璃浓度调节，以及 4 款视觉主题（Olive / Dark / Light / Nord）。
- **`⏰ 时钟`（状态条 + 开机自启 + 温度单位 + 关闭行为）**：
  - 集中管理**靠近屏幕边缘自动缩小为状态条**、**开机自动启动**、**24小时制/秒数/完整日期/今日节日显示**、**默认表盘样式**、**悬浮窗显示位置**、**温度计量单位（摄氏度 `°C` / 华氏度 `°F`）** 以及 **窗口右上角关闭按钮（`✕`）行为**。
- **`👑 会员`（试用与订阅中心）**：
  - 查看 **7 天免费试用** 剩余天数，支持开通 **按月订阅（`¥3 / 月`）** 或 **按年订阅（`¥28 / 年`）**，提供恢复购买与测试状态切换台。
- **`ℹ️ 关于`**：查看 GPU 原生渲染架构说明与版本信息。

---

## ⚠️ 注意事项与常见问题 (Notes & Precautions)

1. **边缘收缩防抖机制说明**：
   - 当您在状态条模式下点击 **`⤢ 恢复正常`** 后，程序会自动将窗口向屏幕内侧偏移一定安全距离并开启防抖保护，避免刚恢复就立即再次缩回状态条；当您把窗口移回屏幕内部后，再次靠近屏幕边缘即可重新触发收缩。
   - 如果您在日常拖动窗口时不希望触发边缘自动收缩，请前往 **设置 → ⏰ 时钟**，在“靠近屏幕边缘自动缩小为状态条”卡片中点击 **`🚫 不启用`**。
2. **关闭按钮（✕）默认最小化到托盘**：
   - 为保证后台闹钟响铃、倒计时与专注计时不被意外中断，点击窗口右上角 `✕` 默认会**隐藏窗口并最小化到系统托盘**（右下角任务栏时钟图标）。
   - 如需点击 `✕` 直接退出程序，可在 **设置 → ⏰ 时钟 → 关闭按钮 (✕) 行为设置** 中切换为 **`✕ 直接关闭退出程序`**，或右键右下角系统托盘图标选择 **`直接退出程序`**。
3. **开机自启与程序移动位置**：
   - 启用“开机自动启动”后，系统启动项记录的是当前可执行文件的路径。如果您后续将编译出的 `.exe` 移动到了其他文件夹，只需在新位置重新打开一次软件，程序会自动刷新注册表中的启动路径。
4. **配置文件与自动容灾备份位置**：
   - 所有用户设置、世界城市、闹钟与自定义倒计时任务均保存在本地 `%APPDATA%\XiaoruIntegratedClockDesktop\settings.json`，并在每次有变更写入前自动通过内存快照备份至 `settings.json.bak`（当配置无实质变化时自动跳过写盘，实现零冗余硬盘写入）；若遇异常断电导致主配置损坏，启动时会自动从 `.bak` 恢复。
5. **离线运行与网络请求说明**：
   - 主时钟、农历、二十四节气、全球 193 国世界时钟、闹钟合成音频、倒计时、番茄钟、睡眠白噪音与壁纸轮播均**100% 纯本地离线运行**，无任何遥测或隐私上传。
   - 仅 **`历史上的今天`** 标签页会在后台按需请求一次公开百科 API（若电脑无网络则自动无缝回退至内置精选历史档案库，不影响任何功能使用）。
6. **关于仓库源码与构建产物**：
   - 本 GitHub 仓库仅托管纯净源代码与资源文件，不上传打包后的二进制可执行文件与 `dist/` 目录。请按照下方步骤使用 `go run ./cmd/worldclock` 直接运行，或执行 `.\build_windows.ps1` 在本地一键编译生成 `dist\小瑞综合时钟桌面.exe`（及 `dist\XiaoruIntegratedClockDesktop.exe`）。

---

## 🛠️ 30 秒快速上手 (Quick Start)

### 环境要求
- **操作系统**：Windows 10 / Windows 11 (64-bit)
- **Go 版本**：Go `1.25+` / `1.26+`

### 1. 克隆仓库
```powershell
git clone https://github.com/vicky-clair/world_Clock_Pro_Desktop_Time_Alarms-Focus_Timer.git
cd world_Clock_Pro_Desktop_Time_Alarms-Focus_Timer
```

### 2. 直接运行开发版
```powershell
go run ./cmd/worldclock
```

### 3. 一键打包带数字签名的 Windows 正式版 (`.exe`)
项目内置了自动化构建脚本 `build_windows.ps1`，会自动嵌入 Windows 原生 PE 版本信息、图标与 DPI 感知清单（`.syso`），并进行本地 Authenticode 代码签名：
```powershell
powershell -ExecutionPolicy Bypass -File .\build_windows.ps1
```
构建完成后，直接双击运行生成的可执行文件：
- `dist\小瑞综合时钟桌面.exe`（或 `dist\XiaoruIntegratedClockDesktop.exe`）

### 4. 运行全套自动化测试与性能基准
无需弹出真实桌面窗口，使用内置 `app.Driver` 即可在毫秒级完成全界面交互测试：
```powershell
# 代码格式化、静态检查与全量单元测试
gofmt -w cmd pkg
go vet ./...
go test ./... -count=1

# 运行壁纸轮播与核心模块性能基准测试
go test -bench=. -benchmem ./pkg/wallpaper
```

---

## ⚡ 性能与硬件寿命保护设计 (Go 优势发挥)

1. **硬盘读写 (Disk I/O) 零磨损保护 (`pkg/storage` & `pkg/ui`)**：
   - **内存字节比对免写 (`lastSaved`)**：每次调用保存时，先在内存比对序列化 JSON 与上一次已落盘字节，若配置无实质变化则**直接跳过磁盘写入（`0 B` 落盘）**。
   - **内存级 `.bak` 容灾备份**：覆写 `settings.json` 前直接从内存缓存写入 `settings.json.bak`，省去每次保存前重复 `os.ReadFile` 读取硬盘的操作。
   - **壁纸轮播零写盘**：后台壁纸定时轮播仅推进内存游标，绝不周期性刷写磁盘，仅在用户主动改设置或退出窗口时保存一次。
2. **CPU 与内存零分配渲染优化 (`pkg/ui` & `pkg/clock`)**：
   - **零切片拷贝标量读取 (`Store.GetPrefs`)** 与 **跨日历法/农历/节日按天缓存 (`DailyClockMeta`)**：将原本每秒重复计算的农历与多国节日降至每天仅计算 1 次；主时钟绘制基准测试（`BenchmarkPaintMainClock`）单帧堆内存分配从 `14,471 B/op` 降至 **`136 B/op`（减少 99.06%）**，CPU 耗时降低 **31.4%**。
3. **壁纸轮播零卡顿 (`pkg/wallpaper`)**：
   - **仅索引路径**：即使导入上千张高清壁纸文件夹，内存也只保存轻量级路径切片 `[]string`，绝不预加载多张大图。
   - **单纹理自动降采样**：切换壁纸时仅解码当前 1 张照片，超过 `1920px` 的 4K/8K 原图自动按比例降采样后再送入 GPU，立刻释放原图内存。
   - **零分配到期判定**：`AdvanceIfDue` 基准测试实测仅耗时 **`14.15 ns/op`**，**`0 B/op`**，**`0 allocs/op`**。
4. **精确边界唤醒 (`pkg/clock` & `pkg/ui`)**：
   - 开启秒数时精确对齐下一秒边界唤醒；关闭秒数时对齐下一分钟边界唤醒，结合壁纸轮播计划时间取最小值调度，彻底杜绝后台空转。

---

## 📂 项目目录结构

```text
the-world-clock-on-the-desktop/
├── cmd/worldclock/          # 应用主入口、单实例进程锁与 Windows 原生 PE 资源 (.syso)
├── pkg/
│   ├── action/              # 自动化动作与提醒 URL 安全校验器
│   ├── alarm/               # 闹钟调度器、重复规则与贪睡管理
│   ├── audio/               # 纯 Go 程序化 16-bit PCM WAV 音频合成与异步播放引擎
│   ├── clock/               # 时间格式化、双历(公历/农历)、二十四节气、全球节日与边界唤醒算法
│   ├── i18n/                # 中 / 英 / 日 三语完整词典与本地化日期格式
│   ├── license/             # 7 天全功能免费试用与 Pro 会员订阅管理 (¥3/月、¥28/年)
│   ├── rss/                 # “历史上的今天”百度百科 / Wikipedia API 同步与离线精选库
│   ├── sleep/               # 睡眠模式、环境白噪音与末段音量线性渐弱引擎
│   ├── storage/             # 零冗余写盘 (lastSaved 比对)、零拷贝读取 (GetPrefs) 与 .bak 容灾恢复
│   ├── timer/               # 多路自定义倒计时器与高精度计圈秒表
│   ├── tracker/             # 番茄工作法专注追踪与勿扰模式管理
│   ├── ui/                  # Viva GPU 声明式界面视图、主题引擎与 app.Driver 无头测试套件
│   ├── wallpaper/           # 高性能壁纸引擎、文件夹扫描、1080p 降采样与随机洗牌轮播
│   ├── weather/             # 城市天气状态与摄氏/华氏温度转换服务
│   └── world/               # 全球 193 国 + 重点城市目录、多语言检索与世界时钟管理
├── assets/                  # 应用图标资源 (icon.ico / icon.png)
├── docs/                    # 中/英/日完整开发指南、项目审计报告、商店上架清单与隐私政策
├── build_windows.ps1        # Windows 一键编译、数字签名与发行包打包脚本
└── viva.toml                # Viva 框架应用构建配置
```

---

# 🇺🇸 English Documentation

## ✨ Overview

**Xiaoru Integrated Clock Desktop (小瑞综合时钟桌面 / 小瑞総合時計デスクトップ)** is a modern, GPU-accelerated desktop time suite built with **Go** and the **[Viva](https://github.com/vivaui/viva)** native UI framework (Direct3D 11 & DirectWrite).

- **No WebView / No Electron Bloat**: Compiles into a single lightweight native binary with instant startup and minimal RAM footprint.
- **Zero-Polling Idle Architecture**: Draws nothing while idle. Frame redraws are scheduled strictly at exact second/minute boundaries or wallpaper rotation deadlines (`0.0%` CPU & GPU when idle).
- **Full Trilingual Localization**: Instant in-app switching between **English**, **Simplified Chinese (简体中文)**, and **Japanese (日本語)** across every view, button, city name, and toggle (`ON`/`OFF`, `是`/`否`, `はい`/`いいえ`).

---

## 🚀 Key Features

- **🕒 Main Clock & Dual Solar/Lunar Calendar**: 2x scaled typography (~120px clock digits, 28px solar date, 24px lunar date, 20px holiday badge) with zero duplicated country text. Pairs standard Gregorian calendar with high-precision offline Chinese Lunar Calendar (GanZhi, Chinese Zodiac, Lunar month/day). Automatically identifies host country/region and timezone (e.g., `📍 Country/Region: United States · New York (UTC-05:00)`). Supports digital & vector analog dials, 12H/24H format, seconds toggle, and live weather (`°C` / `°F`).
- **🎉 World Holidays, 24 Solar Terms & Shopping/Tech Festivals**: Comprehensive holiday engine for CN, US, JP, GB, and FR, astronomical 24 Solar Terms (Lidong, etc.), and tech/shopping festivals (1024 Programmers' Day, Black Friday, Cyber Monday, Double 11, Double 12, 618, 520). Displays "🎉 Today's Holiday: None" when no holiday falls on the current day.
- **🌍 Global World Clock (200+ Cities & 193 Countries)**: Real-time search across English, Chinese, Japanese, country codes, and IANA timezones (`America/New_York`, `Asia/Tokyo`, `Europe/London`, etc.). Supports continent filters, favorites (`★`), custom ordering (`↑`/`↓`), duplicate detection alert dialogs, and per-city local holiday displays.
- **⏰ Smart Alarms & Procedural Audio Chimes**: Procedural Go audio synthesizer (Classic Bell, Gentle Chime, Electronic Beep) with instant **`🔊 Preview`**, Morning/Noon/Evening presets, custom hour/minute steppers, One-time/Daily/Weekday recurrence, alert banner, and **Snooze (+5m)**.
- **⏱️ Custom Multi-Timers & Split-Lap Stopwatch**: Create, edit, and delete custom timers with arbitrary names and `HH:MM:SS` duration; includes 6 fast presets (Tea, Noodles, Nap, Pomodoro, Reading, Workout). All edits and deletions are atomically persisted to disk (deleted timers never resurrect on app restart). Monotonic centisecond stopwatch with unlimited lap history.
- **🎯 Pomodoro Focus Tracker**: Automated Work → Short Break → Big Break cycles, **Do Not Disturb (`DND: ON/OFF`)**, break skip/postpone, and daily focus statistics.
- **🌙 Night Sleep & Ambient Mode**: Low-light OLED-friendly night clock with `15m / 30m / 60m` sleep timer, procedural ambient soundscapes (**Gentle Rain**, **Forest Stream**, **Ocean Waves**, **White Noise**), sound preview, and 5-minute linear volume fade-out.
- **📜 On This Day in History (API Sync)**: Automatically fetches authentic historical milestones, science discoveries, and cultural events via Baidu Baike and Wikipedia public APIs, backed by an offline curated archive.
- **🖼️ High-Performance Wallpaper Carousel**: Load single/multiple local photos, **entire photo folders**, or drag-and-drop onto the window. Features **`1m / 5m / 10m / 15m`** rotation intervals, **Shuffle (Fisher-Yates)** or **Sequential** playback, automatic `1920px` downsampling, 5 fit modes, and frosted-glass card opacity controls.
- **📌 Desktop Popout Mini-Clock with Alignment Controls**: Enlarged 88px floating clock widget (`↗ Popout`) with dedicated on-screen buttons to position the clock at the **[⬆ Top]**, **[⏺ Center]**, or **[⬇ Bottom]** of the window with persistent memory. Includes **Always-on-Top (`📌 Pin: ON/OFF`)**, dual calendar, and clean tray minimization.
- **➖ Auto-Shrink to Edge Status Bar (`Edge Bar`)**: Automatically shrinks into a compact `560×118` status bar showing time and solar/lunar dates when dragged near any screen edge, with a **`⤢ Restore Normal`** button to restore full size. Includes prominent **`✅ Enable Edge Bar`** and **`🚫 Do Not Enable`** buttons in **Settings → ⏰ Clock**.
- **🚀 Launch at System Startup**: Built-in **`✅ Enable Autostart`** and **`🚫 Disable`** buttons in **Settings → ⏰ Clock** to automatically register or remove login startup entries.

---

## 🛠️ Quick Start

### Prerequisites
- **OS**: Windows 10 / 11 (x64)
- **Go**: `1.25+` or `1.26+`

### Build & Run
```powershell
# 1. Clone the repository
git clone https://github.com/vicky-clair/world_Clock_Pro_Desktop_Time_Alarms-Focus_Timer.git
cd world_Clock_Pro_Desktop_Time_Alarms-Focus_Timer

# 2. Run directly from source
go run ./cmd/worldclock

# 3. Build signed production Windows executable (dist\小瑞综合时钟桌面.exe & dist\XiaoruIntegratedClockDesktop.exe)
powershell -ExecutionPolicy Bypass -File .\build_windows.ps1

# 4. Run headless UI & unit tests
go test ./... -count=1
```

---

# 🇯🇵 日本語ドキュメント (Japanese)

## ✨ プロジェクト概要

**小瑞総合時計デスクトップ（Xiaoru Integrated Clock Desktop / 小瑞综合时钟桌面）** は、**Go 言語** と **Viva GPU ネイティブ UI フレームワーク**（Direct3D 11 / DirectWrite ハードウェアアクセラレーション）で構築された、軽量かつ高速なデスクトップ時計・アラーム・集中タイマーアプリです。

- **WebView / Electron 不使用**：単一のネイティブ実行ファイルにコンパイルされ、瞬時に起動しメモリ消費を最小限に抑えます。
- **ゼロポーリング・アイドル設計**：待機中はバックグラウンドループを一切実行せず、秒・分の境界や壁紙切り替えタイミングでのみ再描画をスケジュールします（アイドル時の CPU/GPU 使用率 `0%`）。
- **日・中・英 3言語完全対応**：設定から **日本語 / 简体中文 / English** をワンクリックで切り替え可能。すべてのトグル（`はい`/`いいえ`、`是`/`否`、`ON`/`OFF`）、都市名、天気、設定項目が完全にローカライズされています。

---

## 🚀 主な機能

- **🕒 メイン時計＆太陽暦・旧暦デュアル表示**：文字サイズ約2倍の特大視認性（115-120px 時間数字、28px 太陽暦日付、24px 旧暦日付、20px 祝日バッジ）と国名・都市名表記の重複を完全解消。高精度オフライン太陰太陽暦（干支・十二支・旧暦月日）を併記。OSの国・地域とタイムゾーンを自動検出し現在地バッジ（例：`📍 現在の国・地域：日本 · 東京 (UTC+09:00)`）を表示。デジタル・アナログ文字盤切替、12H/24H、秒表示スイッチ、リアルタイム天気・気温（`°C` / `°F`）を完備。
- **🎉 世界の祝日・二十四節気・ネット＆セール記念日**：日本（JP）、中国（CN）、アメリカ（US）、イギリス（GB）、フランス（FR）の祝日、高精度天文学アルゴリズムによる二十四節気（立冬など）、およびネット記念日・ショッピング祭（1024プログラマーの日、ブラックフライデー、サイバーマンデー、独身の日等）。祝日がない日は「🎉 今日の祝日・記念日：なし」と明示表示。
- **🌍 世界時計（193カ国＋主要都市）**：東京、大阪、ニューヨーク、ロンドン、上海、パリなど世界中の都市を**日本語・英語・中国語・タイムゾーンでリアルタイム検索**。大陸別フィルター、お気に入り固定（`★`）、並び替え（`↑`/`↓`）、重複追加警告ダイアログ、各都市現地の祝日・節気表示に対応。
- **⏰ スマートアラーム＆純Goプロシージャル音声合成**：純Go製PCMオーディオ合成（ベル音、ウィンドチャイム、電子音）とワンクリック **`🔊 試聴`**、朝・昼・夕方のプリセット、1時間/5分単位の微調整、1回のみ/毎日/平日の繰り返し設定、**スヌーズ（+5分）**、アラーム通知バナーを搭載。
- **⏱️ カウントダウンタイマーの自由追加・編集・削除＆ストップウォッチ**：任意の名前と時間（時・分・秒）でタイマーを自由に追加・編集・削除可能（お茶、ラーメン、昼寝、ポモドーロ、読書、ワークアウトの6大テンプレート付属）。削除したタイマーは永続化され、アプリ再起動後も復活しません。1/100秒精度の高精度ストップウォッチ（無制限ラップ計測）も搭載。
- **🎯 ポモドーロ集中トラッカー**：作業セッション・小休憩・長休憩の自動サイクル管理、**🔕 集中モード（`はい`/`いいえ`）**、休憩のスキップ・延長、累計集中時間の統計表示。
- **🌙 スリープ & 環境音モード**：目に優しい暗色ナイトクロック、`15分 / 30分 / 60分` のスリープタイマー、**穏やかな雨音 / 森林の小川 / 波の音 / ホワイトノイズ**、環境音の試聴、終了前5分間の自動音量フェードアウト機能。
- **📜 今日は何の日（歴史・API連携）**：Baidu Baike および Wikipedia の公開 API と連携し、今日の歴史的出来事や記念日を自動取得。オフライン精選アーカイブも内蔵。
- **🖼️ 高性能壁紙スライドショー**：複数画像や**写真フォルダ全体の一括読み込み**、ウィンドウへのドラッグ＆ドロップに対応。**`1分 / 5分 / 10分 / 15分`** 間隔、**シャッフル / 順番再生**、自動ダウンサンプリングによる省メモリ描画、すりガラス透明度調整を備えています。
- **📌 表示位置切替対応ポップアウト・ミニ時計**：88px 特大文字盤のコンパクト表示（`↗ ミニ時計`）。画面上部のボタンで **[⬆ 上部]**、**[⏺ 中央]**、**[⬇ 下部]** の垂直位置をワンクリック切替可能（設定は自動保存）。最前面固定（`📌 最前面固定: はい / いいえ`）とトレイ格納にも対応。
- **➖ 画面端でのミニバー自動縮小（エッジバー）**：ウィンドウを画面端に近づけると日時のみのスリムなステータスバー（`560×118`）に自動縮小し、**`⤢ 通常表示に戻す`** ボタンで元のサイズに復元します。**設定 → ⏰ 時計** にて **`✅ ミニバー縮小を有効化`** / **`🚫 無効にする`** をワンクリックで切り替え可能。
- **🚀 OS起動時の自動実行**：**設定 → ⏰ 時計** にて **`✅ 自動起動を有効化`** / **`🚫 無効にする`** ボタンでPCサインイン時の自動起動を設定可能。

---

## 🛠️ クイックスタート (使い方)

```powershell
# 1. リポジトリのクローン
git clone https://github.com/vicky-clair/world_Clock_Pro_Desktop_Time_Alarms-Focus_Timer.git
cd world_Clock_Pro_Desktop_Time_Alarms-Focus_Timer

# 2. ソースコードから直接実行
go run ./cmd/worldclock

# 3. 署名付き Windows 実行ファイル (dist\小瑞综合时钟桌面.exe) のビルド
powershell -ExecutionPolicy Bypass -File .\build_windows.ps1

# 4. 全自動テストの実行
go test ./... -count=1
```

---

## 📄 开发者文档与架构审计 (Documentation)

- 🇨🇳 **中文完整开发文档**: [`docs/DEVELOPER_GUIDE_ZH.md`](docs/DEVELOPER_GUIDE_ZH.md)
- 🇺🇸 **English Developer Guide**: [`docs/DEVELOPER_GUIDE_EN.md`](docs/DEVELOPER_GUIDE_EN.md)
- 🇯🇵 **日本語開発ガイド**: [`docs/DEVELOPER_GUIDE_JA.md`](docs/DEVELOPER_GUIDE_JA.md)
- 🔍 **项目深度审计报告**: [`docs/PROJECT_AUDIT_REPORT.md`](docs/PROJECT_AUDIT_REPORT.md)
- 🔒 **Privacy Policy**: Zero telemetry, zero external tracking, local configuration storage (`%APPDATA%\XiaoruIntegratedClockDesktop\settings.json`). See [`docs/PRIVACY_POLICY.md`](docs/PRIVACY_POLICY.md) for details.
- 🏪 **Microsoft Store Guide**: See [`docs/MICROSOFT_STORE_LISTING.md`](docs/MICROSOFT_STORE_LISTING.md).
