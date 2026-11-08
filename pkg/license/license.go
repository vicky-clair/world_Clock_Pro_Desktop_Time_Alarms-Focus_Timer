// Package license 实现 小瑞综合时钟桌面 (Xiaoru Integrated Clock Desktop) 的 7 天免费试用期计算、月度(¥3/月)/年度(¥28/年)订阅状态管理，
// 以及 Microsoft Store (Windows.Services.Store) 内购商品 ID 映射与完整测试版免付费解锁支持。
package license

import (
	"sync"
	"time"
)

// PlanType 定义用户的订阅或试用计划类型。
type PlanType string

const (
	PlanTrial    PlanType = "trial"     // 7 天全功能免费试用期
	PlanMonthly  PlanType = "monthly"   // 按月订阅会员（¥3 / 月，含 7 天免费试用）
	PlanYearly   PlanType = "yearly"    // 按年订阅会员（¥28 / 年，含 7 天免费试用）
	PlanFullTest PlanType = "full_test" // 完整测试版模式（免付费永久解锁全部功能）
)

const (
	// TrialDurationDays 定义新用户免费试用天数（7 天）。
	TrialDurationDays = 7
	// PriceMonthlyCNY 定义月度订阅人民币价格（3 元/月）。
	PriceMonthlyCNY = 3
	// PriceYearlyCNY 定义年度订阅人民币价格（28 元/年）。
	PriceYearlyCNY = 28

	// StoreIDMonthly 是 Microsoft Partner Center 中“月度订阅 Add-on (¥3/月)”的商品 Store ID。
	// 正式提交微软商店审核时，可在合作伙伴中心创建 Add-on 后替换为实际分配的 12 位 Store ID。
	StoreIDMonthly = "9N0000PRO_MONTHLY_3CNY"

	// StoreIDYearly 是 Microsoft Partner Center 中“年度订阅 Add-on (¥28/年)”的商品 Store ID。
	StoreIDYearly = "9N0000PRO_YEARLY_28CNY"

	// MicrosoftServicesURL 是微软官方账户订阅管理与取消订阅页面地址（符合 Microsoft Store 政策 10.8 要求）。
	MicrosoftServicesURL = "https://account.microsoft.com/services"
)

// Config 保存可持久化到 settings.json 的订阅与试用配置。
type Config struct {
	Plan                PlanType `json:"plan"`                // 当前计划类型："trial"、"monthly"、"yearly"、"full_test"
	TrialStartUnix      int64    `json:"trialStartUnix"`      // 首次开启 7 天免费试用的 Unix 秒级时间戳
	SubscriptionEndUnix int64    `json:"subscriptionEndUnix"` // 当前订阅周期到期时间戳（0 表示测试版无期限）
}

// Manager 线程安全地管理应用运行时的试用期计算、Pro 订阅激活与 Microsoft Store 授权状态恢复。
type Manager struct {
	mu         sync.RWMutex
	cfg        Config
	lastNotice string // 最近一次购买、切换或恢复购买的操作反馈提示码
}

// NewManager 根据持久化配置初始化授权管理器；若首次运行未记录试用时间，则以当前时间自动开启 7 天免费试用。
func NewManager(cfg Config, now time.Time) *Manager {
	if now.IsZero() {
		now = time.Now()
	}
	if cfg.Plan == "" {
		cfg.Plan = PlanTrial
	}
	if cfg.TrialStartUnix <= 0 {
		cfg.TrialStartUnix = now.Unix()
	}
	return &Manager{
		cfg: cfg,
	}
}

// Config 返回当前授权配置的快照。
func (m *Manager) Config() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

// Plan 返回当前激活的计划类型。
func (m *Manager) Plan() PlanType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg.Plan
}

// LastNotice 返回最近一次订阅或恢复操作的提示类型（"activated_monthly", "activated_yearly", "restored", "trial_reset", "full_test"）。
func (m *Manager) LastNotice() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastNotice
}

// ClearNotice 清除最近一次操作反馈提示。
func (m *Manager) ClearNotice() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastNotice = ""
}

// IsProActive 判断用户当前是否处于已开通的 Pro 会员状态（含月度订阅、年度订阅或完整测试版）。
func (m *Manager) IsProActive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg.Plan == PlanMonthly || m.cfg.Plan == PlanYearly || m.cfg.Plan == PlanFullTest
}

// IsFeatureUnlocked 判断当前是否允许使用全部高级功能。
// 规则：
// 1. 若处于 7 天免费试用期内 (TrialDaysRemaining > 0)，全功能免费解锁；
// 2. 若为按月/按年订阅 (PlanMonthly/PlanYearly)，检查到期时间戳是否在有效期内；
// 3. 若为 PlanFullTest（本地调试测试版模式），全功能免付费解锁；
// 4. 若试用已过期且未订阅，返回 false（锁定高级功能，提示开通）。
func (m *Manager) IsFeatureUnlocked(now time.Time) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if now.IsZero() {
		now = time.Now()
	}
	switch m.cfg.Plan {
	case PlanFullTest:
		return true
	case PlanMonthly, PlanYearly:
		if m.cfg.SubscriptionEndUnix == 0 {
			return true
		}
		return now.Unix() <= m.cfg.SubscriptionEndUnix
	case PlanTrial:
		return m.trialDaysRemainingLocked(now) > 0
	default:
		return m.trialDaysRemainingLocked(now) > 0
	}
}

// TrialDaysRemaining 计算 7 天免费试用期的剩余天数（范围 0 ~ 7）。
func (m *Manager) TrialDaysRemaining(now time.Time) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.trialDaysRemainingLocked(now)
}

func (m *Manager) trialDaysRemainingLocked(now time.Time) int {
	if now.IsZero() {
		now = time.Now()
	}
	start := time.Unix(m.cfg.TrialStartUnix, 0)
	elapsed := now.Sub(start)
	if elapsed <= 0 {
		return TrialDurationDays
	}
	totalTrial := time.Duration(TrialDurationDays) * 24 * time.Hour
	if elapsed >= totalTrial {
		return 0
	}
	remHours := (totalTrial - elapsed).Hours()
	days := int((remHours + 23.999) / 24.0)
	if days > TrialDurationDays {
		days = TrialDurationDays
	}
	if days < 1 {
		days = 1
	}
	return days
}

// ActivateMonthly 激活“按月订阅 (¥3/月)”方案（在测试版中免付费直接激活 31 天周期）。
func (m *Manager) ActivateMonthly(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if now.IsZero() {
		now = time.Now()
	}
	m.cfg.Plan = PlanMonthly
	m.cfg.SubscriptionEndUnix = now.Add(31 * 24 * time.Hour).Unix()
	m.lastNotice = "activated_monthly"
}

// ActivateYearly 激活“按年订阅 (¥28/年)”方案（在测试版中免付费直接激活 365 天周期）。
func (m *Manager) ActivateYearly(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if now.IsZero() {
		now = time.Now()
	}
	m.cfg.Plan = PlanYearly
	m.cfg.SubscriptionEndUnix = now.Add(365 * 24 * time.Hour).Unix()
	m.lastNotice = "activated_yearly"
}

// ResetToTrial 重置为全新的 7 天免费试用状态（供测试验证试用期 UI 与倒计时显示）。
func (m *Manager) ResetToTrial(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if now.IsZero() {
		now = time.Now()
	}
	m.cfg.Plan = PlanTrial
	m.cfg.TrialStartUnix = now.Unix()
	m.cfg.SubscriptionEndUnix = 0
	m.lastNotice = "trial_reset"
}

// ActivateFullTest 切换为永久免付费完整测试版模式。
func (m *Manager) ActivateFullTest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg.Plan = PlanFullTest
	m.cfg.SubscriptionEndUnix = 0
	m.lastNotice = "full_test"
}

// RestorePurchases 模拟向 Microsoft Store 刷新并恢复用户的历史订阅授权。
func (m *Manager) RestorePurchases(_ time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastNotice = "restored"
}
