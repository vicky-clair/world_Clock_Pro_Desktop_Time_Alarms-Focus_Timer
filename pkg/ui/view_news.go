package ui

import (
	"fmt"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawNews 渲染“历史上的今天”（On This Day in History）视图，
// 支持百度百科 / Wikipedia API 实时同步、按类别（全部/历史大事/科技探索/人文纪事）过滤以及中英日三语显示。
func DrawNews(f *flex.Frame, state *AppState, theme ui.Theme) {
	category := state.NewsCategory
	if category == "" {
		category = "All"
	}

	now := state.Now.Local()
	// 进入页面时确保后台触发一次今日历史 API 拉取
	if state.News != nil {
		state.News.EnsureFetchedAsync(now, string(state.Lang()), false, state.Invalidate)
	}

	articles := state.News.Articles(category)
	apiStatus, apiSource := state.News.Status()

	var dateBadge string
	switch state.Lang() {
	case "zh", "ja":
		dateBadge = fmt.Sprintf("📅 %02d月%02d日", int(now.Month()), now.Day())
	default:
		dateBadge = fmt.Sprintf("📅 %02d/%02d", int(now.Month()), now.Day())
	}

	var statusLabel string
	switch apiStatus {
	case "online":
		statusLabel = fmt.Sprintf(state.T("news.status_online"), apiSource)
	case "fetching":
		statusLabel = state.T("news.status_fetching")
	default:
		statusLabel = state.T("news.status_curated")
	}

	f.Column().Fill().Pad(24).Gap(14).Children(func() {
		// Row 1: Title, Current Date Badge, API Status & Refresh Button
		f.Row().FillWidth().Items(flex.Center).Gap(10).Children(func() {
			f.Text(state.T("news.title")).FontSize(22).Bold().Color(theme.Foreground)
			f.Row().PadX(10).PadY(4).Background(theme.Muted).Radius(8).Items(flex.Center).Children(func() {
				f.Text(dateBadge).FontSize(13).Bold().Color(theme.Primary)
			})
			f.Text(statusLabel).FontSize(12).Muted()
			f.Spacer()
			if f.Button(IDBtnRefreshHistory, state.T("news.refresh")).Small().Outline().Clicked() {
				state.News.EnsureFetchedAsync(now, string(state.Lang()), true, state.Invalidate)
			}
		})

		// Row 2: Category Filter Tabs
		f.Row().FillWidth().Items(flex.Center).Gap(8).Children(func() {
			btnAll := f.Button(IDNewsCatAll, state.T("news.all")).Small()
			if category == "All" {
				btnAll.Outline()
			} else {
				btnAll.Ghost()
			}
			if btnAll.Clicked() {
				state.NewsCategory = "All"
			}

			btnWorld := f.Button(IDNewsCatWorld, state.T("news.world")).Small()
			if category == "World" {
				btnWorld.Outline()
			} else {
				btnWorld.Ghost()
			}
			if btnWorld.Clicked() {
				state.NewsCategory = "World"
			}

			btnTech := f.Button(IDNewsCatTech, state.T("news.tech")).Small()
			if category == "Tech" {
				btnTech.Outline()
			} else {
				btnTech.Ghost()
			}
			if btnTech.Clicked() {
				state.NewsCategory = "Tech"
			}

			btnSci := f.Button(IDNewsCatScience, state.T("news.sci")).Small()
			if category == "Science" {
				btnSci.Outline()
			} else {
				btnSci.Ghost()
			}
			if btnSci.Clicked() {
				state.NewsCategory = "Science"
			}
		})

		// Historical Events Scrollable List
		scrollNews := f.ScrollArea(IDScrollNews).Fill()
		scrollNews.Children(func() {
			f.Column().FillWidth().Gap(12).Children(func() {
				for _, art := range articles {
					a := art
					catLabel := a.Category
					switch a.Category {
					case "Tech":
						catLabel = state.T("news.tech")
					case "World":
						catLabel = state.T("news.world")
					case "Science":
						catLabel = state.T("news.sci")
					}

					titleStr, sourceStr, summaryStr := localizeHistoryArticle(state, a.ID, a.Title, a.Source, a.Summary)
					yearBadge := a.Year
					if yearBadge == "" {
						yearBadge = "1900"
					}

					f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Gap(8).Children(func() {
						// Year chip, Category chip, Source & Date
						f.Row().Items(flex.Center).Gap(8).Children(func() {
							f.Row().PadX(8).PadY(2).Background(theme.Muted).Radius(6).Items(flex.Center).Children(func() {
								f.Text(fmt.Sprintf(state.T("news.year_fmt"), yearBadge)).FontSize(12).Bold().Color(theme.Primary)
							})
							f.Text(fmt.Sprintf("[%s]", catLabel)).FontSize(12).Bold().Color(theme.Accent)
							f.Text(sourceStr).FontSize(12).Bold().Color(theme.Foreground)
							f.Text("•").FontSize(12).Muted()
							f.Text(dateBadge).FontSize(12).Muted()
						})

						// Title
						f.Text(titleStr).FontSize(16).Bold().Color(theme.Foreground)

						// Summary
						f.Text(summaryStr).FontSize(13).Color(theme.MutedForeground)
					})
				}
			})
		})
		DrawScrollbar(f.Canvas(), scrollNews, state.View.ScrollState(IDScrollNews), theme)
	})
}

func localizeHistoryArticle(state *AppState, id, fallbackTitle, fallbackSource, fallbackSummary string) (string, string, string) {
	lang := string(state.Lang())
	switch id {
	case "hist-world-1":
		switch lang {
		case "zh":
			return "丹麦天文学家奥勒·罗默首次通过木卫一蚀测算出光速有限", "历史科学档案", "奥勒·罗默通过长达数年的木卫一掩食观测，在人类历史上首次证明光具有有限传播速度，奠定了现代天体物理与精密授时的基石。"
		case "ja":
			return "デンマークの天文学者オーレ・レーマーが光速の有限性を初めて実証", "歴史科学アーカイブ", "木星の衛星イオの食の観測から光が有限の速度で伝わることを証明し、近代物理学と精密計時学の基礎を築きました。"
		}
	case "hist-tech-1":
		switch lang {
		case "zh":
			return "国际计量大会正式采用铯-133原子跃迁周期定义国际单位制“秒”", "国际计量与科技史", "第13届国际计量大会决定以铯-133原子基态两个超精细能级之间跃迁辐射的 9,192,631,770 个周期定义 1 秒，人类正式步入原子钟纪元。"
		case "ja":
			return "国際度量衡総会がセシウム133原子時計による「秒」の定義を採択", "科学技術アーカイブ", "第13回国際度量衡総会にてセシウム133原子の放射周期に基づく秒の定義が採択され、人類は原子時計の時代へと入りました。"
		}
	case "hist-sci-1":
		switch lang {
		case "zh":
			return "国际子午线会议确立格林尼治本初子午线与全球 24 时区体系", "世界文明遗产", "来自25个国家的代表正式选定通过伦敦格林尼治天文台的经线作为全球零度经线，并确立全球24个标准时区制度。"
		case "ja":
			return "国際子午線会議がグリニッジ本初子午線と世界24タイムゾーンを制定", "世界文化遺産", "25カ国の代表が集まり、グリニッジ子午線を本初子午線として採択し、現代の世界標準時と24時間帯の基礎を確立しました。"
		}
	case "hist-tech-2":
		switch lang {
		case "zh":
			return "欧洲核子研究中心 (CERN) 宣布万维网 (WWW) 协议与源码向全人类免费开放", "互联网里程碑", "蒂姆·伯纳斯-李与 CERN 签署声明，将万维网核心协议与浏览器代码无偿捐献给公共领域，彻底点燃全球互联网浪潮。"
		case "ja":
			return "CERNがWorld Wide Web (WWW) のプロトコルとソースコードを無償公開", "インターネット史", "ティム・バーナーズ＝リーとCERNがWeb技術をパブリックドメインとして開放し、世界のインターネット普及を加速させました。"
		}
	case "hist-world-2":
		switch lang {
		case "zh":
			return "人类首批常驻宇航员乘组成功对接国际空间站（远征1号任务）", "航天与世界历史", "联盟 TM-31 载人飞船与国际空间站顺利对接，开启了人类在近地轨道不间断常驻科学实验的新纪元。"
		case "ja":
			return "国際宇宙ステーション (ISS) に初の長期滞在クルーがドッキング（第1次長期滞在）", "宇宙・世界史", "ソユーズTM-31が国際宇宙ステーションにドッキングし、地球低軌道における人類の継続的な滞在が始まりました。"
		}
	}
	return fallbackTitle, fallbackSource, fallbackSummary
}
