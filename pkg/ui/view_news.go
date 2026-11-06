package ui

import (
	"fmt"
	"time"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawNews renders the news and RSS feed aggregator screen.
func DrawNews(f *flex.Frame, state *AppState, theme ui.Theme) {
	category := state.NewsCategory
	if category == "" {
		category = "All"
	}

	articles := state.News.Articles(category)

	f.Column().Fill().Pad(24).Gap(16).Children(func() {
		// Header & Category Bar
		f.Row().FillWidth().Items(flex.Center).Gap(12).Children(func() {
			f.Text(state.T("news.title")).FontSize(22).Bold().Color(theme.Foreground)
			f.Spacer()

			// Categories
			btnAll := f.Button(IDNewsCatAll, state.T("news.all"))
			if category == "All" {
				btnAll.Outline()
			} else {
				btnAll.Ghost()
			}
			if btnAll.Clicked() {
				state.NewsCategory = "All"
			}

			btnTech := f.Button(IDNewsCatTech, state.T("news.tech"))
			if category == "Tech" {
				btnTech.Outline()
			} else {
				btnTech.Ghost()
			}
			if btnTech.Clicked() {
				state.NewsCategory = "Tech"
			}

			btnWorld := f.Button(IDNewsCatWorld, state.T("news.world"))
			if category == "World" {
				btnWorld.Outline()
			} else {
				btnWorld.Ghost()
			}
			if btnWorld.Clicked() {
				state.NewsCategory = "World"
			}

			btnSci := f.Button(IDNewsCatScience, state.T("news.sci"))
			if category == "Science" {
				btnSci.Outline()
			} else {
				btnSci.Ghost()
			}
			if btnSci.Clicked() {
				state.NewsCategory = "Science"
			}
		})

		// Article Feed List
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

					titleStr, sourceStr, summaryStr := localizeArticle(state, a.ID, a.Title, a.Source, a.Summary)

					f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Gap(8).Children(func() {
						// Category chip & Source & Time
						f.Row().Items(flex.Center).Gap(8).Children(func() {
							f.Text(fmt.Sprintf("[%s]", catLabel)).FontSize(12).Bold().Color(theme.Primary)
							f.Text(sourceStr).FontSize(12).Bold().Color(theme.Foreground)
							f.Text("•").FontSize(12).Muted()
							elapsed := time.Since(a.PublishedAt)
							var timeAgoStr string
							if elapsed >= time.Hour {
								hrs := int(elapsed.Hours())
								if hrs < 1 {
									hrs = 1
								}
								timeAgoStr = fmt.Sprintf(state.T("news.hrs_ago"), hrs)
							} else {
								mins := int(elapsed.Minutes())
								if mins < 1 {
									mins = 1
								}
								timeAgoStr = fmt.Sprintf(state.T("news.mins_ago"), mins)
							}
							f.Text(timeAgoStr).FontSize(12).Muted()
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

func localizeArticle(state *AppState, id, fallbackTitle, fallbackSource, fallbackSummary string) (string, string, string) {
	lang := string(state.Lang())
	switch id {
	case "tech-1":
		switch lang {
		case "zh":
			return "新一代高性能 GPU 原生桌面框架演进", "科技视界", "解析现代 Go 桌面架构如何实现零内存分配绘制与瞬时 60fps 丝滑响应。"
		case "ja":
			return "次世代の高性能GPUネイティブ・デスクトップフレームワーク", "テック・ホライズン", "最新のGoデスクトップ設計がいかにしてゼロアロケーション描画と60fpsの高速応答を実現するか。"
		}
	case "world-1":
		switch lang {
		case "zh":
			return "2026 全球时区同步标准正式更新", "世界时讯", "国际标准局发布最新 IANA 时区数据库更新，涵盖全球夏令时调整规则。"
		case "ja":
			return "2026年世界タイムゾーン同期標準がアップデート", "ワールド・クロノ", "国際機関が最新のIANAタイムゾーン更新を発表し、サマータイム移行ルールを刷新。"
		}
	case "sci-1":
		switch lang {
		case "zh":
			return "深空望远镜在遥远星系发现新型脉冲星计时源", "科学日报", "天文学家证实新发现的宇宙脉冲星信号具有天然原子钟级别的超高规律性。"
		case "ja":
			return "深宇宙望遠鏡が遠方銀河に新たなパルサー時計を発見", "サイエンス・デイリー", "天文学チームが新発見の宇宙パルサー信号に原子時計並みの規則性を確認。"
		}
	case "tech-2":
		switch lang {
		case "zh":
			return "桌面应用能效革新：零轮询空闲架构解析", "开发者洞察", "实测桌面应用在用户事件之间完全休眠所带来的显著续航提升与硬件减负。"
		case "ja":
			return "デスクトップアプリの省電力化：ゼロポーリング設計", "デブ・インサイト", "イベント間で完全スリープするデスクトップアーキテクチャの省エネ性能を検証。"
		}
	}
	return fallbackTitle, fallbackSource, fallbackSummary
}
