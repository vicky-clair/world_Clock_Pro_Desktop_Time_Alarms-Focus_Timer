// Package rss 管理“历史上的今天”（On This Day in History）历史事件流与在线公开 API 同步，
// 支持百度百科历史上的今天 API 与 Wikipedia OnThisDay API 双源拉取、离线精选库兜底以及按类别过滤。
package rss

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// Article 表示单条“历史上的今天”事件或精选纪事条目。
// Article represents an "On This Day in History" event item.
type Article struct {
	ID          string    `json:"id"`          // 事件唯一 ID
	Year        string    `json:"year"`        // 历史年份（如 "1860", "1991", "2007"）
	Title       string    `json:"title"`       // 事件标题
	Source      string    `json:"source"`      // 数据来源名称（如 "百度百科 API", "Wikipedia API"）
	Category    string    `json:"category"`    // 所属分类 ("World"=历史大事, "Tech"=科技探索, "Science"=人文纪事)
	Link        string    `json:"link"`        // 详情百科链接
	Summary     string    `json:"summary"`     // 历史事件详细描述
	PublishedAt time.Time `json:"publishedAt"` // 记录或同步时间戳
}

// FeedStore 并发安全地管理“历史上的今天”事件集合，支持在线 API 异步拉取、分类检索与去重插入。
type FeedStore struct {
	mu          sync.RWMutex
	articles    []Article
	apiStatus   string // 当前数据源状态标识 ("online", "fetching", "curated")
	apiSource   string // 最近成功同步的 API 名称
	fetchedKey  string // 已完成拉取的 "MM-DD-lang" 缓存键，避免同一天重复请求
	isFetching  bool   // 是否正在后台执行网络请求
	baiduURLFmt string // 可在测试中覆写的百度百科 API URL 模板
	wikiURLFmt  string // 可在测试中覆写的 Wikipedia API URL 模板
}

// NewFeedStore 创建并初始化带有离线精选历史事件的 FeedStore 实例。
func NewFeedStore() *FeedStore {
	store := &FeedStore{
		apiStatus:   "curated",
		apiSource:   "Offline Curated Archive",
		baiduURLFmt: "https://baike.baidu.com/cms/home/eventsOnHistory/%02d.json",
		wikiURLFmt:  "https://%s.wikipedia.org/api/rest_v1/feed/onthisday/events/%02d/%02d",
	}
	store.populateCurated()
	return store
}

// populateCurated 初始化内置精选“历史上的今天”事件，确保无网络或请求返回前界面依然饱满直观。
func (s *FeedStore) populateCurated() {
	now := time.Now()
	s.articles = []Article{
		{
			ID:          "hist-world-1",
			Year:        "1676",
			Title:       "Danish Astronomer Ole Rømer First Demonstrates the Finite Speed of Light",
			Source:      "History Archive",
			Category:    "World",
			Link:        "https://en.wikipedia.org/wiki/Speed_of_light",
			Summary:     "By observing eclipses of Jupiter's moon Io, Ole Rømer proved that light travels at a finite speed, transforming modern physics and timekeeping.",
			PublishedAt: now.Add(-10 * time.Minute),
		},
		{
			ID:          "hist-tech-1",
			Year:        "1967",
			Title:       "International System of Units (SI) Defines the Atomic Second via Cesium-133",
			Source:      "Chrono & Tech Archive",
			Category:    "Tech",
			Link:        "https://example.com/tech/gpu-frameworks",
			Summary:     "The 13th General Conference on Weights and Measures officially defined the second using 9,192,631,770 cycles of radiation from a Cesium-133 atom.",
			PublishedAt: now.Add(-20 * time.Minute),
		},
		{
			ID:          "hist-sci-1",
			Year:        "1884",
			Title:       "International Meridian Conference Establishes Greenwich Mean Time & 24 Global Time Zones",
			Source:      "World Heritage",
			Category:    "Science",
			Link:        "https://en.wikipedia.org/wiki/International_Meridian_Conference",
			Summary:     "Delegates from 25 nations gathered to adopt the Greenwich Meridian as the universal prime meridian and standardize global 24-hour timekeeping.",
			PublishedAt: now.Add(-45 * time.Minute),
		},
		{
			ID:          "hist-tech-2",
			Year:        "1993",
			Title:       "CERN Releases the World Wide Web Protocol and Source Code into the Public Domain",
			Source:      "Tech Milestones",
			Category:    "Tech",
			Link:        "https://en.wikipedia.org/wiki/History_of_the_World_Wide_Web",
			Summary:     "Tim Berners-Lee and CERN made the World Wide Web technology freely available to humanity without royalties, sparking the global internet era.",
			PublishedAt: now.Add(-2 * time.Hour),
		},
		{
			ID:          "hist-world-2",
			Year:        "2000",
			Title:       "First Permanent Crew Docks with the International Space Station (Expedition 1)",
			Source:      "Space & World History",
			Category:    "World",
			Link:        "https://en.wikipedia.org/wiki/Expedition_1",
			Summary:     "Soyuz TM-31 docked with the International Space Station, beginning an uninterrupted era of human presence in low Earth orbit.",
			PublishedAt: now.Add(-3 * time.Hour),
		},
	}
}

// Status 返回当前数据源同步状态 ("online", "fetching", "curated") 与来源名称。
func (s *FeedStore) Status() (status string, source string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.apiStatus, s.apiSource
}

// EnsureFetchedAsync 检查当日指定语言的“历史上的今天”API 数据是否已拉取；若未拉取则在后台发起一次非阻塞 HTTP 请求。
// 遵循 Viva 零轮询准则：仅在请求完成时调用一次 onComplete 唤醒 UI 重绘。
func (s *FeedStore) EnsureFetchedAsync(now time.Time, lang string, force bool, onComplete func()) {
	if testing.Testing() && !force {
		return
	}
	month := int(now.Month())
	day := now.Day()
	if lang == "" {
		lang = "zh"
	}
	key := fmt.Sprintf("%02d-%02d-%s", month, day, lang)

	s.mu.Lock()
	if s.isFetching || (!force && s.fetchedKey == key) {
		s.mu.Unlock()
		return
	}
	s.isFetching = true
	s.apiStatus = "fetching"
	s.mu.Unlock()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()

		items, src, err := s.FetchHistoryToday(ctx, month, day, lang)

		s.mu.Lock()
		s.isFetching = false
		if err == nil && len(items) > 0 {
			s.articles = items
			s.apiStatus = "online"
			s.apiSource = src
			s.fetchedKey = key
		} else {
			if s.apiStatus == "fetching" {
				s.apiStatus = "curated"
			}
			// 标记当前 key 已尝试，避免离线环境下每帧重复发起网络请求；用户可随时点击“刷新”强制重试
			s.fetchedKey = key
		}
		s.mu.Unlock()

		if onComplete != nil {
			onComplete()
		}
	}()
}

// FetchHistoryToday 同步调用公开 API 拉取指定月、日的“历史上的今天”事件列表。
// 中文模式优先调用百度百科官方历史上的今天接口（国内直连秒开），回退使用中文维基百科接口；
// 英文与日文模式优先调用对应语种的 Wikipedia OnThisDay REST API，回退使用百度百科接口。
func (s *FeedStore) FetchHistoryToday(ctx context.Context, month, day int, lang string) ([]Article, string, error) {
	s.mu.RLock()
	baiduFmt := s.baiduURLFmt
	wikiFmt := s.wikiURLFmt
	s.mu.RUnlock()

	if lang == "zh" {
		if items, err := fetchBaiduHistory(ctx, baiduFmt, month, day); err == nil && len(items) > 0 {
			return items, "百度百科 API (baike.baidu.com)", nil
		}
		if items, err := fetchWikipediaHistory(ctx, wikiFmt, "zh", month, day); err == nil && len(items) > 0 {
			return items, "Wikipedia OnThisDay API (zh)", nil
		}
	} else {
		wikiLang := "en"
		if lang == "ja" {
			wikiLang = "ja"
		}
		if items, err := fetchWikipediaHistory(ctx, wikiFmt, wikiLang, month, day); err == nil && len(items) > 0 {
			return items, fmt.Sprintf("Wikipedia OnThisDay API (%s)", wikiLang), nil
		}
		if items, err := fetchBaiduHistory(ctx, baiduFmt, month, day); err == nil && len(items) > 0 {
			return items, "Baidu Baike History API", nil
		}
	}
	return nil, "", fmt.Errorf("no history API reachable for %02d-%02d (%s)", month, day, lang)
}

type baiduRawEvent struct {
	Year  string `json:"year"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Type  string `json:"type"`
	Link  string `json:"link"`
}

func fetchBaiduHistory(ctx context.Context, urlFmt string, month, day int) ([]Article, error) {
	url := fmt.Sprintf(urlFmt, month)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) WorldClockPro/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("baidu status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, err
	}

	var monthMap map[string]map[string][]baiduRawEvent
	if err := json.Unmarshal(body, &monthMap); err != nil {
		return nil, err
	}

	mmKey := fmt.Sprintf("%02d", month)
	mmddKey := fmt.Sprintf("%02d%02d", month, day)
	dayMap, ok := monthMap[mmKey]
	if !ok {
		return nil, fmt.Errorf("month key %s not found", mmKey)
	}
	rawList, ok := dayMap[mmddKey]
	if !ok || len(rawList) == 0 {
		return nil, fmt.Errorf("day key %s empty", mmddKey)
	}

	now := time.Now()
	var out []Article
	// 倒序遍历使近现代重大历史事件排在前面，更贴近阅读习惯
	for i := len(rawList) - 1; i >= 0; i-- {
		ev := rawList[i]
		title := cleanHTMLText(ev.Title)
		desc := cleanHTMLText(ev.Desc)
		if title == "" {
			continue
		}
		if desc == "" {
			desc = title
		} else if !strings.HasSuffix(desc, "。") && !strings.HasSuffix(desc, ".") {
			desc += "……"
		}
		cat := classifyHistoryCategory(ev.Type, title+" "+desc)
		out = append(out, Article{
			ID:          fmt.Sprintf("baidu-%s-%d", mmddKey, i),
			Year:        strings.TrimSpace(ev.Year),
			Title:       title,
			Source:      "百度百科 · 历史上的今天",
			Category:    cat,
			Link:        ev.Link,
			Summary:     desc,
			PublishedAt: now,
		})
	}
	return out, nil
}

type wikiResponse struct {
	Events []struct {
		Text  string `json:"text"`
		Year  int    `json:"year"`
		Pages []struct {
			Title       string `json:" normalizedtitle"`
			Extract     string `json:"extract"`
			Description string `json:"description"`
		} `json:"pages"`
	} `json:"events"`
}

func fetchWikipediaHistory(ctx context.Context, urlFmt string, wikiLang string, month, day int) ([]Article, error) {
	url := fmt.Sprintf(urlFmt, wikiLang, month, day)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "WorldClockPro/1.0 (Desktop Time & History Suite)")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wikipedia status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, err
	}

	var parsed wikiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Events) == 0 {
		return nil, fmt.Errorf("no events in wikipedia response")
	}

	now := time.Now()
	var out []Article
	limit := len(parsed.Events)
	if limit > 30 {
		limit = 30
	}
	for i := range limit {
		ev := parsed.Events[i]
		title := cleanHTMLText(ev.Text)
		if title == "" {
			continue
		}
		summary := title
		if len(ev.Pages) > 0 && ev.Pages[0].Extract != "" {
			summary = cleanHTMLText(ev.Pages[0].Extract)
			if len([]rune(summary)) > 160 {
				r := []rune(summary)
				summary = string(r[:160]) + "..."
			}
		}
		cat := classifyHistoryCategory("event", title+" "+summary)
		// 保证三个分类标签下均有可筛选内容
		if cat == "World" && i%4 == 1 {
			cat = "Tech"
		} else if cat == "World" && i%4 == 2 {
			cat = "Science"
		}
		out = append(out, Article{
			ID:          fmt.Sprintf("wiki-%s-%02d%02d-%d", wikiLang, month, day, i),
			Year:        fmt.Sprintf("%d", ev.Year),
			Title:       title,
			Source:      fmt.Sprintf("Wikipedia (%s)", strings.ToUpper(wikiLang)),
			Category:    cat,
			Link:        url,
			Summary:     summary,
			PublishedAt: now,
		})
	}
	return out, nil
}

// classifyHistoryCategory 将历史事件自动归类为 "Tech"（科技探索）、"Science"（人文与人物）或 "World"（历史大事）。
func classifyHistoryCategory(evType string, text string) string {
	lower := strings.ToLower(text)
	techKeywords := []string{
		"科技", "科学", "发明", "发现", "卫星", "火箭", "飞船", "太空", "天文", "计算机",
		"互联网", "物理", "化学", "医学", "铁路", "飞机", "诺贝尔", "专利", "核", "电",
		"science", "technology", "invent", "space", "launch", "satellite", "nasa",
		"computer", "internet", "physics", "chemistry", "medical", "railway", "aircraft", "nobel",
		"科学", "技術", "発明", "宇宙", "衛星", "ロケット", "ノーベル", "鉄道", "コンピュータ",
	}
	for _, kw := range techKeywords {
		if strings.Contains(lower, kw) {
			return "Tech"
		}
	}
	if strings.EqualFold(evType, "birth") || strings.EqualFold(evType, "death") {
		return "Science"
	}
	cultureKeywords := []string{
		"出生", "诞辰", "逝世", "去世", "大学", "文学", "艺术", "音乐", "电影", "奥运", "博物",
		"born", "died", "birth", "university", "art", "music", "olympic", "museum", "novel",
		"誕生", "死去", "大学", "芸術", "音楽", "オリンピック", "文学",
	}
	for _, kw := range cultureKeywords {
		if strings.Contains(lower, kw) {
			return "Science"
		}
	}
	return "World"
}

// cleanHTMLText 剥离百度百科或维基百科 JSON 字符串中的内嵌 HTML 标签（如 <a href=...>...</a>）与多余空白。
func cleanHTMLText(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	inTag := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			b.WriteByte(ch)
		}
	}
	res := b.String()
	res = strings.ReplaceAll(res, "&nbsp;", " ")
	res = strings.ReplaceAll(res, "&quot;", "\"")
	res = strings.ReplaceAll(res, "&amp;", "&")
	return strings.TrimSpace(res)
}

// Articles 返回按分类（"All"、"Tech"、"World"、"Science"）过滤后的历史事件切片副本。
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
	// 若在线 API 当天返回的数据恰巧在某个子分类下为空，则回退展示全部事件，避免页面空白
	if len(res) == 0 && len(s.articles) > 0 {
		res = make([]Article, len(s.articles))
		copy(res, s.articles)
	}
	return res
}

// AddArticle 向列表头部插入一条新事件；若 ID 或 Link 已存在则返回 false 拒绝重复插入。
func (s *FeedStore) AddArticle(a Article) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.articles {
		if item.ID == a.ID || (a.Link != "" && item.Link == a.Link) {
			return false
		}
	}
	s.articles = append([]Article{a}, s.articles...)
	return true
}
