// Package rss 管理内置与自定义资讯文章列表，支持按类别（全部/科技/国际/科学）过滤与去重添加。
package rss

import (
	"strings"
	"sync"
	"time"
)

// Article 表示单篇新闻资讯条目。
// Article represents a news item.
type Article struct {
	ID          string    `json:"id"`          // 文章唯一 ID
	Title       string    `json:"title"`       // 文章标题
	Source      string    `json:"source"`      // 来源媒体名称
	Category    string    `json:"category"`    // 所属分类 ("Tech", "World", "Science")
	Link        string    `json:"link"`        // 原文链接
	Summary     string    `json:"summary"`     // 摘要内容
	PublishedAt time.Time `json:"publishedAt"` // 发布时间
}

// FeedStore 并发安全地管理资讯文章集合，支持分类检索与去重插入。
// FeedStore manages RSS news articles with category filtering and deduplication.
type FeedStore struct {
	mu       sync.RWMutex
	articles []Article
}

// NewFeedStore 创建并填充精选资讯文章的 FeedStore 实例。
// NewFeedStore creates a news feed store populated with curated news items.
func NewFeedStore() *FeedStore {
	store := &FeedStore{}
	store.populateCurated()
	return store
}

// populateCurated 初始化内置精选科技、国际时讯与科学文章。
func (s *FeedStore) populateCurated() {
	now := time.Now()
	s.articles = []Article{
		{
			ID:          "tech-1",
			Title:       "Next-Generation High Performance GPU Native Desktop Frameworks",
			Source:      "Tech Horizons",
			Category:    "Tech",
			Link:        "https://example.com/tech/gpu-frameworks",
			Summary:     "How modern Go desktop architectures achieve zero-allocation painting and instant 60fps responses.",
			PublishedAt: now.Add(-15 * time.Minute),
		},
		{
			ID:          "world-1",
			Title:       "Global Time Zone Synchronization Standards Updated for 2026",
			Source:      "World Chrono",
			Category:    "World",
			Link:        "https://example.com/world/timezones",
			Summary:     "International bureau releases latest IANA timezone updates covering daylight saving transitions.",
			PublishedAt: now.Add(-45 * time.Minute),
		},
		{
			ID:          "sci-1",
			Title:       "Deep Space Telescope Uncovers New Pulsar Timekeeper In Distant Galaxy",
			Source:      "Science Daily",
			Category:    "Science",
			Link:        "https://example.com/science/pulsars",
			Summary:     "Astronomers confirm natural atomic-clock regularity in newly identified cosmic pulsar signals.",
			PublishedAt: now.Add(-2 * time.Hour),
		},
		{
			ID:          "tech-2",
			Title:       "Energy Efficiency in Desktop Apps: Zero-Polling Idle Architectures",
			Source:      "Dev Insights",
			Category:    "Tech",
			Link:        "https://example.com/tech/zero-polling",
			Summary:     "Benchmarking battery savings when desktop applications sleep soundly between user events.",
			PublishedAt: now.Add(-3 * time.Hour),
		},
	}
}

// Articles 返回按分类（"All"、"Tech"、"World"、"Science"）过滤后的资讯文章切片副本。
// Articles returns articles filtered by category ("All", "Tech", "World", "Science").
func (s *FeedStore) Articles(category string) []Article {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if category == "" || strings.EqualFold(category, "All") {
		res := make([]Article, len(s.articles))
		copy(res, s.articles)
		return res
	}

	var res []Article
	for _, a := range s.articles {
		if strings.EqualFold(a.Category, category) {
			res = append(res, a)
		}
	}
	return res
}

// AddArticle 向列表头部插入一篇新文章；若 ID 或 Link 已存在则返回 false 拒绝重复插入。
// AddArticle adds an article if not duplicate by link or ID.
func (s *FeedStore) AddArticle(a Article) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.articles {
		if item.ID == a.ID || item.Link == a.Link {
			return false
		}
	}
	s.articles = append([]Article{a}, s.articles...)
	return true
}
