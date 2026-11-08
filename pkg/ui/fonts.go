// Package ui 实现基于 Viva GPU 原生框架的声明式桌面界面、帧重绘调度、主题配色与交互事件路由。
package ui

import (
	"sort"
	"strings"

	"worldclock/pkg/i18n"

	"github.com/vivaui/viva/app"
)

// CuratedFontOption 描述一款推荐的跨平台/系统内置通用字体及其多语言显示名称。
type CuratedFontOption struct {
	Family  string // 系统字体族标准名称（空字符串 "" 表示系统默认字体）
	LabelZH string // 简体中文友好显示名称
	LabelEN string // 英文友好显示名称
	LabelJA string // 日文友好显示名称
}

// curatedUniversalFonts 定义在 Windows / macOS 等主流电脑系统中广泛内置、支持高清渲染的 8 款核心精选字体（排成紧凑 2 行 × 4 列）。
// 其余所有系统已安装字体均可通过“📂 调取系统全部已安装字体”面板完整浏览与选用。
var curatedUniversalFonts = []CuratedFontOption{
	{Family: "", LabelZH: "系统默认 (自动适配)", LabelEN: "System Default", LabelJA: "システム標準"},
	{Family: "Microsoft YaHei", LabelZH: "微软雅黑 (YaHei)", LabelEN: "Microsoft YaHei", LabelJA: "雅黒 (YaHei)"},
	{Family: "SimHei", LabelZH: "黑体 (SimHei)", LabelEN: "SimHei", LabelJA: "黒体 (SimHei)"},
	{Family: "KaiTi", LabelZH: "楷体 (KaiTi)", LabelEN: "KaiTi", LabelJA: "楷書体 (KaiTi)"},
	{Family: "SimSun", LabelZH: "宋体 (SimSun)", LabelEN: "SimSun", LabelJA: "明朝 / 宋体"},
	{Family: "Segoe UI", LabelZH: "Segoe UI (现代)", LabelEN: "Segoe UI", LabelJA: "Segoe UI"},
	{Family: "Bahnschrift", LabelZH: "Bahnschrift (数显)", LabelEN: "Bahnschrift", LabelJA: "Bahnschrift"},
	{Family: "Consolas", LabelZH: "Consolas (等宽)", LabelEN: "Consolas", LabelJA: "Consolas"},
}

// fallbackSystemFonts 当在无头测试环境或不支持枚举字体的平台上 app.FontFamilies() 返回空切片时，
// 提供一套安全的通用基准字体列表，确保界面选择器与单元测试在任何环境下均一致可用。
var fallbackSystemFonts = []string{
	"Arial",
	"Bahnschrift",
	"Calibri",
	"Cambria",
	"Consolas",
	"Courier New",
	"FangSong",
	"Georgia",
	"KaiTi",
	"Meiryo",
	"Microsoft YaHei",
	"Microsoft YaHei UI",
	"Segoe UI",
	"SimHei",
	"SimSun",
	"Tahoma",
	"Times New Roman",
	"Trebuchet MS",
	"Verdana",
	"Yu Gothic UI",
}

// symbolFontBlacklist 记录操作系统中将普通字母/数字映射为图标符号的专用符号字体（小写匹配）。
// 过滤这些字体可从根源杜绝用户误选后导致文字和数字变成符号乱码的问题。
var symbolFontBlacklist = map[string]struct{}{
	"wingdings":              {},
	"wingdings 2":            {},
	"wingdings 3":            {},
	"webdings":               {},
	"marlett":                {},
	"symbol":                 {},
	"mt extra":               {},
	"bookshelf symbol 7":     {},
	"ms reference specialty": {},
	"ms outlook":             {},
	"segoe mdl2 assets":      {},
	"segoe fluent icons":     {},
	"hololens mdl2 assets":   {},
	"apple symbols":          {},
	"zapf dingbats":          {},
}

// IsReadableSystemFont 判断给定的系统字体族名称是否适合作为应用全局可读文字与数字字体。
// 自动过滤竖排字体 (@ 开头)、控制字符、超长截断名以及纯符号/图标字体。
func IsReadableSystemFont(name string) bool {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len(trimmed) >= 31 || trimmed[0] == '@' {
		return false
	}
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] < 0x20 {
			return false
		}
	}
	lower := strings.ToLower(trimmed)
	if _, blocked := symbolFontBlacklist[lower]; blocked {
		return false
	}
	if strings.Contains(lower, "wingdings") ||
		strings.Contains(lower, "webdings") ||
		strings.Contains(lower, "mdl2 assets") ||
		strings.Contains(lower, "fluent icons") ||
		strings.Contains(lower, "dingbats") {
		return false
	}
	return true
}

// FilterReadableSystemFonts 对系统原始枚举出的字体族列表进行清洗、可读性过滤、去重与字典序排序。
func FilterReadableSystemFonts(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, name := range raw {
		trimmed := strings.TrimSpace(name)
		if !IsReadableSystemFont(trimmed) {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		out = append(out, fallbackSystemFonts...)
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i]) < strings.ToLower(out[j])
	})
	return out
}

// LoadInstalledSystemFonts 调用 Viva 原生 app.FontFamilies() 读取当前电脑已安装的字体库并完成安全过滤。
func LoadInstalledSystemFonts() []string {
	if isTestBinary() {
		return append([]string(nil), fallbackSystemFonts...)
	}
	return FilterReadableSystemFonts(app.FontFamilies())
}

// ResolveSafeFontFamily 校验用户请求或配置文件保存的字体名称 requested 是否在当前电脑已安装字体列表 installed 中可用。
// 若目标字体未在当前电脑安装（例如跨电脑迁移了配置文件）或属于不可读符号字体，则自动安全回退为 ""（系统默认字体），
// 配合 DirectWrite (IDWriteFactory2::GetSystemFontFallback) 硬件级全字符集回退链，保证在任何电脑上均能 100% 正常显示。
func ResolveSafeFontFamily(requested string, installed []string) string {
	trimmed := strings.TrimSpace(requested)
	if trimmed == "" || !IsReadableSystemFont(trimmed) {
		return ""
	}
	for _, f := range installed {
		if strings.EqualFold(f, trimmed) {
			return f
		}
	}
	return ""
}

// AvailableCuratedFonts 返回当前电脑已安装的精选推荐字体列表（第一项始终为“系统默认”）。
func AvailableCuratedFonts(installed []string) []CuratedFontOption {
	installedMap := make(map[string]string, len(installed))
	for _, f := range installed {
		installedMap[strings.ToLower(f)] = f
	}
	result := make([]CuratedFontOption, 0, len(curatedUniversalFonts))
	for _, opt := range curatedUniversalFonts {
		if opt.Family == "" {
			result = append(result, opt)
			continue
		}
		if realName, ok := installedMap[strings.ToLower(opt.Family)]; ok {
			item := opt
			item.Family = realName
			result = append(result, item)
		}
	}
	return result
}

// CuratedFontLabel 根据当前界面语言返回精选字体选项的本地化显示名称。
func CuratedFontLabel(opt CuratedFontOption, lang i18n.Language) string {
	switch lang {
	case i18n.LangZH:
		return opt.LabelZH
	case i18n.LangJA:
		return opt.LabelJA
	default:
		return opt.LabelEN
	}
}

// FormatFontDisplayName 返回当前生效字体族在界面上的友好显示名称。
func FormatFontDisplayName(family string, lang i18n.Language) string {
	trimmed := strings.TrimSpace(family)
	for _, opt := range curatedUniversalFonts {
		if strings.EqualFold(opt.Family, trimmed) {
			return CuratedFontLabel(opt, lang)
		}
	}
	return trimmed
}

// ApplyProcessFontFamily 将经过安全校验的字体族名称应用到 Viva 全局文本渲染引擎。
func ApplyProcessFontFamily(family string) {
	if isTestBinary() {
		return
	}
	_ = app.SetInterfaceFontFamily(family)
}
