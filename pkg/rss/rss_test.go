package rss

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFeedStore(t *testing.T) {
	store := NewFeedStore()
	all := store.Articles("All")
	if len(all) < 3 {
		t.Fatalf("expected at least 3 curated history items, got %d", len(all))
	}

	tech := store.Articles("Tech")
	if len(tech) == 0 {
		t.Fatal("expected Tech history items")
	}
	for _, a := range tech {
		if a.Category != "Tech" {
			t.Errorf("expected Tech category, got %s", a.Category)
		}
	}

	// Deduplication test
	dup := store.AddArticle(Article{
		ID:          "hist-tech-1",
		Title:       "Duplicate",
		Category:    "Tech",
		Link:        "https://example.com/tech/gpu-frameworks",
		PublishedAt: time.Now(),
	})
	if dup {
		t.Fatal("expected duplicate article to be rejected")
	}

	// Add unique
	added := store.AddArticle(Article{
		ID:          "custom-1",
		Year:        "2026",
		Title:       "Brand New Milestone",
		Category:    "Tech",
		Link:        "https://example.com/tech/brand-new",
		PublishedAt: time.Now(),
	})
	if !added {
		t.Fatal("expected unique article to be added")
	}
}

func TestFetchHistoryTodayMockAPIs(t *testing.T) {
	mockSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/baidu/11.json" {
			_, _ = fmt.Fprint(w, `{
				"11": {
					"1106": [
						{
							"year": "1860",
							"title": "<a target=\"_blank\" href=\"https://baike.baidu.com/item/1\">亚伯拉罕·林肯</a>当选美国第16任总统",
							"desc": "1860年11月6日，亚伯拉罕·林肯当选美国总统。",
							"type": "event",
							"link": "https://baike.baidu.com/item/1"
						},
						{
							"year": "1962",
							"title": "联合国大会通过决议谴责种族隔离并推动科学合作",
							"desc": "推进全球科学与技术交流",
							"type": "event",
							"link": "https://baike.baidu.com/item/2"
						}
					]
				}
			}`)
			return
		}
		_, _ = fmt.Fprint(w, `{
			"events": [
				{
					"text": "Abraham Lincoln is elected the 16th President of the United States.",
					"year": 1860,
					"pages": [{"extract": "United States presidential election of 1860."}]
				}
			]
		}`)
	}))
	defer mockSrv.Close()

	store := NewFeedStore()
	store.baiduURLFmt = mockSrv.URL + "/baidu/%02d.json"
	store.wikiURLFmt = mockSrv.URL + "/wiki/%s/%02d/%02d"

	zhItems, src, err := store.FetchHistoryToday(context.Background(), 11, 6, "zh")
	if err != nil {
		t.Fatalf("FetchHistoryToday zh failed: %v", err)
	}
	if len(zhItems) != 2 {
		t.Fatalf("expected 2 zh items, got %d", len(zhItems))
	}
	if zhItems[1].Year != "1860" || zhItems[1].Title != "亚伯拉罕·林肯当选美国第16任总统" {
		t.Fatalf("unexpected cleaned title/year: %+v", zhItems[1])
	}
	if src == "" {
		t.Fatal("expected non-empty API source")
	}

	enItems, _, err := store.FetchHistoryToday(context.Background(), 11, 6, "en")
	if err != nil || len(enItems) != 1 {
		t.Fatalf("FetchHistoryToday en failed: err=%v, len=%d", err, len(enItems))
	}
}
