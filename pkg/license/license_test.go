package license

import (
	"testing"
	"time"
)

func TestLicenseTrialAndSubscriptions(t *testing.T) {
	now := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	mgr := NewManager(Config{}, now)

	// 1. 初始默认开启 7 天免费试用，且完整测试版全功能解锁
	if mgr.Plan() != PlanTrial {
		t.Fatalf("expected initial plan %q, got %q", PlanTrial, mgr.Plan())
	}
	if got := mgr.TrialDaysRemaining(now); got != 7 {
		t.Fatalf("expected 7 trial days remaining at start, got %d", got)
	}
	if !mgr.IsFeatureUnlocked(now) {
		t.Fatal("expected IsFeatureUnlocked to be true in full testable build")
	}

	// 2. 经过 3 天后剩余 4 天试用期
	after3Days := now.Add(3 * 24 * time.Hour)
	if got := mgr.TrialDaysRemaining(after3Days); got != 4 {
		t.Fatalf("expected 4 trial days remaining after 3 days, got %d", got)
	}

	// 3. 激活按月订阅 (¥3/月)
	mgr.ActivateMonthly(now)
	if mgr.Plan() != PlanMonthly || !mgr.IsProActive() {
		t.Fatalf("expected monthly pro active, got plan=%s", mgr.Plan())
	}
	if mgr.LastNotice() != "activated_monthly" {
		t.Fatalf("expected notice activated_monthly, got %s", mgr.LastNotice())
	}

	// 4. 激活按年订阅 (¥28/年)
	mgr.ActivateYearly(now)
	if mgr.Plan() != PlanYearly || !mgr.IsProActive() {
		t.Fatalf("expected yearly pro active, got plan=%s", mgr.Plan())
	}
	if mgr.LastNotice() != "activated_yearly" {
		t.Fatalf("expected notice activated_yearly, got %s", mgr.LastNotice())
	}
	if PriceYearlyCNY != 28 {
		t.Fatalf("expected PriceYearlyCNY=28, got %d", PriceYearlyCNY)
	}

	// 5. 切换为完整测试版模式与恢复购买
	mgr.ActivateFullTest()
	if mgr.Plan() != PlanFullTest || !mgr.IsProActive() {
		t.Fatalf("expected full_test active, got plan=%s", mgr.Plan())
	}
	mgr.RestorePurchases(now)
	if mgr.LastNotice() != "restored" {
		t.Fatalf("expected notice restored, got %s", mgr.LastNotice())
	}

	// 6. 重置回 7 天免费试用态
	mgr.ResetToTrial(now)
	if mgr.Plan() != PlanTrial || mgr.TrialDaysRemaining(now) != 7 {
		t.Fatalf("expected reset to 7-day trial, got plan=%s days=%d", mgr.Plan(), mgr.TrialDaysRemaining(now))
	}

	// 7. 试用期过期后正确报告功能锁定
	after8Days := now.Add(8 * 24 * time.Hour)
	if mgr.TrialDaysRemaining(after8Days) != 0 {
		t.Fatalf("expected 0 trial days after 8 days, got %d", mgr.TrialDaysRemaining(after8Days))
	}
	if mgr.IsFeatureUnlocked(after8Days) {
		t.Fatal("expected IsFeatureUnlocked to be false after trial expires without subscription")
	}
}
