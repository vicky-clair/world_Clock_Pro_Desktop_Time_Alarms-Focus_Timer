# Olive Clock 同类产品完整开发技术设计文档

**版本：1.0 | 日期：2026-11-05 | 平台：Windows 10/11 x64**


# 文档说明与边界

本文件是一份面向独立开发者/小型团队的 Olive Clock 同类产品完整开发技术设计文档。它以 Olive Clock 官网、公开功能页及官方使用文档中可观察到的功能为需求参考，给出一套可实际实施的重新设计方案。

重要边界：
- 本文不是 Olive Clock 官方内部设计文档，也不代表其真实源代码、数据库结构、算法或内部服务实现。
- 公开资料可以确认产品功能与部分用户操作行为；无法确认的实现细节均标注为“建议方案”或“工程推断”。
- 文档的目标是实现功能相近、架构独立的 Windows 桌面时间管理产品，不复制其商标、图标、界面资产、专有代码或受版权保护的内容。
- 第一阶段目标平台为 Windows 10/11 x64。若未来跨平台，应重新评估 WPF、系统电源管理、音频和桌面窗口实现。


# 1. 产品定位与范围

## 1.1 产品定位
将产品定位为 Windows 桌面综合时间管理工具：以主时钟为中心，组合闹钟与日程、倒计时、秒表、桌面悬浮时钟、天气、世界时钟、RSS 新闻、音乐/网络电台、主题/照片轮播、睡眠模式、时间追踪、多配置文件及个性化设置。

## 1.2 公开资料确认的功能范围
官方功能页公开列出：
- Daily、One Time、Recurring 闹钟；自定义音频、广播电台或播放列表；自动停止/自动关闭；睡眠唤醒。
- 闹钟动作：启动程序、打开 URL、发送 GET 请求。
- 桌面 Popout Clock，支持置顶。
- 多个计时器与秒表，支持置顶。
- 当前地点天气和温度；多城市世界时钟及当地天气。
- 内置新闻类别与自定义 RSS。
- 音乐和电台播放列表。
- 内置主题、自定义照片相册、背景轮播。
- 多配置文件，配置文件隔离闹钟、主题和设置。
- 多种模拟时钟表盘。
- 睡眠模式、睡眠音乐和渐弱停止。
- 时间追踪，支持基于计时与基于活动的模式。

## 1.3 本项目范围
必须实现（核心版本）：
1. 主时钟与窗口设置
2. 闹钟与重复规则
3. 倒计时与秒表
4. 系统托盘、开机启动、后台生命周期
5. Popout 悬浮时钟与置顶
6. SQLite 本地持久化
7. 配置文件
8. 基础主题和用户设置
9. 世界时钟

增强版本：
10. 天气
11. RSS 新闻
12. 本地音乐与网络电台
13. 睡眠模式和渐弱
14. 闹钟动作
15. 时间追踪
16. 睡眠唤醒
17. 授权、更新、安装和诊断

## 1.4 非目标
- 不承诺在关机状态下启动电脑。
- 不承诺所有硬件/固件/电源状态都支持定时唤醒。
- 不实现云账户和跨设备同步，除非另立项目。
- 不把天气、RSS、网络电台作为核心时钟的运行依赖。


# 2. 技术栈与架构决策

## 2.1 推荐技术栈
| 层次 | 技术建议 | 说明 |
|---|---|---|
| 语言 | C# | 主开发语言 |
| Runtime | 当前受支持的 .NET LTS | 固定 SDK 与 TargetFramework |
| 桌面 UI | WPF | Windows 原生桌面窗口、数据绑定、模板 |
| 架构 | MVVM | View 与业务逻辑分离 |
| MVVM 工具 | CommunityToolkit.Mvvm | ObservableObject、RelayCommand |
| DI/配置/日志 | Microsoft.Extensions.* | 依赖注入、Options、Logging |
| ORM | EF Core | 数据访问与迁移 |
| 数据库 | SQLite | 单机本地数据 |
| 时区 | TimeZoneInfo + TimeZoneConverter | IANA/Windows ID 兼容 |
| 音频 | NAudio；必要时 LibVLCSharp | 本地音频/流媒体按格式能力选择 |
| RSS | System.ServiceModel.Syndication 或成熟解析库 | RSS/Atom |
| HTTP | HttpClient/IHttpClientFactory | 天气、RSS、授权、更新 |
| JSON | System.Text.Json | 配置与 API |
| 测试 | xUnit + FluentAssertions（可选） | 单元与集成测试 |
| 安装 | MSIX 或 Inno Setup | 按分发渠道选择 |
| CI | GitHub Actions/Azure Pipelines | 构建、测试、打包 |

## 2.2 架构原则
- 分层：Presentation、Application、Domain、Infrastructure。
- 依赖方向向内：Domain 不依赖 WPF、SQLite、HTTP。
- 业务服务通过接口暴露。
- UI 不直接访问 DbContext、文件系统或网络。
- 时间调度集中管理，不能每个闹钟创建永久循环。
- 外部数据源故障不得阻塞核心时钟和闹钟服务。
- 所有持久化模型应带有版本迁移策略。
- 对系统时间变化、时区规则更新、睡眠恢复进行显式处理。

## 2.3 解决方案结构
```text
OliveLikeClock.sln
src/
  Clock.App/                 # WPF Views、ViewModels、Resources、Controls
  Clock.Domain/              # Entities、ValueObjects、Enums、Domain rules
  Clock.Application/         # UseCases、DTOs、Interfaces、Validators
  Clock.Infrastructure/      # EF Core、SQLite、HTTP、Audio、RSS、OS adapters
  Clock.Windows/              # Win32/PInvoke、托盘、启动项、电源管理
tests/
  Clock.Domain.Tests/
  Clock.Application.Tests/
  Clock.Infrastructure.Tests/
  Clock.Windows.Tests/
docs/
  architecture/
  decisions/
  operations/
```

## 2.4 关键模块边界
- ClockDisplay：显示时间，不负责闹钟调度。
- Alarm：闹钟规则、下一次触发时间和触发状态。
- Scheduler：统一调度到期任务。
- Audio：播放/停止/音量渐变，不负责决定何时触发。
- Weather/RSS：独立后台刷新服务。
- Profile：隔离用户偏好和业务数据。
- Licensing：授权状态与功能门控。


# 3. 功能需求与验收

## 3.1 主时钟
需求：
- 显示本机当地时间、日期、星期。
- 12/24 小时制；可显示/隐藏秒和完整日期。
- 数字时钟与模拟表盘主题。
- 可设置时钟位置、大小、前景颜色/透明度、背景透明度。
- 支持窗口大小与位置记忆、最大化、全屏、启动时最小化到托盘、隐藏菜单栏。
验收：
- 系统时间改变后，界面在合理时间内刷新。
- 多显示器 DPI 缩放下不出现窗口定位错误。
- 关闭主窗口时按用户设置退出或隐藏到托盘。

## 3.2 闹钟
需求：
- 新建、编辑、删除、启用/禁用。
- 一次性、每日、按星期重复、每月某日、每隔 N 分钟/小时。
- 标题、铃声来源、音量渐增、响铃停止时间、贪睡、自动关闭。
- 静音闹钟/仅执行动作。
- 可选启动程序、打开 URL、HTTP GET。
- 可选睡眠/休眠唤醒。
验收：
- 重启应用后恢复启用闹钟。
- 重复规则在跨月、闰年、夏令时转换时有定义明确的行为。
- 同一触发实例只执行一次（幂等）。
- 错过闹钟有明确策略：忽略、启动时补触发或提示用户。

## 3.3 计时器与秒表
计时器：
- 多实例、名称、时长、开始/暂停/继续/重置/取消。
- 到期事件、声音、通知、置顶小窗。
秒表：
- 开始/暂停/继续/重置。
- Lap 分段记录。
- 应用重启后的恢复策略由产品决定；建议默认不恢复运行状态，但保留最后记录。
验收：
- 使用 Stopwatch/单调时钟测量时间间隔，不通过每秒累加避免漂移。
- UI 卡顿不改变实际经过时间。

## 3.4 Popout Clock
- 独立 WPF Window。
- Topmost 置顶开关。
- 无边框、可调整大小、透明背景。
- 可设置数字/模拟样式、颜色、透明度、秒和日期。
- 保存位置、尺寸、显示器标识。
- 多显示器断开时自动将窗口恢复到可见区域。

## 3.5 世界时钟
- 城市目录、城市搜索、添加/删除/排序/收藏。
- 每个城市显示当地时间、日期、星期、UTC offset。
- 使用 IANA 时区标识，正确处理 DST。
- 时间离线可用。
- 城市时钟共用同一个 UTC 时间快照。
- 天气信息通过独立天气服务获取，不影响时钟刷新。

## 3.6 天气
- 当前地点天气和温度；世界时钟城市天气为可选扩展。
- 摄氏/华氏单位。
- 定时刷新、缓存、超时、取消和失败退避。
- 显示数据更新时间；缓存过期时标记陈旧数据。
- 地理位置须经用户许可；允许手动选择城市。
- API Key 不得硬编码在客户端源代码中。

## 3.7 RSS 新闻
- 内置新闻源（若有合法、稳定的聚合 API）和自定义 RSS/Atom。
- 添加、验证、编辑、删除 Feed。
- 定时刷新、条目去重、保留数量限制。
- 标题、摘要、发布时间、链接。
- 点击条目通过系统默认浏览器打开。
- 防止恶意 XML：禁用外部实体/DTD，限制响应体大小和超时。

## 3.8 音乐与网络电台
- 播放列表、曲目/电台条目、添加/删除/排序。
- 本地文件、HTTP(S) 流媒体。
- 播放、暂停、停止、上一首、下一首、音量。
- 闹钟可引用单个铃声、广播站或播放列表。
- 媒体不可用时显示错误并回退到默认铃声。
- 明确处理版权：软件不捆绑未经授权的音乐内容。

## 3.9 主题与照片相册
- 内置主题与自定义主题。
- 用户选择图片目录或图片文件。
- 背景图片轮播与切换间隔。
- 前景颜色/透明度、背景矩形透明度。
- 图片加载失败时跳过并记录日志。
- 大图异步解码、缩略图缓存，避免 UI 线程阻塞。

## 3.10 睡眠模式
- 进入专注/低干扰时钟界面。
- 播放指定睡眠音乐播放列表。
- 设置睡眠时长。
- 结束前渐弱音量并停止播放。
- 支持用户提前退出。
- 睡眠模式不能覆盖闹钟的优先级；临近闹钟时应按策略停止睡眠音频并触发闹钟。

## 3.11 多配置文件
- 创建、重命名、切换、删除 Profile。
- 闹钟、主题、播放列表、时钟偏好等按 Profile 隔离。
- 系统级设置（启动项、更新、日志目录等）可以全局共享。
- 删除 Profile 前确认并提供导出/备份选项。
- 当前 Profile 切换必须原子化，防止旧配置数据串入新配置。

## 3.12 时间追踪
- Timer-based：按工作时间累计。
- Activity-based：检测键鼠活动，空闲超过阈值后暂停。
- 短休息、大休息、停止工作时间。
- 短休息与长休息冲突时长休息优先。
- 提前提示、休息窗口、延后、结束休息。
- 勿扰模式。
- 可配置工作日、开始/结束时间。
- 可选系统动作：锁定、注销、睡眠、重启、关机（必须二次确认与清晰提示）。
- 活动检测应遵循最小权限原则，不记录键盘内容，不采集具体按键。

## 3.13 设置与账户/授权
- 设置分组：Startup、Clock、Popout Clock、Weather、News、Audio、Profiles、Time Tracker、License、About。
- 免费/专业功能边界由产品配置管理。
- v2 风格采用 License Key，不强制注册登录可作为产品决策参考。
- 授权验证需支持离线宽限策略、设备绑定策略和撤销策略。
- 不在客户端保存明文支付信息。


# 4. 核心架构与运行生命周期

## 4.1 应用生命周期
1. 单实例检测。
2. 加载配置与数据库。
3. 执行数据库迁移。
4. 加载 Profile。
5. 初始化系统托盘。
6. 恢复启用闹钟并计算下一次触发。
7. 初始化 Scheduler。
8. 按设置显示主窗口或最小化到托盘。
9. 后台启动天气/RSS 刷新（带取消令牌）。
10. 退出时保存窗口状态、停止后台任务、释放音频/HTTP/数据库资源。

## 4.2 单实例
使用 Mutex 或 Windows App SDK 的单实例机制。第二次启动时向现有进程发送激活消息，现有进程打开主窗口或显示托盘菜单。

## 4.3 托盘
托盘图标提供：
- 打开主窗口
- 闹钟列表
- 新建计时器
- Popout Clock
- 睡眠模式
- 退出

退出必须真正停止调度器和后台任务；隐藏主窗口不应停止闹钟服务。

## 4.4 调度器
使用集中式 Scheduler：
- 维护按触发时间排序的队列（PriorityQueue）。
- 只等待最近一个到期任务。
- 到期后重新计算重复规则的下一次触发。
- 系统时间变化、恢复睡眠、Profile 切换时重建队列。
- 持久化闹钟定义和触发记录；不必持久化易失的内存队列。
- 触发记录使用唯一键（AlarmId + ScheduledOccurrenceUtc）实现幂等。

## 4.5 时间源
- 日历时间：DateTimeOffset.UtcNow + TimeZoneInfo。
- 时间间隔：Stopwatch 或 TimeProvider.GetTimestamp。
- 不以 DispatcherTimer Tick 次数作为真实经过时间。
- 所有内部触发时间使用 UTC instant 存储；重复规则保存当地墙上时间与时区语义。


# 5. 数据模型与数据库

## 5.1 数据库策略
SQLite + EF Core。默认数据库放在用户 LocalAppData 应用目录。数据库文件、备份、日志和媒体引用路径分开管理。

建议路径：
%LOCALAPPDATA%/YourProduct/
  data/app.db
  backups/
  logs/
  cache/
  media-index/

## 5.2 核心表
- Profiles
- AppSettings
- Alarms
- AlarmSchedules
- AlarmActions
- AlarmOccurrences
- Timers（可选：仅保存模板/预设）
- StopwatchSessions（可选）
- UserWorldClocks
- CityCatalogMetadata（若城市目录随版本升级）
- WeatherCache
- RssFeeds
- RssArticles
- Playlists
- PlaylistItems
- Themes
- ThemeImages
- TimeTrackerSettings
- TimeTrackerSessions
- LicenseState
- UpdateState

## 5.3 关键实体字段建议
Profile:
- Id (Guid, PK)
- Name
- CreatedAtUtc
- UpdatedAtUtc
- IsDefault

Alarm:
- Id (Guid, PK)
- ProfileId (FK)
- Title
- IsEnabled
- ScheduleType
- LocalTime (TimeOnly 或分钟数)
- TimeZoneId（若闹钟采用指定时区；否则使用系统本地时区策略）
- SoundSourceType
- SoundSource
- SnoozeMinutes
- StopAfterSeconds
- FadeInSeconds
- AutoDismissSeconds
- WakeFromSleep
- CreatedAtUtc
- UpdatedAtUtc

AlarmOccurrence:
- Id
- AlarmId
- ScheduledOccurrenceUtc
- FiredAtUtc
- Result
- Unique(AlarmId, ScheduledOccurrenceUtc)

UserWorldClock:
- Id
- ProfileId
- CityId
- CustomLabel
- SortOrder
- IsFavorite
- CreatedAtUtc

City:
- Id
- DisplayName
- CountryCode
- IanaTimeZoneId
- Latitude
- Longitude
- SearchAliases
- DataVersion

## 5.4 数据库约束
- 外键开启。
- Profile 删除使用明确级联策略。
- AlarmOccurrence 唯一约束防止重复触发。
- 排序字段允许稀疏值或在重排时统一重编号。
- 所有 UTC 时间字段统一命名为 *Utc。
- 迁移必须向前兼容或提供备份回滚流程。


# 6. 时间与时区实现规范

## 6.1 世界时钟
城市数据使用 IANA ID，例如 Asia/Singapore、Europe/London、America/New_York。使用 TimeZoneConverter 将 IANA ID 转换为当前平台可识别的 TimeZoneInfo；同时验证时区存在。

示例：
```csharp
var tz = TZConvert.GetTimeZoneInfo("America/New_York");
var utcNow = DateTimeOffset.UtcNow;
var local = TimeZoneInfo.ConvertTime(utcNow, tz);
var offset = tz.GetUtcOffset(utcNow.UtcDateTime);
var isDst = tz.IsDaylightSavingTime(utcNow.UtcDateTime);
```

每轮 UI 刷新只获取一次 utcNow，并将同一快照转换到所有城市。

## 6.2 闹钟时间语义
闹钟必须明确是：
A. 系统本地时间闹钟：例如“每天本机 07:30”；
B. 指定时区闹钟：例如“纽约当地时间每天 08:00”。

不要仅保存一个 UTC 时间并据此推导所有未来重复事件。重复规则要保存本地时间、时区 ID、日期规则及 DST 处理策略。

## 6.3 DST 策略
当地时间可能无效（春季跳时）或歧义（秋季回拨）。
产品必须明确策略：
- 无效时间：跳过该次，或顺延到第一个有效时间。
- 歧义时间：选择第一次、第二次，或提示用户。
建议提供统一策略配置，并对每个边界写自动化测试。

## 6.4 系统时间变化
监听 Windows 时间变化/时区变化事件；收到事件后重新计算待触发任务和城市显示。不要继续沿用旧 UTC 偏移缓存。


# 7. 闹钟、动作与电源管理

## 7.1 Schedule 类型
- OneTime：具体日期时间。
- Daily：每天指定本地时间。
- Weekly：选定星期。
- MonthlyDay：每月指定日期；需定义 29/30/31 日缺失时策略。
- Interval：每隔 N 分钟/小时；需定义是按墙上时间还是按绝对间隔。

## 7.2 闹钟声音
SoundSourceType：
- BuiltIn
- LocalFile
- RadioStream
- Playlist

音频接口建议：
```csharp
public interface IAudioPlaybackService
{
    Task PlayAsync(AudioSource source, CancellationToken ct);
    Task StopAsync(CancellationToken ct);
    Task SetVolumeAsync(float volume, CancellationToken ct);
    Task FadeToAsync(float targetVolume, TimeSpan duration, CancellationToken ct);
}
```

## 7.3 贪睡与自动关闭
- Snooze：创建新的临时触发时间，不改变原始重复规则。
- StopAfter：停止声音但保留弹窗。
- AutoDismiss：关闭弹窗并记录结果。
- 同一 AlarmOccurrence 不得重复触发。
- 用户手动关闭、贪睡、自动关闭都要记录事件。

## 7.4 闹钟动作
动作类型：
- LaunchProcess
- OpenUrl
- HttpGet

安全约束：
- URL 只允许 http/https。
- 防止命令注入；程序路径与参数分开传递，禁止拼接到 shell 字符串。
- 对 HTTP 请求设置超时、取消、重试上限和日志。
- 默认禁止访问 loopback、link-local、私有网络等敏感地址，若产品允许高级用户配置需显式警告，防止 SSRF。
- 执行前提供测试按钮和明确权限提示。
- 不要为执行普通动作要求管理员权限。

## 7.5 睡眠唤醒
Windows waitable timer 可通过 SetWaitableTimer 请求唤醒，但实际是否唤醒取决于硬件、固件、电源策略和睡眠状态。
- 只在支持的设备上显示可用状态。
- 通过能力检测和实际测试向用户说明。
- 关机状态不保证唤醒。
- 睡眠/休眠、Modern Standby、传统 S3 行为需在真实设备测试。
- 唤醒成功后应用可能需要由系统恢复启动项/后台进程逻辑补齐。
- 不得把唤醒功能宣传为所有设备必定可用。


# 8. 天气、RSS、音乐和主题的实现

## 8.1 天气
WeatherProvider 抽象不同服务商。以经纬度查询；缓存按城市与单位管理。网络层使用 IHttpClientFactory，设置超时、User-Agent、重试与限流。温度单位只在展示层转换或在 DTO 中明确单位。

## 8.2 RSS
FeedFetcher 获取 RSS/Atom；FeedParser 解析；ArticleStore 去重并保存。去重键优先使用 FeedId + Guid/Link；无稳定 ID 时用规范化 URL + 标题 + 发布时间哈希。
安全要求：限制 XML 大小、禁用 DTD/外部实体、限制重定向次数、只允许 HTTP(S)。

## 8.3 音频
将播放器抽象为 AudioBackend，先实现本地文件，再扩展 HTTP 流。不同流媒体编码兼容性需实际验证。网络电台失败应回退到备用铃声。播放队列、音量渐变和闹钟优先级由 AudioPlaybackService 统一管理。

## 8.4 主题/照片
Theme 包含背景色、图片列表、切换间隔、前景颜色、透明度。图片加载异步完成，解码尺寸按窗口实际显示大小缩放；使用缓存，避免一次性将整个相册载入内存。用户删除原图片时显示占位并提供重新定位。

## 8.5 睡眠模式
SleepModeService 管理时长、播放列表、淡出曲线与退出状态。建议使用线性或平滑曲线渐弱；用户手动调音量、闹钟响起或退出模式时取消当前淡出任务。


# 9. 时间追踪模块

## 9.1 模式
Timer-based：以启动时间和单调时间源计算累计工作时间。
Activity-based：使用 Windows GetLastInputInfo 等系统接口判断最后输入时间；只判断空闲时长，不记录键盘内容或按键序列。

## 9.2 计时规则
- Short Break：默认建议 10 分钟工作后休息 1 分钟。
- Big Break：默认建议 50 分钟工作后休息 10 分钟。
- Stop At：达到工作结束时间停止追踪。
- Big Break 优先级高于 Short Break。
- 休息开始前可提前提醒。
- 用户可延后、结束休息或切换勿扰模式。

以上默认值为建议初始配置，可由用户修改，不应视为医疗建议。

## 9.3 系统动作风险
锁定、注销、睡眠、重启、关机属于高影响操作：
- 默认关闭自动关机类动作。
- 配置时二次确认。
- 执行前显示倒计时取消入口。
- 记录操作日志。


# 10. UI/UX 页面清单

| 页面/窗口 | 主要内容 |
|---|---|
| MainWindow | 主时钟、天气、新闻、菜单 |
| AlarmManagerWindow | 闹钟列表、新建/编辑/删除 |
| AlarmEditorWindow | 规则、铃声、动作、高级选项 |
| TimerWindow | 多计时器 |
| StopwatchWindow | 秒表与 Lap |
| PopoutClockWindow | 悬浮时钟 |
| WorldClockView | 多城市时间 |
| CityPickerWindow | 城市搜索与添加 |
| WeatherSettingsView | 地点、单位、刷新策略 |
| NewsSettingsView | 新闻类别、RSS Feed |
| PlaylistManagerWindow | 音乐/电台播放列表 |
| ThemeManagerWindow | 主题、照片相册 |
| SleepModeWindow | 睡眠时长与播放列表 |
| TimeTrackerWindow | 工作、短休息、长休息、停止时间 |
| ProfileManagerWindow | 配置文件管理 |
| SettingsWindow | 全局和 Profile 设置 |
| LicenseWindow | License Key 验证 |
| AboutWindow | 版本、更新、诊断信息 |

UI 规范：
- MVVM；Code Behind 仅处理纯 UI 事件或平台限制。
- 长任务显示加载状态且支持取消。
- 所有空状态、错误状态、加载状态均需设计。
- 高 DPI、多显示器、键盘导航、屏幕阅读器可访问性纳入验收。


# 11. 设置模型与默认值

建议将设置按 GlobalSettings 与 ProfileSettings 分离。

GlobalSettings：
- StartWithWindows
- StartMinimizedToTray
- StartupWindowState
- UpdateChannel
- LogLevel
- DataDirectory

ProfileSettings：
- TimeFormat (12h/24h)
- ShowSeconds
- ShowFullDate
- ClockStyle
- ClockPosition
- ClockSize
- ForegroundColor
- ForegroundOpacity
- BackgroundOpacity
- WeatherUnit
- WeatherEnabled
- NewsEnabled
- NewsRefreshMinutes
- ThemeId
- SleepFadeDuration
- TimeTrackerPreferences

所有设置都应有默认值、校验范围、迁移策略和恢复默认操作。


# 12. 安全、隐私与可靠性

- 使用最小权限原则，不默认要求管理员权限。
- 不采集键盘内容；Activity tracking 只读取空闲时间。
- 天气定位优先使用手动城市选择；使用系统定位时明确征求许可。
- 外部 URL、RSS、音频流都视为不可信输入。
- 限制 HTTP 响应大小、重定向次数和超时。
- 使用安全 XML 解析设置，禁用 DTD。
- 日志中避免记录 License Key、用户文件完整敏感路径和不必要的个人数据。
- License Key 本地存储应使用 Windows DPAPI 或 Credential Locker 等受保护存储。
- 数据库定期备份；升级前自动备份。
- 崩溃时保存诊断日志，但提供用户审阅和选择是否提交。
- 更新包应验证签名或可信哈希；下载和安装过程避免执行未经验证的二进制。


# 13. 授权、更新、安装与发布

## 13.1 授权
建议将授权服务与 UI 解耦：
- LicenseValidator：验证密钥、产品、期限、设备策略。
- FeatureEntitlement：根据授权状态返回功能权限。
- LicenseCache：保存签名后的授权凭据和最后成功验证时间。
- OfflineGrace：允许短期离线使用，期限由商业策略决定。

客户端不应单独决定授权有效性；服务端签发签名凭据，客户端使用公钥验证签名。不得把服务端私钥嵌入客户端。

## 13.2 更新
- 启动时检查更新（可配置）。
- 手动检查更新。
- 显示版本、变更日志、下载进度。
- 下载后校验签名/哈希。
- 更新前备份数据库。
- 更新后执行迁移。
- 安装版若由 Microsoft Store 分发，应遵循 Store 自动更新流程。

## 13.3 安装
可选：
- MSIX：适合现代 Windows 部署和 Store。
- Inno Setup：适合传统桌面安装程序和自定义流程。

安装和卸载必须区分程序文件与用户数据。卸载默认不删除用户数据库，除非用户明确选择删除数据。


# 14. 测试策略与验收标准

## 14.1 单元测试
- 重复闹钟下一次触发时间计算。
- 月末、闰年、跨年。
- DST 无效/歧义时间。
- 时区转换。
- 贪睡和自动关闭。
- Scheduler 幂等。
- 计时器与秒表时间精度。
- 天气缓存过期。
- RSS 去重。
- Profile 隔离。

## 14.2 集成测试
- SQLite Migration。
- 重启恢复闹钟。
- 音频播放与停止。
- HTTP 超时与取消。
- RSS 恶意 XML 输入。
- 授权离线缓存。
- 更新前备份与升级。

## 14.3 Windows 系统测试
- Windows 10/11 x64。
- 登录启动。
- 托盘运行。
- 锁屏与解锁。
- 睡眠/休眠恢复。
- 多显示器插拔。
- DPI 变化。
- 系统时间与时区改变。
- 真实设备唤醒能力测试。

## 14.4 发布门槛
- Release 构建无错误。
- 所有必需单元测试通过。
- 无未处理的关键异常。
- 数据迁移和回滚方案经过测试。
- 核心闹钟触发无重复、无静默丢失。
- 天气/RSS/网络故障不影响核心时钟。
- 安装、升级、卸载和用户数据保留行为经过验证。
- 不得在未实际运行测试时宣称测试通过。


# 15. 开发里程碑与优先级

## Milestone 0：工程骨架
- 建立解决方案、项目分层、DI、日志、配置。
- CI 构建与测试。
- 数据库迁移框架。
交付：可运行空壳应用。

## Milestone 1：核心时钟
- 主时钟、时间格式、设置、托盘、开机启动。
- 窗口状态记忆。
交付：可长期后台运行的桌面时钟。

## Milestone 2：闹钟引擎
- 闹钟 CRUD、重复规则、Scheduler、弹窗、音频。
- 贪睡、停止响铃、自动关闭。
交付：稳定的本地闹钟。

## Milestone 3：计时器/秒表/Popout
- 多计时器、秒表、置顶悬浮窗。
交付：时间工具集。

## Milestone 4：世界时钟与 Profile
- 城市目录、时区转换、城市管理、多配置文件。
交付：可扩展的多城市时钟。

## Milestone 5：主题/天气/RSS
- 天气服务、缓存、RSS、照片主题。
交付：信息聚合主界面。

## Milestone 6：音频/睡眠模式
- 播放列表、电台、渐弱、睡眠定时。
交付：媒体与睡眠体验。

## Milestone 7：时间追踪/系统集成
- Activity-based、休息提醒、系统动作、唤醒能力。
交付：高级生产力功能。

## Milestone 8：商业化与发布
- 授权、更新、安装包、签名、诊断。
交付：正式发布版本。

建议采用纵向切片交付：每个里程碑结束都应有可运行、可测试的软件，而不是先写完所有 UI 再补业务逻辑。


# 16. 建议接口清单

```csharp
public interface IClockService
{
    DateTimeOffset GetNow();
}

public interface IWorldClockService
{
    DateTimeOffset Convert(DateTimeOffset utcInstant, string ianaTimeZoneId);
}

public interface IAlarmService
{
    Task<Guid> CreateAsync(AlarmDraft draft, CancellationToken ct);
    Task UpdateAsync(Guid id, AlarmDraft draft, CancellationToken ct);
    Task DeleteAsync(Guid id, CancellationToken ct);
    Task SetEnabledAsync(Guid id, bool enabled, CancellationToken ct);
}

public interface IScheduler
{
    Task StartAsync(CancellationToken ct);
    Task RebuildAsync(CancellationToken ct);
    Task StopAsync(CancellationToken ct);
}

public interface IAudioPlaybackService
{
    Task PlayAsync(AudioSource source, CancellationToken ct);
    Task StopAsync(CancellationToken ct);
    Task FadeToAsync(float volume, TimeSpan duration, CancellationToken ct);
}

public interface IWeatherService
{
    Task<WeatherSnapshot?> GetCurrentAsync(
        GeoLocation location, CancellationToken ct);
}

public interface IRssService
{
    Task<IReadOnlyList<RssArticle>> RefreshAsync(
        FeedDefinition feed, CancellationToken ct);
}

public interface IProfileService
{
    Task<IReadOnlyList<Profile>> GetAllAsync(CancellationToken ct);
    Task SwitchAsync(Guid profileId, CancellationToken ct);
}
```

接口仅为建议契约。实现时应根据具体需求补充结果类型、领域异常、分页、并发控制和事务边界。


# 17. 风险清单

| 风险 | 影响 | 缓解措施 |
|---|---|---|
| DST 边界计算错误 | 闹钟提前/延迟/漏触发 | 明确墙上时间语义，边界测试 |
| 系统睡眠/唤醒不一致 | 闹钟无法准时触发 | 能力检测、设备测试、醒后补偿 |
| 音频格式不兼容 | 播放失败 | 后端抽象、格式验证、备用铃声 |
| RSS 恶意 XML | 安全风险 | 禁用 DTD、限制大小与超时 |
| 天气 API 不可用 | 信息缺失 | 缓存、退避、核心模块隔离 |
| 多 Profile 数据串用 | 用户数据错误 | ProfileId 外键、切换事务 |
| 系统时间被修改 | 调度错乱 | 监听时间变化并重建队列 |
| 更新破坏数据库 | 数据丢失 | 自动备份、迁移测试 |
| 多显示器变化 | 窗口不可见 | 启动时校正窗口边界 |
| 授权服务离线 | Pro 功能误锁 | 签名缓存与离线宽限策略 |


# 18. AI 辅助开发协作规范

如果使用多个 AI 编程工具协同开发，必须先提交并维护以下项目文档：
- docs/architecture/overview.md
- docs/architecture/module-boundaries.md
- docs/architecture/data-model.md
- docs/architecture/time-and-timezone.md
- docs/decisions/ADR-*.md
- docs/operations/build-test-release.md

协作规则：
1. 任何 AI 不得擅自改变已确认技术栈。
2. 修改公共接口前先更新接口文档和相关测试。
3. 不允许两个代理同时修改同一核心文件，除非明确分工。
4. 每个任务必须包含目标、范围、不可修改项、验收命令。
5. 完成后必须报告修改文件、设计决策、测试命令和真实结果。
6. 不得伪造构建、测试、运行或性能结果。
7. 依赖升级必须说明原因、版本和兼容性。
8. 先实现最小闭环，再扩展功能。


# 19. 参考资料与事实来源

以下资料用于确认 Olive Clock 的公开功能与用户可见行为，不代表其内部实现：

1. Olive Clock 官网  
   https://oliveclock.com/

2. 官方 Features 页面  
   https://oliveclock.com/features  
   确认闹钟、动作、悬浮时钟、计时器、天气、世界时钟、RSS、音乐、电台、主题、Profile、睡眠模式、时间追踪等公开功能。

3. 官方 Getting Started  
   https://oliveclock.com/docs  
   确认安装、Windows 登录启动、更新、数据目录与卸载说明。

4. 官方 Settings  
   https://oliveclock.com/docs/clock-settings  
   确认主时钟、悬浮时钟、启动、天气和新闻相关设置。

5. 官方 Alarms  
   https://oliveclock.com/docs/alarms  
   确认闹钟规则、铃声来源、贪睡、渐增音量、唤醒和动作配置。

6. 官方 Popout Clock Widget  
   https://oliveclock.com/docs/popout-clock-widget

7. 官方 Time Tracker  
   https://oliveclock.com/docs/time-tracker

8. 官方 v1.x 到 v2.x 授权迁移说明  
   https://oliveclock.com/announcements/migrate-to-license-keys  
   确认 v2.0 起使用 License Key，不再要求应用登录。

实施时还应查阅：
- Microsoft .NET / WPF / TimeZoneInfo 官方文档
- Windows SetWaitableTimer / Power Management 官方文档
- EF Core SQLite 官方文档
- IANA Time Zone Database
- TimeZoneConverter 项目文档
- 所选音频库、RSS 库及天气 API 官方文档


# 附录 A：首个可交付版本（MVP）

为避免项目一开始范围失控，首个版本建议只交付：
- WPF 主窗口与 MVVM
- 数字主时钟
- 12/24 小时制
- 系统托盘与登录启动
- SQLite 设置持久化
- 闹钟 CRUD 与每日/一次性/按星期重复
- 集中调度器
- 内置铃声与本地音频
- 贪睡和停止响铃
- 多计时器与秒表
- Popout Clock
- 世界时钟城市目录与时区转换
- 基础日志与自动化测试

暂缓：天气、RSS、在线电台、时间追踪、License 商业化、睡眠唤醒高级兼容、云同步。


# 附录 B：每个开发任务的 Definition of Done

任务只有同时满足以下条件才算完成：
- 代码符合项目分层和命名约定。
- 不存在未经说明的新增依赖。
- 公共接口和数据模型已更新文档。
- 关键业务逻辑有自动化测试。
- 通过格式化、构建和相关测试。
- 错误路径和取消路径经过处理。
- 日志不包含密钥或敏感信息。
- UI 有加载、空数据和错误状态。
- 数据库变更包含 Migration。
- PR/提交说明包含实际验证命令和结果。
