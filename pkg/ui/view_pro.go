package ui

import (
	"fmt"

	"worldclock/pkg/license"

	"github.com/vivaui/viva/flex"
	"github.com/vivaui/viva/ui"
)

// DrawProModal 在顶部导航栏点击“👑 Pro 会员 / 试用 X 天”按钮时渲染全屏/模态订阅与试用中心视图。
// DrawProModal renders the Pro Subscription & 7-Day Free Trial modal view.
func DrawProModal(f *flex.Frame, state *AppState, theme ui.Theme) {
	f.Column().Fill().Pad(20).Gap(14).Children(func() {
		// 顶部标题栏 + 关闭模态窗按钮
		f.Row().FillWidth().Items(flex.Center).Gap(12).Children(func() {
			f.Text(state.T("pro.title")).FontSize(20).Bold().Color(theme.Primary)
			f.Spacer()
			if f.Button(IDBtnCloseProModal, state.T("pro.close")).Outline().Small().Clicked() {
				state.ShowProModal = false
				state.License.ClearNotice()
			}
		})

		// 可滚动内容区
		scrollBox := f.ScrollArea(IDScrollSettings).Fill()
		scrollBox.Children(func() {
			DrawProCenter(f, state, theme)
		})
		ss := state.View.ScrollState(IDScrollSettings)
		DrawScrollbar(f.Canvas(), scrollBox, ss, theme)
	})
}

// DrawTrialExpiredLockout 渲染试用期 7 天结束时的全屏强制锁定屏障（Lockout Paywall）。
// 严格阻断任何未经订阅授权的时钟、世界时钟、闹钟、倒计时、番茄钟、白噪音与贴边条/悬浮窗访问。
func DrawTrialExpiredLockout(f *flex.Frame, state *AppState, theme ui.Theme) {
	scrollBox := f.ScrollArea(IDScrollLockout).Fill().Pad(24).Gap(16)
	scrollBox.Children(func() {
		// 1. 顶部醒目的危险/锁定警告卡片
		f.Column().FillWidth().Pad(20).Background(theme.Card).Radius(12).Border(2, theme.Danger).Gap(10).Children(func() {
			f.Row().Items(flex.Center).Gap(10).Children(func() {
				f.Text(state.T("pro.lockout_title")).FontSize(20).Bold().Color(theme.Danger)
			})
			f.Text(state.T("pro.lockout_desc")).FontSize(13).Color(theme.Foreground)

			// 操作反馈即时提示条
			if noticeCode := state.License.LastNotice(); noticeCode != "" {
				var noticeI18nKey string
				switch noticeCode {
				case "activated_monthly":
					noticeI18nKey = "pro.notice_monthly"
				case "activated_yearly":
					noticeI18nKey = "pro.notice_yearly"
				case "restored":
					noticeI18nKey = "pro.notice_restored"
				case "trial_reset":
					noticeI18nKey = "pro.notice_trial"
				case "full_test":
					noticeI18nKey = "pro.notice_full"
				}
				if noticeI18nKey != "" {
					f.Row().FillWidth().Pad(10).Background(theme.Background).Radius(8).Border(1, theme.Accent).Items(flex.Center).Children(func() {
						f.Text(state.T(noticeI18nKey)).FontSize(13).Bold().Color(theme.Accent)
					})
				}
			}
		})

		// 2. 核心订阅方案对比卡片区（按月订阅 ¥3/月 vs 按年订阅 ¥28/年）
		f.Row().FillWidth().Gap(14).Children(func() {
			// 方案 A：按月订阅（¥3 / 月）
			f.Column().Grow(1).Pad(18).Background(theme.Card).Radius(12).Border(1, theme.Border).Gap(10).Children(func() {
				f.Text(state.T("pro.plan_monthly_title")).FontSize(16).Bold().Color(theme.Foreground)
				f.Text(state.T("pro.plan_monthly_price")).FontSize(26).Bold().Color(theme.Primary)
				f.Text(state.T("pro.plan_monthly_desc")).FontSize(12).Muted()
				f.Spacer()
				if f.Button(IDBtnProSubMonthly, state.T("pro.plan_monthly_btn")).Clicked() {
					state.License.ActivateMonthly(state.Now)
					state.Persist()
				}
			})

			// 方案 B：按年订阅（¥28 / 年 · 立省 22%）
			f.Column().Grow(1).Pad(18).Background(theme.Card).Radius(12).Border(1, theme.Primary).Gap(10).Children(func() {
				f.Text(state.T("pro.plan_yearly_title")).FontSize(16).Bold().Color(theme.Primary)
				f.Text(state.T("pro.plan_yearly_price")).FontSize(26).Bold().Color(theme.Accent)
				f.Text(state.T("pro.plan_yearly_desc")).FontSize(12).Muted()
				f.Spacer()
				if f.Button(IDBtnProSubYearly, state.T("pro.plan_yearly_btn")).Clicked() {
					state.License.ActivateYearly(state.Now)
					state.Persist()
				}
			})
		})

		// 3. Pro 会员核心权益说明（展示订阅后即可解锁的所有能力）
		f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Gap(8).Children(func() {
			f.Text(state.T("pro.benefits_title")).FontSize(15).Bold().Color(theme.Primary)
			f.Text(state.T("pro.benefit_1")).FontSize(13).Color(theme.Foreground)
			f.Text(state.T("pro.benefit_2")).FontSize(13).Color(theme.Foreground)
			f.Text(state.T("pro.benefit_3")).FontSize(13).Color(theme.Foreground)
			f.Text(state.T("pro.benefit_4")).FontSize(13).Color(theme.Foreground)
		})

		// 4. 底部出路操作栏：恢复购买 / 管理订阅 / 退出程序 / 测试通道
		f.Row().FillWidth().Items(flex.Center).Gap(10).Children(func() {
			if f.Button(IDBtnProRestore, state.T("pro.restore_btn")).Small().Outline().Clicked() {
				state.License.RestorePurchases(state.Now)
				state.Persist()
			}
			if f.Button(IDBtnProManage, state.T("pro.manage_btn")).Small().Secondary().Clicked() {
				state.License.RestorePurchases(state.Now)
				state.Persist()
			}
			if f.Button(IDBtnLockoutExit, state.T("pro.lockout_exit")).Small().Secondary().Clicked() {
				state.ForceQuit = true
				if state.Window != nil {
					state.Window.Close()
				}
				if state.Application != nil {
					state.Application.Quit()
				}
			}

			f.Spacer()

			// 调试测试环境免付费通道（方便本地开发、测试与审核）
			if f.Button(IDBtnProSimFullTest, state.T("pro.test_sim_full")).Small().Ghost().Clicked() {
				state.License.ActivateFullTest()
				state.Persist()
			}
			if f.Button(IDBtnProSimTrial, state.T("pro.test_sim_trial")).Small().Ghost().Clicked() {
				state.License.ResetToTrial(state.Now)
				state.Persist()
			}
		})
	})
	ss := state.View.ScrollState(IDScrollLockout)
	DrawScrollbar(f.Canvas(), scrollBox, ss, theme)
}

// DrawProCenter 渲染 Pro 会员订阅、7 天免费试用倒计时、恢复购买与完整免付测试版切换中心。
// DrawProCenter renders the Pro subscription plans (¥3/month, ¥28/year), 7-day trial status, and free full-test mode switcher.
func DrawProCenter(f *flex.Frame, state *AppState, theme ui.Theme) {
	plan := state.License.Plan()
	daysLeft := state.License.TrialDaysRemaining(state.Now)

	f.Column().FillWidth().Gap(16).Children(func() {
		// 卡片 1：当前授权状态横幅与免付费完整测试说明
		f.Column().FillWidth().Pad(18).Background(theme.Card).Radius(12).Border(1, theme.Border).Gap(10).Children(func() {
			f.Text(state.T("pro.subtitle")).FontSize(13).Muted()

			// 当前状态高亮显示
			var statusText string
			switch plan {
			case license.PlanMonthly:
				statusText = state.T("pro.status_monthly")
			case license.PlanYearly:
				statusText = state.T("pro.status_yearly")
			case license.PlanFullTest:
				statusText = state.T("pro.status_full_test")
			default:
				statusText = fmt.Sprintf(state.T("pro.status_trial"), daysLeft)
			}

			f.Row().FillWidth().Pad(12).Background(theme.Background).Radius(8).Border(1, theme.Primary).Items(flex.Center).Children(func() {
				f.Text(statusText).FontSize(15).Bold().Color(theme.Primary)
			})

			// 操作结果即时反馈提示条（例如：已免费激活按月/按年订阅、已恢复购买等）
			if noticeCode := state.License.LastNotice(); noticeCode != "" {
				var noticeI18nKey string
				switch noticeCode {
				case "activated_monthly":
					noticeI18nKey = "pro.notice_monthly"
				case "activated_yearly":
					noticeI18nKey = "pro.notice_yearly"
				case "restored":
					noticeI18nKey = "pro.notice_restored"
				case "trial_reset":
					noticeI18nKey = "pro.notice_trial"
				case "full_test":
					noticeI18nKey = "pro.notice_full"
				}
				if noticeI18nKey != "" {
					f.Row().FillWidth().Pad(10).Background(theme.Background).Radius(8).Border(1, theme.Accent).Items(flex.Center).Children(func() {
						f.Text(state.T(noticeI18nKey)).FontSize(13).Bold().Color(theme.Accent)
					})
				}
			}
		})

		// 卡片 2：订阅方案对比卡片（7天免费试用 + 按月 ¥3/月 vs 按年 ¥28/年）
		f.Row().FillWidth().Gap(14).Children(func() {
			// 方案 A：按月连续订阅（¥3 / 月）
			monthlyBorder := theme.Border
			if plan == license.PlanMonthly {
				monthlyBorder = theme.Primary
			}
			f.Column().Grow(1).Pad(18).Background(theme.Card).Radius(12).Border(1, monthlyBorder).Gap(10).Children(func() {
				f.Text(state.T("pro.plan_monthly_title")).FontSize(16).Bold().Color(theme.Foreground)
				f.Text(state.T("pro.plan_monthly_price")).FontSize(26).Bold().Color(theme.Primary)
				f.Text(state.T("pro.plan_monthly_desc")).FontSize(12).Muted()

				f.Spacer()

				btnMonthlyLbl := state.T("pro.plan_monthly_btn")
				btnMonthly := f.Button(IDBtnProSubMonthly, btnMonthlyLbl)
				if plan == license.PlanMonthly {
					btnMonthly.Label(state.T("pro.plan_monthly_active")).Outline()
				}
				if btnMonthly.Clicked() {
					state.License.ActivateMonthly(state.Now)
					state.Persist()
				}
			})

			// 方案 B：按年连续订阅（¥28 / 年 · 立省 22%）
			yearlyBorder := theme.Border
			if plan == license.PlanYearly {
				yearlyBorder = theme.Primary
			}
			f.Column().Grow(1).Pad(18).Background(theme.Card).Radius(12).Border(1, yearlyBorder).Gap(10).Children(func() {
				f.Text(state.T("pro.plan_yearly_title")).FontSize(16).Bold().Color(theme.Primary)
				f.Text(state.T("pro.plan_yearly_price")).FontSize(26).Bold().Color(theme.Accent)
				f.Text(state.T("pro.plan_yearly_desc")).FontSize(12).Muted()

				f.Spacer()

				btnYearlyLbl := state.T("pro.plan_yearly_btn")
				btnYearly := f.Button(IDBtnProSubYearly, btnYearlyLbl)
				if plan == license.PlanYearly {
					btnYearly.Label(state.T("pro.plan_yearly_active")).Outline()
				}
				if btnYearly.Clicked() {
					state.License.ActivateYearly(state.Now)
					state.Persist()
				}
			})
		})

		// 卡片 3：完整免付费测试控制台 & Microsoft Store 商店规范入口（恢复购买 / 管理订阅）
		f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Gap(12).Children(func() {
			f.Text(state.T("pro.test_panel_title")).FontSize(15).Bold().Color(theme.Primary)

			// 第一排：测试状态切换按钮（7天试用状态 / 永久免付完整测试版）
			f.Row().Items(flex.Center).Gap(10).Children(func() {
				btnSimTrial := f.Button(IDBtnProSimTrial, state.T("pro.test_sim_trial")).Small()
				if plan == license.PlanTrial {
					btnSimTrial.Outline()
				} else {
					btnSimTrial.Secondary()
				}
				if btnSimTrial.Clicked() {
					state.License.ResetToTrial(state.Now)
					state.Persist()
				}

				btnSimFull := f.Button(IDBtnProSimFullTest, state.T("pro.test_sim_full")).Small()
				if plan == license.PlanFullTest {
					btnSimFull.Outline()
				} else {
					btnSimFull.Secondary()
				}
				if btnSimFull.Clicked() {
					state.License.ActivateFullTest()
					state.Persist()
				}
			})

			f.Divider()

			// 第二排：微软商店上架必备按钮（🔄 恢复购买 / ⚙️ 管理或取消订阅）
			f.Row().Items(flex.Center).Gap(10).Children(func() {
				if f.Button(IDBtnProRestore, state.T("pro.restore_btn")).Small().Outline().Clicked() {
					state.License.RestorePurchases(state.Now)
					state.Persist()
				}

				if f.Button(IDBtnProManage, state.T("pro.manage_btn")).Small().Secondary().Clicked() {
					state.License.RestorePurchases(state.Now)
					state.Persist()
				}
			})
		})

		// 卡片 4：Pro 会员全部已解锁核心权益一览
		f.Column().FillWidth().Pad(16).Background(theme.Card).Radius(12).Border(1, theme.Border).Gap(8).Children(func() {
			f.Text(state.T("pro.benefits_title")).FontSize(15).Bold().Color(theme.Primary)
			f.Text(state.T("pro.benefit_1")).FontSize(13).Color(theme.Foreground)
			f.Text(state.T("pro.benefit_2")).FontSize(13).Color(theme.Foreground)
			f.Text(state.T("pro.benefit_3")).FontSize(13).Color(theme.Foreground)
			f.Text(state.T("pro.benefit_4")).FontSize(13).Color(theme.Foreground)
		})
	})
}
