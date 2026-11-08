# Viva 框架下桌面时钟应用开发实施文档

**版本：** 1.1（三平台桌面版）  
**编制日期：** 2026-11-05  
**目标框架：** `github.com/vivaui/viva`  
**主要语言：** Go  
**目标产品：** 面向 Windows、macOS、Linux 的原生桌面时间管理软件（`小瑞综合时钟桌面 (Xiaoru Integrated Clock Desktop)`）  
**重点增强：** 基于屏幕尺寸、DPI 和用户偏好的自适应背景图片/主题系统

> 本文是基于 Viva 当前公开仓库与公开文档制定的应用开发方案。产品需求参考某款主流桌面时钟软件的常见功能；工程架构、数据模型与图片适配系统为本项目原创设计。Viva 版本、接口与后端支持会持续变化，项目必须锁定实际采用的版本并以该版本 API 为准。

---

# 1. 技术决策摘要

## 1.1 为什么选择 Viva

Viva 是 Go-first、GPU-native 的桌面 UI 框架。它通过平台原生窗口和 GPU API 绘制，不以 WebView、HTML/CSS/JavaScript 或捆绑浏览器运行时作为主 UI。其公开 README 说明的后端包括 macOS Cocoa/Metal、Windows Win32/Direct3D 11、Linux XCB/Vulkan；系统 WebView 是可选集成，而不是默认渲染方式。

这意味着本项目应采用 Viva 的绘制模型：

- 以 `app.Config.Paint` 描述每次重绘时的整张界面。
- 使用 `ui.Canvas` 绘制，使用 `flex.View`/`flex.Frame` 做声明式布局。
- 业务状态由应用自行持有；Paint 读取状态，不在 Paint 中执行副作用。
- 用 `ui.Tree` 和稳定 ID 管理交互目标。
- 事件发生时更新状态并使窗口重绘；空闲时不运行轮询式渲染循环。
- UI 组件优先使用 Viva 自带的 `components` 与 `flex`，不得默认引入 Web UI 思路。
- 外部图片、音频和天气数据在后台加载；Paint 只读取已经准备好的资源与状态。

## 1.2 适用范围与平台策略

本项目**只面向三类桌面操作系统交付：Windows、macOS 和 Linux**。不开发移动端版本，也不把浏览器/WebAssembly 作为产品交付形态。三平台共享 Go 业务核心、数据模型、图片处理算法和大部分 UI 代码；系统托盘、开机启动、通知、显示器枚举、窗口行为等由平台适配层实现。

| 目标平台 | 首发支持范围 | 实施策略 |
|---|---|---|
| Windows | Windows 10/11 x64 | 第一阶段主开发与验收平台；优先完成安装包、托盘、开机启动和多显示器/DPI 验证 |
| macOS | 当前 Viva 支持的 macOS 版本与架构 | 与 Windows 并行建立构建流水线；验证 Apple Silicon，Intel 仅在 Viva 版本和依赖实际支持时纳入 |
| Linux | 主流 x86_64 桌面发行版 | 先以 X11 路径完成基线，再验证 Wayland；明确记录发行版、桌面环境、显示服务器和 GPU 后端 |

**平台范围边界：**
- Windows ARM64 属于 Windows 平台的可选扩展架构，不作为首发阻塞项。
- Linux 不承诺“所有发行版均无条件支持”；首发应选定并测试代表性发行版/桌面环境，再逐步扩大兼容矩阵。
- Linux X11 与 Wayland 都属于 Linux 桌面运行环境，但必须分别验证。若 Viva 的 GTK3 shell/Wayland 路径在锁定版本中仍有限制，应将 Wayland 标记为实验性或暂缓发布，不能仅凭可编译就宣称正式支持。
- 不包含 Android、iOS、iPadOS、Web、WebAssembly/WASM 等交付目标。
- Windows、macOS、Linux 都是正式目标平台，不应将 macOS/Linux 永久放在“以后有空再支持”的状态；但具体功能可按平台能力分阶段开放。

项目采用“共享核心 + 平台适配器”策略，禁止在业务层直接调用某一操作系统 API。每个发布版本必须提供三平台构建状态、已验证环境、已知限制和安装/升级说明。

Viva README 当前声明 Go 1.26 或更新版本；具体版本应从仓库 README、Release 和 `viva init` 实际生成的 go.mod 再次确认。不要使用浮动的 `@latest` 作为团队长期构建基线。

## 1.3 建议产品范围

### MVP 必做
1. 主时钟与主窗口
2. 闹钟（一次性、每天、按星期重复）
3. 倒计时与秒表
4. 系统托盘和后台生命周期
5. 世界时钟
6. 自适应背景主题系统（本项目重点增强）
7. 用户设置、本地持久化
8. 日志、诊断和自动化测试
9. 安装包和版本信息

### 后续版本
- 天气、RSS
- 本地音乐与网络电台
- 睡眠模式与渐弱
- 桌面悬浮时钟
- 多配置文件
- 时间追踪
- License Key 与自动更新
- 云同步（单独立项）

---

# 2. Viva 官方资源与 Skill 的正确使用方式

## 2.1 必须先理解的资源

- 仓库：<https://github.com/vivaui/viva>
- 中文文档：<https://vivaui.dev/zh/>
- 入门指南：<https://vivaui.dev/zh/docs/getting-started/>
- AI 编程助手指南：<https://vivaui.dev/zh/docs/coding-agents/>
- 主题与字体：<https://vivaui.dev/zh/docs/theming/>
- 绘制与布局：<https://vivaui.dev/zh/docs/drawing-and-layout/>
- 组件指南：<https://vivaui.dev/zh/docs/components/>
- 测试指南：<https://vivaui.dev/zh/docs/testing/>
- 构建与打包：<https://vivaui.dev/zh/docs/build-and-package/>
- Skill 原文：<https://github.com/vivaui/viva/blob/main/docs/skills/building-with-viva/SKILL.md>

仓库内对应的 Skill 路径为：
`docs/skills/building-with-viva/SKILL.md`

## 2.2 Skill 安装

在项目根目录执行（Claude Code 与遵循该目录约定的代理）：

```powershell
New-Item -ItemType Directory -Force .claude/skills/building-with-viva
Invoke-WebRequest `
  https://vivaui.dev/skills/building-with-viva/SKILL.md `
  -OutFile .claude/skills/building-with-viva/SKILL.md
```

同时为通用 Agent 目录准备副本：

```powershell
New-Item -ItemType Directory -Force .agents/skills/building-with-viva
Copy-Item `
  .claude/skills/building-with-viva/SKILL.md `
  .agents/skills/building-with-viva/SKILL.md
```

若使用的 Agent 支持从仓库动态安装 Skill，也可按其官方方式安装，但应将 Skill 文件随项目版本控制。升级 Viva 后重新核对 Skill 版本，避免框架升级而约束仍停留在旧版。

## 2.3 Skill 的使用规则

每次让 AI 修改 Viva UI、布局、输入、窗口或组件时，任务开头必须要求：

1. 先完整读取 `AGENTS.md`。
2. 再完整读取 `building-with-viva/SKILL.md`。
3. 根据任务读取对应 Viva 官方指南和示例，不猜测 API。
4. 先确认功能属于应用层还是可复用框架能力。
5. 优先写 `app.Driver` 测试，再做 headless 测试；确实涉及真实系统窗口、DPI、托盘、拖放、唤醒时才做真实窗口测试。
6. 完成后执行格式化、静态检查、测试和必要的性能基准。
7. 报告真实执行的命令及输出，不允许声称未执行的测试已通过。

## 2.4 不得违反的 Viva 编程规则

- 不按 Web 页面写法设计 UI；没有 CSS、DOM 或浏览器布局引擎。
- 不在 Paint 中修改业务状态（声明控件并读取 `Clicked()` 后的事件分支除外）。
- 不在 Paint 中执行文件 I/O、数据库访问、HTTP 请求、图片解码或昂贵图像处理。
- 不使用永久 goroutine、ticker 或轮询循环维持空闲界面刷新。
- 不使用反射、`map[string]any` 等动态属性包隐藏绘制/布局/输入数据。
- 不手工给 Hit Target 编号；用 `ui.Tree` 生成稳定 ID。
- 不缓存 `ui.Canvas`、`flex.Box` 或 `Children` 回调到 Paint 之外。
- 不在每次绘制中重复解码背景图片。
- 不宣称零分配、低 CPU 或高性能，除非提供可复现的测试结果。
- 不为单一产品功能修改 Viva 框架；只有可被多个应用复用的框架能力才考虑上游贡献。

---

# 3. 开发环境与项目初始化

## 3.1 环境要求

- Go：Viva 当前文档要求 Go 1.26 或更新版本。
- Windows：Win32、Direct3D 11、DirectWrite 系统组件。
- Viva CLI：安装官方 `viva` 命令。
- Git：版本控制。
- IDE：VS Code、GoLand 或其他 Go IDE。
- AI 工具：能够读取项目指令文件和 Skill 的编程代理。

## 3.2 安装 Viva CLI

Windows PowerShell 示例：

```powershell
scoop bucket add vivaui https://github.com/vivaui/homebrew-tap
scoop install viva

viva version
viva doctor
```

若使用 Go 安装方式，按仓库 README 当前给出的命令执行。安装后先执行 `viva doctor`，确认宿主环境满足要求。

## 3.3 创建项目

```powershell
viva init worldclock
Set-Location worldclock

go test ./...
viva dev .
```

`viva init` 会生成窗口示例、测试、`viva.toml` 和 `AGENTS.md`。这些文件是项目真实起点，不要先删除再让 AI 重写。

接下来：
1. 将项目放入 Git。
2. 固定 Viva 版本。
3. 安装 Skill。
4. 审阅 `AGENTS.md`。
5. 建立架构文档和 ADR。
6. 运行最小窗口和 headless 测试。
7. 再开始功能开发。

## 3.4 版本固定

Viva 当前尚未到 1.0，API 可能在版本之间变化。不要假定旧版示例适用于新版本。以 `viva init` 生成的 go.mod 为基准，并将 Viva 依赖和 CLI tool 版本固定。升级时单独开分支，运行全量测试和视觉回归。

## 3.5 viva.toml

以初始化项目生成文件为准。配置中的 `[app]` 至少核对：
- `name`：产品名
- `id`：稳定的应用标识符
- `package`：主程序包路径

不要直接复制其他项目的配置键；Viva 对未知表和键采用严格解码，必须以当前版本构建指南为准。

---

# 4. 项目架构与目录

建议以“产品层模块化单体”实现，不要一开始拆微服务或把每个功能拆成独立进程。

```text
worldclock/
├── AGENTS.md
├── CLAUDE.md                         # 可选，引用 AGENTS.md 与 Skill
├── .agents/
│   └── skills/building-with-viva/SKILL.md
├── .claude/
│   └── skills/building-with-viva/SKILL.md
├── cmd/
│   └── worldclock/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── bootstrap.go
│   │   ├── lifecycle.go
│   │   └── commands.go
│   ├── domain/
│   │   ├── alarm/
│   │   ├── clock/
│   │   ├── worldclock/
│   │   ├── theme/
│   │   └── profile/
│   ├── application/
│   │   ├── alarm/
│   │   ├── scheduler/
│   │   ├── worldclock/
│   │   ├── theme/
│   │   └── settings/
│   ├── infrastructure/
│   │   ├── persistence/
│   │   ├── filesystem/
│   │   ├── weather/
│   │   ├── rss/
│   │   └── logging/
│   ├── platform/
│   │   ├── windows/
│   │   ├── darwin/
│   │   └── linux/
│   ├── media/
│   │   ├── image/
│   │   └── audio/
│   └── ui/
│       ├── shell/
│       ├── pages/
│       ├── widgets/
│       ├── dialogs/
│       ├── theme/
│       └── state/
├── assets/
│   ├── icons/
│   ├── themes/
│   └── defaults/
├── migrations/
├── docs/
│   ├── architecture/
│   ├── decisions/
│   ├── modules/
│   ├── testing/
│   └── operations/
├── scripts/
├── go.mod
├── go.sum
└── viva.toml
```

目录职责：

- `domain`：纯业务实体、规则、值对象，不依赖 Viva UI 和具体数据库。
- `application`：用例、服务接口、调度器协调。
- `infrastructure`：SQLite、文件、网络、图片资源存储。
- `platform`：Windows、macOS、Linux 的平台特定实现，如托盘、电源事件、显示器信息、开机启动和通知。
- `media`：图片解码、缩放、裁剪、缓存和音频资源管理。
- `ui`：Viva Paint、布局、交互和页面状态。
- `cmd`：组合依赖并启动应用。

Go package 不要大量使用 `common`、`utils`、`helpers`、`base` 等含糊名称。接口放在实际使用方附近，只有存在多个实现时才抽象。

---

# 5. 功能模块拆分

## 5.1 MVP 核心模块

| 模块 | 职责 | 优先级 |
|---|---|---|
| App Shell | 主窗口、导航、启动与退出 | P0 |
| Clock | 主时钟、日期、格式 | P0 |
| Alarm | 闹钟规则、铃声、贪睡 | P0 |
| Scheduler | 集中式定时调度 | P0 |
| Timer | 多倒计时、秒表 | P0 |
| World Clock | 城市目录、时区换算 | P0 |
| Theme Engine | 主题令牌与背景图适配 | P0 |
| Settings | 用户偏好与持久化 | P0 |
| Tray | 系统托盘、隐藏/唤醒窗口 | P1，依当前 Viva 平台能力 |
| Popout | 桌面悬浮时钟 | P1 |

## 5.2 后续模块

| 模块 | 主要实现 |
|---|---|
| Weather | 天气 API、缓存、单位转换 |
| RSS | Feed 管理、解析、去重 |
| Audio | 本地播放、网络电台、播放列表 |
| Sleep | 睡眠模式、音量渐弱 |
| Profiles | 多配置文件隔离 |
| Time Tracker | 工作/休息周期与活动检测 |
| Licensing | License Key、签名凭据、离线宽限 |
| Updater | 签名更新 Feed、哈希校验、安装器调用 |

---

# 6. Viva UI 实现规范

## 6.1 基本绘制模型

Viva 的 Paint 回调描述一帧完整 UI。应用状态保存在 model/state 结构中；Paint 每次读取状态并重画。用户动作通过 UI 控件的事件结果触发业务状态变化。

典型模式：

```go
var (
    tree = ui.Tree{}
    page = tree.Feature("main")
    toggleThemeID = tree.ID(page, "toggle-theme")
)

type model struct {
    dark bool
    view flex.View
}

func (m *model) theme() ui.Theme {
    if m.dark {
        return ui.DarkTheme()
    }
    return ui.LightTheme()
}

func newConfig(m *model) app.Config {
    config := app.Config{
        Tree:   &tree,
        Title:  "小瑞综合时钟桌面 (Xiaoru Integrated Clock Desktop)",
        Width:  1100,
        Height: 760,
    }
    config.Paint = func(canvas *ui.Canvas, size ui.Size) {
        theme := m.theme()
        bounds := ui.Rect{Width: size.Width, Height: size.Height}
        canvas.Fill(bounds, theme.Background)

        m.view.Draw(canvas, theme, bounds, func(f *flex.Frame) {
            f.Column().Fill().Pad(24).Gap(16).Children(func() {
                if f.Button(toggleThemeID, "切换主题").Clicked() {
                    m.dark = !m.dark
                }
                f.Text("小瑞综合时钟桌面 (Xiaoru Integrated Clock Desktop)")
            })
        })
    }
    return config
}
```

代码仅展示 Viva 文档中的模式。真实项目应以当前固定版本的入门示例为准。

## 6.2 ID 管理

- 每个页面创建自己的 Feature。
- 每个可交互控件使用 `tree.ID(feature, "meaningful-name")`。
- ID 必须在帧之间稳定。
- 列表项目使用可稳定复现的业务 ID，不使用当前行号作为长期身份。
- 不要手动给 Hit Target 编号。

## 6.3 布局

优先用 `flex.View.Draw` 与 `flex.Frame` 的 Row、Column、Stack、Grid、ScrollArea、Grow、Basis、Min/Max 等规则布局。

避免把整个界面写成硬编码绝对坐标。只有图片裁剪、表盘指针、绘图画布等视觉内容才需要显式几何计算。

## 6.4 组件复用

开发前查看 Viva 组件目录和示例。复用现有按钮、输入框、菜单、Tabs、Dialog、Slider、ScrollArea 等组件。若要自定义组件：
- 使用静态类型 Go API。
- 明确状态输入和事件输出。
- 不引入运行时 CSS/parser。
- 不复制整个组件实现后再私自修改。
- 先检查是否已有可用组件或组件选项。

---

# 7. 背景图片与主题系统：核心增强设计

这是本项目区别于传统桌面时钟软件的重点。不要只做“居中、拉伸、填充”三个下拉框，而应把它设计为独立的 Background Engine。

## 7.1 需要解决的问题

屏幕环境多样：
- 16:9、16:10、21:9、32:9 等宽高比。
- 1080p、1440p、4K 和不同缩放比例。
- 多显示器分辨率、DPI 和方向不同。
- 用户图片可能是横图、竖图、全景图、低分辨率图。
- 图片主体可能位于边缘或中心偏移。
- 同一张图片在不同屏幕上需要不同裁剪位置。
- 用户可能只想显示完整图片，不接受裁剪。
- 背景必须保证前景文字可读。

建议提供“智能适配”作为默认模式，同时保留用户可精确控制的模式。

## 7.2 图像适配模式

每个背景图提供以下选项：

1. **智能填充（Smart Cover）**：按屏幕比例裁剪并填满，不留空白；默认模式。
2. **完整显示（Contain）**：保持原图比例完整展示，剩余区域使用背景色/模糊图填充。
3. **填满裁剪（Cover）**：保持比例放大至覆盖整个区域，允许裁剪。
4. **拉伸（Stretch）**：强制匹配屏幕尺寸，可能造成变形；明确提示。
5. **居中原尺寸（Center）**：以原始像素尺寸居中显示。
6. **平铺（Tile）**：按原尺寸重复铺设。
7. **焦点适配（Focal Cover）**：用户指定主体焦点，裁剪窗口尽量保留该位置。
8. **模糊填充（Blurred Contain）**：背景使用放大模糊图，前景完整显示原图。

## 7.3 适配算法

设：
- 源图像尺寸为 `Sw × Sh`
- 显示区域物理像素尺寸为 `Dw × Dh`
- `scaleX = Dw / Sw`
- `scaleY = Dh / Sh`

Cover：
- `scale = max(scaleX, scaleY)`
- 缩放后尺寸：`Rw = Sw × scale`, `Rh = Sh × scale`
- 溢出：`overflowX = Rw - Dw`, `overflowY = Rh - Dh`
- 裁剪起点受焦点 `fx, fy` 控制，焦点归一化到 0..1。
- `cropX = clamp(fx × Rw - Dw/2, 0, overflowX)`
- `cropY = clamp(fy × Rh - Dh/2, 0, overflowY)`

Contain：
- `scale = min(scaleX, scaleY)`
- 缩放后图像完整可见，剩余空白区域由背景填充策略处理。

Stretch：
- 分别按 X/Y 比例缩放，不保持宽高比。

Center：
- 不缩放；超出区域时可裁切，不足时填背景色。

模糊填充：
1. 生成 Cover 背景层并裁剪到画布。
2. 对背景层低分辨率降采样后做高斯模糊。
3. 叠加半透明遮罩。
4. 叠加 Contain 原图层。
5. 可选加边框或阴影。

所有模式都必须先确定目标显示区域，再计算绘制矩形；不要将图片直接无条件拉伸到窗口大小。

## 7.4 多显示器与 DPI

建立 `DisplayService`，抽象以下信息：
- MonitorID（稳定标识；优先平台设备路径/稳定标识）
- LogicalBounds
- PixelBounds
- ScaleFactor / DPI
- RefreshRate（可选）
- Orientation（横屏/竖屏）
- Primary
- WorkArea
- HDR/色彩信息（后续能力）

背景图片选择策略：
- 按显示器单独指定背景。
- 全部显示器使用同一图像但独立裁剪。
- 跨显示器拼接（高级模式）。
- 断开显示器后把配置保留，重新连接后恢复。
- 显示器分辨率/DPI变化时重新计算目标尺寸和缓存键。

Viva 当前公开 API 是否暴露完整显示器列表、DPI 与工作区，需要逐项核对固定版本 API。若没有，应将其视为平台适配层需求：先判断是否属于跨 Viva 应用通用能力；如果是，可向 Viva 提出功能请求或贡献通用 API。产品特有的图片策略、用户偏好和主题数据留在应用层。不要在业务 Paint 中直接散布 Win32 P/Invoke。

## 7.5 图片资源管线

建立异步管线：

`文件选择/拖放 → 文件校验 → 元数据读取 → 解码 → 方向纠正 → 色彩/尺寸处理 → 生成适配缓存 → UI 发布资源`

要求：
- 图片加载不能阻塞 Paint。
- 解码和缩放在后台执行。
- 支持 PNG、JPEG、WebP 等格式时要通过实际依赖验证，不要凭空承诺。
- 尊重 EXIF Orientation。
- 对极大图像设置像素总量上限，防止内存耗尽。
- 支持文件拖放（若 Viva 当前 Config.OnFileDrop API 可用）。
- 原图保持只读，不覆盖用户文件。
- 缓存为可重建数据，缓存损坏时可以重新生成。
- 取消过期的加载任务，避免快速切换背景时旧结果覆盖新选择。
- 使用 generation/version token：只有最新请求可提交 UI 状态。

## 7.6 缓存策略

缓存键至少包含：
- AssetHash（原图哈希）
- MonitorID 或目标显示区域尺寸
- DPI/ScaleFactor
- FitMode
- FocalPoint
- BlurRadius
- Overlay/Color treatment
- Renderer/格式版本

缓存结果包括：
- 已解码原图（受内存预算控制）
- 适配后的缩放图或纹理
- 模糊图
- 缩略图

缓存应采用 LRU 或容量上限策略。不要把所有原始大图永久留在内存。屏幕尺寸或焦点变化时只重建对应派生资源。

Viva 的 Canvas 是否直接支持目标图片格式、纹理生命周期与跨窗口资源复用，必须依据固定版本 API 检查。建议将图像解码与适配算法独立成 `internal/media/image`，再通过一个薄的 Viva 绘制适配器展示。不要臆造不存在的 `Canvas.DrawImage` 或纹理 API。

## 7.7 用户体验设置项

背景设置页面建议分为：

**图片来源**
- 选择图片
- 选择文件夹
- 添加多张图片
- 拖放图片
- 预览当前显示器效果

**布局方式**
- 智能填充
- 完整显示
- 填满裁剪
- 拉伸
- 居中原尺寸
- 平铺
- 模糊填充

**焦点与裁剪**
- 焦点位置 X/Y（0..100%）
- 鼠标拖动预览裁剪区域
- 重置焦点
- 锁定当前裁剪
- 对所有显示器应用/单独设置

**画面增强**
- 背景模糊
- 暗色遮罩
- 亮度
- 饱和度
- 对比度
- 渐变遮罩
- 前景文字阴影或衬底

**显示器**
- 当前显示器选择
- 每台显示器独立预览
- 全部显示器统一应用
- 分辨率/DPI 变化后的重新适配

## 7.8 预览交互

设置页最好有：
- 左侧图片库/主题库
- 中间大尺寸预览
- 右侧适配参数
- 显示器切换
- 预览缩放（Fit/100%）
- 拖动焦点
- “应用”“取消”“恢复默认”

预览尺寸与实际主窗口或屏幕的宽高比一致。预览区域变化时只更新适配参数，不重新读取磁盘原图。用户点击应用后再持久化配置。

## 7.9 主题与背景分离

`Theme` 负责颜色、字体和控件视觉 token；`BackgroundProfile` 负责图片、裁剪、焦点、蒙版和显示器规则。不要把图片文件路径塞进 `ui.Theme`。

建议模型：

```go
type FitMode string

const (
    FitSmartCover FitMode = "smart-cover"
    FitContain    FitMode = "contain"
    FitCover      FitMode = "cover"
    FitStretch    FitMode = "stretch"
    FitCenter     FitMode = "center"
    FitTile       FitMode = "tile"
    FitBlurred    FitMode = "blurred-contain"
)

type FocalPoint struct {
    X float32 // normalized 0..1
    Y float32 // normalized 0..1
}

type BackgroundProfile struct {
    ID            string
    AssetID       string
    FitMode       FitMode
    FocalPoint    FocalPoint
    BlurRadius    float32
    OverlayAlpha  float32
    Brightness    float32
    Saturation    float32
    PerMonitor    bool
}
```

配置结构需校验范围；持久化前对路径做权限/存在性检查，但不要要求图片必须一直处于某个固定目录。

## 7.10 文字可读性

图片背景下内容可读性不可依赖用户自行判断。建议：
- 自动计算背景亮度/对比度作为提示。
- 允许全局暗遮罩/亮遮罩。
- 为时钟、日期和关键数据使用局部渐变底。
- 重要文本提供描边/阴影。
- 主题预览时展示实际文字样例。
- 无法可靠估算时允许用户手动调节，不要宣称满足某个对比度标准而未测量。

## 7.11 自动适配与用户控制的优先级

建议优先级：
1. 用户针对当前显示器的显式配置
2. 用户针对全局的配置
3. 图片元数据/焦点
4. 智能默认策略

显示器变化时：
- 保留用户选择的 FitMode/FocalPoint。
- 重新计算目标裁剪。
- 生成新的缓存项。
- 不修改原图。
- 预览和正式界面使用同一套计算函数，保证结果一致。

---

# 8. 世界时钟

## 8.1 城市与时区
- 城市目录内置 JSON 或嵌入资源。
- 城市记录包含名称、国家码、经纬度和 IANA timezone ID。
- 收藏列表单独存储。
- 时间换算基于同一 UTC snapshot。
- 使用标准时区库/系统数据处理 DST，禁止硬编码 UTC offset。
- 无效时区要显示可诊断状态，不得导致整页崩溃。

## 8.2 卡片展示
- 城市名称、国家
- 本地时间、日期、星期
- UTC offset
- 与本机时差
- 可选天气
- 排序、收藏、删除

## 8.3 刷新
时间展示只在需要秒时每秒触发短时更新；没有秒显示时可以降低频率。Viva 不允许为了画面刷新常驻 heartbeat。采用应用层调度器/事件失效机制，只在时间文本即将变化时安排下一次必要的绘制，并在窗口隐藏、最小化或页面不可见时停止无意义刷新。

---

# 9. 闹钟、计时器和调度

## 9.1 时间语义
- 一次性闹钟：保存明确的 UTC instant 及用户所见时区信息。
- 每日/每周闹钟：保存当地墙上时间、重复规则和时区策略。
- DST 无效时间/歧义时间必须有明确处理策略。
- 不以 UI ticker 的 Tick 次数作为真实时间。
- 统一 Scheduler 维护下一次到期项；到期后重新计算下次发生时间。

## 9.2 Scheduler
- 使用一个统一调度组件，而不是每个闹钟一个 goroutine。
- 允许取消、暂停、恢复和重建。
- 处理系统时间变化、时区变化、睡眠恢复。
- 将触发记录持久化，避免重启或重复事件造成双响。
- 触发音频、通知或动作通过接口派发。
- 后台调度服务与 Paint 完全分离。

## 9.3 倒计时
- 多个独立实例
- 开始/暂停/继续/重置
- 使用单调时钟计算经过时间
- 过期只触发一次
- 可选通知/声音/置顶小窗

## 9.4 秒表
- 使用单调时钟
- 支持 Lap 分段
- UI 显示格式化累计时间
- 不因窗口重绘频率而产生时间漂移

## 9.5 系统唤醒
电源唤醒能力与具体硬件/操作系统行为需要分别核实。Windows、macOS、Linux 的睡眠唤醒机制和权限条件不同，不能用单一平台实现推断其他平台行为。必须通过平台适配、能力检测和实机验证；不支持或无法确认时应明确禁用相关选项，不得显示“保证唤醒”。

---

# 10. 数据存储与配置

## 10.1 数据存储选型
Viva 是 UI 框架，不强制数据库。建议：
- SQLite：持久化闹钟、设置、城市收藏和 Profile。
- JSON/TOML：少量可读配置和内置城市目录。
- 文件系统：用户图片、音频。
- Cache：可删除、可重建的派生资源。

Go SQLite driver 需要根据目标平台、cgo 与静态/动态构建要求选择。先用最小实验分别确认 Windows、macOS、Linux 的构建、打包、运行时依赖和许可证，再确定数据库驱动。不要未经评估就引入大 ORM；可使用 `database/sql` + migration 工具，或轻量查询生成工具。

## 10.2 建议数据表
- `profiles`
- `app_settings`
- `alarms`
- `alarm_schedules`
- `alarm_occurrences`
- `timers`
- `world_clock_cities`
- `background_assets`
- `background_profiles`
- `monitor_backgrounds`
- `themes`
- `theme_tokens`
- `weather_cache`
- `rss_feeds`
- `rss_articles`
- `playlists`
- `playlist_items`
- `time_tracker_sessions`
- `license_state`

所有时间字段统一明确 UTC 或本地语义；路径使用相对/可迁移引用策略；数据库迁移必须有版本号和升级测试。

## 10.3 背景数据表建议
`background_assets`：
- id
- original_path
- content_hash
- source_filename
- source_width
- source_height
- orientation
- imported_at_utc
- last_seen_at_utc

`background_profiles`：
- id
- theme_id
- asset_id
- fit_mode
- focal_x
- focal_y
- blur_radius
- overlay_alpha
- brightness
- saturation
- enabled

`monitor_backgrounds`：
- id
- profile_id
- monitor_key
- fit_mode_override
- focal_x_override
- focal_y_override
- last_known_width
- last_known_height
- last_known_scale

---

# 11. 平台适配设计

## 11.1 平台接口
平台能力统一定义接口：
- `DisplayService`
- `TrayService`
- `StartupService`
- `PowerEventService`
- `WindowService`
- `MediaPicker`
- `NotificationService`

Windows、macOS、Linux 分别实现适配器，不要把 Win32、Cocoa、Linux/X11/Wayland API 调用散布到 Domain/Application/UI。

## 11.2 Viva 已有能力优先
先查看当前 Viva API 是否已提供：
- 多窗口
- 窗口隐藏/显示
- 状态区图标/托盘
- 文件拖放
- 窗口全屏/置顶相关控制
- 文件选择/图片资源
- 系统外观或窗口材质
- 系统通知

若已有稳定 API，直接使用；若缺失：
1. 判断是否通用框架能力。
2. 写最小可复现用例。
3. 检查是否可以在应用层封装。
4. 只有跨应用通用能力才向 Viva 上游提需求/贡献。
5. 产品特有策略保持在应用内部。

## 11.3 不要假设 API 存在
Viva 当前 API 与状态在持续发展。任何代码生成前，应从固定版本源码/官方 API 文档核实签名。AI 不得编造 `Canvas.DrawImage`、显示器枚举、系统主题检测或窗口置顶接口。

---

# 12. 测试策略

## 12.1 测试层
1. Domain 单元测试：纯规则和图像裁剪数学。
2. Application 测试：用例、调度、数据协调。
3. `app.Driver`：驱动真实 painter、点击、键盘、输入和滚动；不必打开窗口。
4. Headless：`vivahead` build tag、`VIVA_HEADLESS=1`、`shot`/`layout`/`tree`/`expect` 等协议。
5. 实机测试：DPI、多显示器、窗口管理器、拖放、托盘、睡眠唤醒、IME 等。

## 12.2 背景图片算法测试
至少覆盖：
- 16:9 图片到 16:9 画布
- 16:9 到 16:10
- 16:9 到 21:9
- 竖图到横屏
- 横图到竖屏
- 超小图放大
- 超大图缩小
- FocalPoint 四角和中心
- FocalPoint 边界 clamp
- 图片 EXIF 旋转
- DPI 100%、125%、150%、200%
- 显示器分辨率变化
- 双屏不同宽高比
- 多屏断开再连接
- 缓存命中、失效、容量限制
- 快速切换背景时旧任务不得覆盖新选择

测试只验证算法与契约，不要把 GPU 的精确像素计数写成脆弱的断言。视觉结果使用固定测试图片和截图对比，但需记录 OS/GPU/缩放环境。

## 12.3 Viva 验证命令
根据 Skill，基础门禁：
```sh
gofmt -w .
go vet ./...
go test ./... -count=1
```

注意：`gofmt -w .` 不是所有仓库都适合的广泛命令；实际执行建议使用：
```sh
gofmt -w $(git ls-files '*.go')
test -z "$(gofmt -l $(git ls-files '*.go'))"
go vet ./...
go test ./... -count=1
```

Paint/Input/Layout 路径还要加 benchmark；涉及平台后端、显示器信息、文件拖放、托盘和通知时，必须在 Windows、macOS、Linux 的实际目标环境分别测试。

## 12.5 三平台发布验收矩阵

每次正式发布前，至少完成以下检查。不能用 CI 编译成功替代真实桌面环境验收。

| 验收项 | Windows | macOS | Linux |
|---|---|---|---|
| 原生窗口启动/退出 | 必测 | 必测 | 必测 |
| 主界面绘制与字体 | 必测 | 必测 | 必测 |
| 图片解码、裁剪、缩放 | 必测 | 必测 | 必测 |
| 多显示器与 DPI/缩放 | 必测 | 必测 | 必测 |
| 文件选择与拖放 | 必测 | 必测 | 必测 |
| 系统托盘 | 必测 | 必测 | 按桌面环境验证 |
| 开机启动 | 必测 | 必测 | 按桌面环境验证 |
| 通知 | 必测 | 必测 | 按桌面环境验证 |
| 睡眠/唤醒及系统时间变化 | 必测 | 必测 | 必测 |
| 安装、升级、卸载 | 必测 | 必测 | 必测 |

Linux 必须区分 X11 与 Wayland 的测试结果；macOS 必须记录 Intel/Apple Silicon 架构；Windows 必须记录 x64/ARM64 架构。某项能力若在平台上不可用，应通过能力探测隐藏或禁用该功能，并在发布说明中列出，不得让应用崩溃或静默失效。

## 12.4 Headless 验收
- 为主窗口建立 `app.Driver` 测试。
- 对背景设置流程验证控件可达、按钮动作、状态更新和可访问名称。
- 运行 headless `layout` 检查文字裁切与重叠。
- 用 `shot` 生成截图供人工审查。
- Headless 无法验证真实 GPU 图片采样、显示器 DPI 和多屏拓扑，不能代替实机测试。

---

# 13. AI 编程工作流

## 13.1 规则文件
根目录 `AGENTS.md` 应包含：
- 产品范围
- 架构边界
- 依赖规则
- 数据库迁移规则
- 测试命令
- Viva Skill 引用
- 代码风格
- 禁止事项
- Definition of Done

`CLAUDE.md` 等工具专属文件只引用根 `AGENTS.md` 和对应 Skill，不维护另一套互相冲突的规范。

## 13.2 每次任务模板
将以下模板放在每个 AI 任务提示词中：

```text
你正在开发一个基于 Go + Viva 的原生桌面应用。

开始前：
1. 阅读根目录 AGENTS.md。
2. 阅读 .agents/skills/building-with-viva/SKILL.md。
3. 阅读与任务直接相关的 Viva 官方指南和 examples。
4. 核实当前 go.mod 固定的 Viva 版本和 API；禁止猜 API。
5. 先说明任务属于 Domain、Application、Infrastructure、Platform 还是 UI。

约束：
- 遵守 Viva Paint、Tree ID、flex、事件驱动重绘和无空闲轮询规则。
- Paint 不执行 I/O、网络、数据库、图片解码或业务副作用。
- 不使用 WebView/HTML/CSS 来实现主 UI。
- 不引入未经评估的依赖。
- 不得擅自修改公共接口和数据库结构。
- 图片资源通过异步加载管线准备，UI 只读取已准备资源。

执行：
1. 先阅读现有代码，不要覆盖性重写。
2. 给出最小实施方案。
3. 编写或更新测试。
4. 实现代码。
5. 运行 gofmt、go vet、go test。
6. 对 Paint/Layout/Input 性能敏感改动执行基准。
7. 报告实际修改文件、命令和结果、未测试的限制。

不得伪造测试结果。
```

## 13.3 多 AI 协同
可以按模块并行，但必须避免多个 Agent 同时编辑同一公共文件：
- Agent A：背景图片算法与测试。
- Agent B：城市/时区/世界时钟。
- Agent C：闹钟 Domain 与 Scheduler。
- Agent D：UI 页面和交互，在接口稳定后接入。
- 主 Agent：维护架构、集成、迁移、构建、评审。

先由主 Agent 定义模型和接口，再分派任务。每个任务结束必须提交测试与变更摘要。

---

# 14. 开发阶段与实施顺序

## 阶段 0：技术验证 Spike
目的：确认 Viva 版本、CLI、三平台窗口后端、组件、绘图、Headless 测试，以及图片解码/绘制链路。
交付：
- 固定版本的最小 Viva 应用
- Skill 安装
- `AGENTS.md`
- 一张图片能够加载并显示的技术验证
- Windows x64 构建测试
- macOS 目标架构构建测试
- Linux x86_64 构建测试（至少验证选定的 X11 环境；Wayland 单独记录结果）
未完成此阶段前不要全面开发业务模块。

## 阶段 1：项目骨架
- 目录与模块
- 日志
- 配置
- SQLite migration
- 主窗口/导航
- `app.Driver` 测试

## 阶段 2：核心时间能力
- 主时钟
- Scheduler
- 闹钟 CRUD
- 倒计时
- 秒表
- 世界时钟

## 阶段 3：背景与主题引擎
- 图片导入与元数据
- 适配算法
- 焦点裁剪
- 模糊填充
- 多显示器配置模型
- 缓存
- 预览 UI
- 显示器变化时重新适配

## 阶段 4：系统集成
- 托盘
- 自动启动
- 通知
- Popout
- 文件拖放
- 睡眠唤醒能力探测

## 阶段 5：增强功能
- 天气
- RSS
- 音频/电台
- 睡眠模式
- 多 Profile
- 时间追踪

## 阶段 6：发布
- 安装包
- 更新
- 授权
- 签名
- 崩溃与诊断
- 性能与资源预算
- Windows、macOS、Linux 实机验收（Linux 的 X11/Wayland 分开记录）

---

# 15. 关键风险和工程决策

| 风险 | 处理 |
|---|---|
| Viva API 尚未稳定到 1.0 | 固定版本；升级单独分支验证 |
| 图片 API/纹理生命周期不明确 | 先做 Spike；不假设 Canvas 有某方法 |
| 多显示器/DPI API 不足 | 平台接口隔离；评估是否上游贡献 |
| 高频时钟更新 | 只在显示需要时调度；无空闲 heartbeat |
| 大图片内存峰值 | 像素上限、异步解码、缓存预算 |
| 图片切换竞态 | generation token/cancel context |
| DPI 与逻辑像素混用 | 明确逻辑单位/物理像素单位 |
| 桌面主题可读性 | 遮罩/渐变/阴影及实际截图验收 |
| 睡眠唤醒不可靠 | 能力检测与实机说明，不作绝对承诺 |
| 依赖使二进制变大 | 逐项测量二进制、启动和 idle 内存 |

必须记录的 ADR：
- ADR-001 Viva 版本与目标平台
- ADR-002 Go UI 结构与模块边界
- ADR-003 数据库和迁移策略
- ADR-004 图片解码库与格式支持
- ADR-005 图片适配与缓存模型
- ADR-006 显示器/DPI API 来源
- ADR-007 Scheduler 与 DST 语义
- ADR-008 系统托盘与开机启动实现

---

# 16. 首个 Sprint（建议）

**Sprint 1：Viva 技术验证与最小产品壳**

任务 1：环境与工具
- 安装 Viva CLI
- `viva doctor`
- `viva init worldclock`
- 固定版本
- 安装 Skill
- Git 初始化

任务 2：阅读框架约束
- 审查生成的 AGENTS.md
- 阅读 Skill
- 阅读 getting-started、programming-model、drawing-and-layout、theming、testing
- 运行官方 showcase 与组件预览

任务 3：UI 骨架
- 主窗口
- 导航区域
- 时钟占位区域
- 设置入口
- 事件驱动交互
- `app.Driver` 测试

任务 4：图片技术验证
- 选择一张 4K 横图、一张竖图、一张超宽图
- 证明图片能在当前 Viva 固定版本中解码和绘制
- 测量内存峰值
- 验证 Cover/Contain 数学
- 记录是否需要平台专用纹理 API

任务 5：验收
- `gofmt`
- `go vet ./...`
- `go test ./... -count=1`
- Windows、macOS、Linux 实际窗口运行
- Linux 先完成 X11 验收，再单独验证 Wayland
- 三个平台至少各完成一次 DPI/缩放与窗口 resize 验证

Sprint 1 不做天气、RSS、授权和云同步。

---

# 17. 交付标准（Definition of Done）

- 需求和验收标准已记录。
- 所有新增代码符合模块边界。
- 公共 API 有注释和示例。
- 状态变化与 Paint 副作用分离。
- 图片、网络与数据库操作不在 Paint 执行。
- 新数据结构有迁移策略。
- 关键算法有单元测试。
- UI 有 app.Driver 或 Headless 测试。
- UI 变更有截图/布局检查方法。
- 性能敏感代码有基准结果。
- Windows、macOS、Linux 三个平台均完成真实运行验证。
- 每个平台记录操作系统版本、CPU 架构、桌面环境/显示服务器、GPU 后端和缩放比例。
- 记录未测试的架构、发行版、桌面环境及限制。
- 文档已同步。

---

# 18. 官方参考与本文件依据

以编写文档时仓库可见内容为依据，正式实施前应再次检查最新版本：

1. Viva GitHub 仓库：<https://github.com/vivaui/viva>
2. Viva 中文 README：<https://github.com/vivaui/viva/blob/main/README.zh-CN.md>
3. Viva Skill：<https://github.com/vivaui/viva/blob/main/docs/skills/building-with-viva/SKILL.md>
4. AI 编程助手指南：<https://vivaui.dev/zh/docs/coding-agents/>
5. 开始使用：<https://vivaui.dev/zh/docs/getting-started/>
6. 主题与字体：<https://vivaui.dev/zh/docs/theming/>
7. 绘制与布局：<https://vivaui.dev/zh/docs/drawing-and-layout/>
8. 测试指南：<https://vivaui.dev/zh/docs/testing/>
9. Windows 材质说明：<https://github.com/vivaui/viva/blob/main/docs/windows-materials.md>

**关键事实：** Viva 主题指南中的 `ui.Theme` 是语义颜色和尺寸 token，不是桌面壁纸/背景图片管理器。它没有因为有 Theme 就自动具备完整图片适配系统。因此本项目应把背景图片作为独立媒体子系统实现，再由 Paint 在明确的绘制区域中消费处理结果。

---

# 附录 A：建议根目录 AGENTS.md

```markdown
# 小瑞综合时钟桌面 (Xiaoru Integrated Clock Desktop) — Project Instructions

This is a Go desktop application built with Viva.

Before changing code:
1. Read this file.
2. Read `.agents/skills/building-with-viva/SKILL.md`.
3. Read the relevant Viva guide and current-version examples.
4. Verify APIs against the pinned Viva dependency; never invent APIs.

Architecture:
- Domain contains pure business rules.
- Application contains use cases and service orchestration.
- Infrastructure owns persistence, files, network and media decoding.
- Platform owns OS-specific integrations.
- UI owns Viva Paint, layout, input and presentation state.
- Keep product behavior in this repository; only reusable framework capabilities belong upstream.

Viva rules:
- Use `ui.Tree` stable identities.
- Use `flex.View`/`flex.Frame` for layout.
- Paint is a pure description of the current frame except handling a delivered control event.
- Never do DB, file, HTTP, image decoding or expensive processing in Paint.
- Do not create idle polling loops or permanent tickers to redraw the UI.
- No HTML/CSS/WebView for the primary application UI.
- Do not invent framework methods.
- Verify headless behavior and real platform behavior separately.

Images and themes:
- Keep Theme tokens separate from BackgroundProfile.
- Image decode and transform happen asynchronously outside Paint.
- Share the exact same fit/crop math between preview and final display.
- Respect DPI and monitor-specific physical pixel bounds.
- Cache derived images with bounded memory and invalidation.
- Old async image loads must not overwrite a newer selection.

Before completion:
- gofmt
- go vet ./...
- go test ./... -count=1
- Add tests for new behavior.
- Report commands actually run and results actually observed.
- State platform cases not tested.
```

# 附录 B：背景图片适配算法伪代码

```go
type Size struct {
    W float64
    H float64
}

type Point struct {
    X float64 // 0..1
    Y float64 // 0..1
}

type Crop struct {
    SourceX float64
    SourceY float64
    SourceW float64
    SourceH float64
    Scale   float64
}

func CoverCrop(src, dst Size, focus Point) Crop {
    scaleX := dst.W / src.W
    scaleY := dst.H / src.H
    scale := math.Max(scaleX, scaleY)

    renderedW := src.W * scale
    renderedH := src.H * scale

    overflowX := math.Max(0, renderedW-dst.W)
    overflowY := math.Max(0, renderedH-dst.H)

    fx := clamp(focus.X, 0, 1)
    fy := clamp(focus.Y, 0, 1)

    cropRenderedX := clamp(fx*renderedW-dst.W/2, 0, overflowX)
    cropRenderedY := clamp(fy*renderedH-dst.H/2, 0, overflowY)

    return Crop{
        SourceX: cropRenderedX / scale,
        SourceY: cropRenderedY / scale,
        SourceW: dst.W / scale,
        SourceH: dst.H / scale,
        Scale: scale,
    }
}
```

注意：此代码是算法示意；生产代码需处理零尺寸、NaN/Inf、尺寸上限、EXIF 方向和浮点误差，并通过测试验证。实际 Viva 绘制调用应使用固定版本中确实存在的图片/纹理 API。

**文档结束**
