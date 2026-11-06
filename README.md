<div align="center">

# 🌍 World Clock Pro: Desktop Time, Alarms & Focus Timer

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

**World Clock Pro: Desktop Time, Alarms & Focus Timer** 是一款基于 **Go 语言** 与 **Viva GPU 原生桌面框架**（Direct3D 11 + DirectWrite 硬件加速）打造的高性能全能桌面时钟与专注力效率工具。

- **无 WebView / 无 Electron 臃肿依赖**：单文件原生编译，启动毫秒级响应，内存占用极低。
- **零轮询空闲架构 (Zero-Polling Idle)**：应用在空闲时**绝不运行后台死循环轮询**，仅在整秒/整分边界或壁纸切换时刻精准唤醒重绘，平时 CPU 与 GPU 占用率保持 `0%`。
- **完整三语无缝切换**：内置 **简体中文 / English / 日本語** 完整本地化，所有开关状态（`是`/`否`、`ON`/`OFF`、`はい`/`いいえ`）、城市名称、天气、资讯与设置项实时同步。

---

## 🚀 核心功能一览

| 模块 | 功能亮点 |
| :--- | :--- |
| 🕒 **主时钟与所在国家识别** | 自动根据电脑操作系统区域与系统时区显示**所在国家/地区**（如 `📍 所在国家/地区：中国 · 北京 (UTC+08:00)`）及本地时间；支持 **数字时钟 / 经典指针表盘** 一键切换、12/24 小时制、秒数显示开关（`是`/`否`）、完整/简短日期以及实时本地天气温度（`°C` / `°F`）。 |
| 🌍 **全球世界时钟** | 覆盖全球 **193 个联合国成员国首都 + 全球核心金融科技城市**。支持**中/英/日/时区实时搜索**、七大洲筛选、**重复添加弹窗提醒防重**、置顶收藏、上下排序与崩溃安全的双阶段持久化保护。 |
| ⏰ **智能闹钟与真实合成铃声** | 内置纯 Go 程序化合成音频引擎（`经典钟声` / `轻柔风铃` / `清脆电子铃`），支持一键 **`🔊 试听`**、早晨/午间/晚间快捷预设、自定义时分微调、一次性/每天/工作日重复周期、响铃横幅提醒与 **贪睡 (+5分钟)**。 |
| ⏱️ **倒计时与秒表** | 多任务并行倒计时器（到期自动播放提示铃声）+ **毫秒级高精度秒表**（支持无限分段计圈 `Lap` 与累计耗时对比）。 |
| 🎯 **番茄专注追踪** | 内置 Pomodoro 工作流（专注工作 → 短休息 → 长休息自动循环）、**🔕 勿扰模式 (`是`/`否`)**、跳过/推迟休息以及累计完成周期与专注时长统计。 |
| 🌙 **睡眠与白噪音合成器** | 极暗护眼夜间时钟界面，支持 `15 / 30 / 60 分钟` 睡眠定时、内置 **🌧️ 雨声 / 🌲 溪流 / 🌊 海浪 / 📻 白噪音** 真实无缝循环音频合成与试听，并在结束前 5 分钟自动线性渐弱音量（Fade-out）。 |
| 📜 **历史上的今天 (API 同步)** | 自动调用 **百度百科 / Wikipedia 历史上的今天公开 API** 获取当日真实历史大事记、科技探索与人文纪事，支持分类筛选、年份徽章展示、手动刷新及离线精选库兜底。 |
| 🖼️ **高性能壁纸轮播** | 支持单张/多张本地照片、**整个照片文件夹导入**或直接拖拽文件/目录到窗口；支持 **`1 / 5 / 10 / 15 分钟`** 轮播间隔、**随机洗牌 (Fisher-Yates) / 顺序播放**、5 种适应模式及毛玻璃卡片透光度调节。 |
| 📌 **桌面悬浮窗与托盘驻留** | 一键切换紧凑型悬浮时钟窗口（`↗ 悬浮时钟`），支持 **`📌 置顶: 是 / 否`**、防多开单实例唤醒以及关闭按钮**最小化到系统托盘或直接退出**设置。 |

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
- `dist\WorldClockPro.exe`

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

## ⚡ 性能架构设计 (Go 优势发挥)

1. **壁纸轮播零卡顿 (`pkg/wallpaper`)**：
   - **仅索引路径**：即使导入上千张高清壁纸文件夹，内存也只保存轻量级路径切片 `[]string`，绝不预加载多张大图。
   - **单纹理自动降采样**：切换壁纸时仅解码当前 1 张照片，超过 `1920px` 的 4K/8K 原图自动按比例降采样后再送入 GPU，立刻释放原图内存。
   - **零分配到期判定**：`AdvanceIfDue` 基准测试实测仅耗时 **`14.15 ns/op`**，**`0 B/op`**，**`0 allocs/op`**。
2. **精确边界唤醒 (`pkg/clock` & `pkg/ui`)**：
   - 开启秒数时精确对齐下一秒边界唤醒；关闭秒数时对齐下一分钟边界唤醒，结合壁纸轮播计划时间取最小值调度，彻底杜绝后台空转。

---

## 📂 项目目录结构

```text
world_Clock_Pro_Desktop_Time_Alarms-Focus_Timer/
├── cmd/worldclock/          # 应用主入口与 Windows 原生 PE 资源 (.syso)
├── pkg/
│   ├── action/              # 自动化动作与提醒触发器
│   ├── alarm/               # 闹钟调度器、重复规则与贪睡管理
│   ├── clock/               # 时间格式化、时区计算与整秒/整分边界唤醒算法
│   ├── i18n/                # 中 / 英 / 日 三语完整词典与本地化日期格式
│   ├── rss/                 # 科技 / 国际 / 科学 精选资讯源管理
│   ├── sleep/               # 睡眠模式、环境白噪音与音量渐弱引擎
│   ├── storage/             # 线程安全的 JSON 配置持久化存储
│   ├── timer/               # 多路倒计时器与高精度计圈秒表
│   ├── tracker/             # 番茄工作法专注追踪与勿扰模式管理
│   ├── ui/                  # Viva GPU 声明式界面视图、主题引擎与无头测试套件
│   ├── wallpaper/           # 高性能壁纸引擎、文件夹扫描、降采样与随机洗牌轮播
│   ├── weather/             # 城市天气状态与摄氏/华氏温度转换服务
│   └── world/               # 全球 193 国 + 重点城市目录、多语言检索与世界时钟管理
├── assets/                  # 应用图标资源 (icon.ico / icon.png)
├── docs/                    # Microsoft Store 商店上架指南与隐私政策文档
├── build_windows.ps1        # Windows 一键编译与数字签名脚本
└── viva.toml                # Viva 框架应用构建配置
```

---

# 🇺🇸 English Documentation

## ✨ Overview

**World Clock Pro: Desktop Time, Alarms & Focus Timer** is a modern, GPU-accelerated desktop time suite built with **Go** and the **[Viva](https://github.com/vivaui/viva)** native UI framework (Direct3D 11 & DirectWrite).

- **No WebView / No Electron Bloat**: Compiles into a single lightweight native binary with instant startup and minimal RAM footprint.
- **Zero-Polling Idle Architecture**: Draws nothing while idle. Frame redraws are scheduled strictly at exact second/minute boundaries or wallpaper rotation deadlines (`0.0%` CPU & GPU when idle).
- **Full Trilingual Localization**: Instant in-app switching between **English**, **Simplified Chinese (简体中文)**, and **Japanese (日本語)** across every view, button, city name, and toggle (`ON`/`OFF`, `是`/`否`, `はい`/`いいえ`).

---

## 🚀 Key Features

- **🕒 Main Clock & Analog/Digital Faces**: Switch between high-contrast Digital typography and a vector Analog dial, 12H/24H formats, seconds toggle, full/short date, and local weather (`°C` / `°F`).
- **🌍 Global World Clock (200+ Cities & 193 Countries)**: Real-time search across English, Chinese, Japanese, country codes, and IANA timezones (`America/New_York`, `Asia/Tokyo`, `Europe/London`, etc.). Supports continent filters, favorites (`★`), custom ordering (`↑`/`↓`), and one-click add/remove.
- **⏰ Smart Alarms**: Quick time presets (Morning, Noon, Evening), fine-grained hour/minute steppers, One-time / Daily / Weekday recurrence, active alert banner, **Snooze (+5m)**, and custom chime labels.
- **⏱️ Multi-Timers & Split-Lap Stopwatch**: Run multiple concurrent countdown timers alongside a millisecond-precision stopwatch with unlimited split laps.
- **🎯 Pomodoro Focus Tracker**: Automated Work → Short Break → Big Break cycles, **Do Not Disturb (`DND: ON/OFF`)**, break skip/postpone, and daily focus statistics.
- **🌙 Night Sleep & Ambient Mode**: Low-light OLED-friendly night clock with `15m / 30m / 60m` sleep timer, ambient soundscapes (**Gentle Rain**, **Ocean Waves**, **White Noise**), and automatic volume fade-out.
- **🖼️ High-Performance Wallpaper Carousel**: Load single/multiple local photos, **entire photo folders**, or drag-and-drop onto the window. Features **`1m / 5m / 10m / 15m`** rotation intervals, **Shuffle (Fisher-Yates)** or **Sequential** playback, automatic `1920px` downsampling, 5 fit modes, and frosted-glass card opacity controls.
- **📌 Desktop Popout Mini-Clock**: Compact floating clock mode (`↗ Popout`) with **Always-on-Top (`📌 Pin: ON/OFF`)**.

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

# 3. Build signed production Windows executable (dist\WorldClockPro.exe)
powershell -ExecutionPolicy Bypass -File .\build_windows.ps1

# 4. Run headless UI & unit tests
go test ./... -count=1
```

---

# 🇯🇵 日本語ドキュメント (Japanese)

## ✨ プロジェクト概要

**World Clock Pro: Desktop Time, Alarms & Focus Timer** は、**Go 言語** と **Viva GPU ネイティブ UI フレームワーク**（Direct3D 11 / DirectWrite ハードウェアアクセラレーション）で構築された、軽量かつ高速なデスクトップ時計・アラーム・集中タイマーアプリです。

- **WebView / Electron 不使用**：単一のネイティブ実行ファイルにコンパイルされ、瞬時に起動しメモリ消費を最小限に抑えます。
- **ゼロポーリング・アイドル設計**：待機中はバックグラウンドループを一切実行せず、秒・分の境界や壁紙切り替えタイミングでのみ再描画をスケジュールします（アイドル時の CPU/GPU 使用率 `0%`）。
- **日・中・英 3言語完全対応**：設定から **日本語 / 简体中文 / English** をワンクリックで切り替え可能。すべてのトグル（`はい`/`いいえ`、`是`/`否`、`ON`/`OFF`）、都市名、天気、設定項目が完全にローカライズされています。

---

## 🚀 主な機能

- **🕒 メイン時計 & アナログ/デジタル表示**：24時間/12時間表示、秒表示（`はい`/`いいえ`）、デジタル・アナログ文字盤の切り替え、日付フォーマット、天気・気温表示（`°C` / `°F`）に対応。
- **🌍 世界時計（193カ国＋主要都市）**：東京、大阪、ニューヨーク、ロンドン、上海、パリなど世界中の都市を**日本語・英語・中国語・タイムゾーンでリアルタイム検索**。大陸別フィルター、お気に入り固定（`★`）、並び替え（`↑`/`↓`）、自由な追加・削除が可能です。
- **⏰ スマートアラーム**：朝・昼・夕方のプリセット、1時間/5分単位の微調整、1回のみ/毎日/平日の繰り返し設定、**スヌーズ（+5分）**、アラーム通知バナーを搭載。
- **⏱️ カウントダウンタイマー & ストップウォッチ**：複数の同時タイマー（お茶、ポモドーロ集中、小休憩など）と、ラップ計測（`ラップ #01`）対応の高精度ストップウォッチ。
- **🎯 ポモドーロ集中トラッカー**：作業セッション・小休憩・長休憩の自動サイクル管理、**🔕 集中モード（`はい`/`いいえ`）**、休憩のスキップ・延長、累計集中時間の統計表示。
- **🌙 スリープ & 環境音モード**：目に優しい暗色ナイトクロック、`15分 / 30分 / 60分` のスリープタイマー、**穏やかな雨音 / 波の音 / ホワイトノイズ**、終了前の自動フェードアウト機能。
- **🖼️ 高性能壁紙スライドショー**：複数画像や**写真フォルダ全体の一括読み込み**、ウィンドウへのドラッグ＆ドロップに対応。**`1分 / 5分 / 10分 / 15分`** 間隔、**シャッフル / 順番再生**、自動ダウンサンプリングによる省メモリ描画、すりガラス透明度調整を備えています。
- **📌 ミニ時計（ポップアウト表示）**：作業中も邪魔にならないコンパクト表示（`↗ ミニ時計`）と **最前面固定（`📌 最前面固定: はい / いいえ`）** に対応。

---

## 🛠️ クイックスタート (使い方)

```powershell
# 1. リポジトリのクローン
git clone https://github.com/vicky-clair/world_Clock_Pro_Desktop_Time_Alarms-Focus_Timer.git
cd world_Clock_Pro_Desktop_Time_Alarms-Focus_Timer

# 2. ソースコードから直接実行
go run ./cmd/worldclock

# 3. 署名付き Windows 実行ファイル (dist\WorldClockPro.exe) のビルド
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
- 🔒 **Privacy Policy**: Zero telemetry, zero external tracking, local configuration storage (`%APPDATA%\OliveClock\settings.json`). See [`docs/PRIVACY_POLICY.md`](docs/PRIVACY_POLICY.md) for details.
- 🏪 **Microsoft Store Guide**: See [`docs/MICROSOFT_STORE_LISTING.md`](docs/MICROSOFT_STORE_LISTING.md).
