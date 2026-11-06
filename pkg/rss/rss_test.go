package rss

import (
	"testing"
	"time"
)

func TestFeedStore(t *testing.T) {
	store := NewFeedStore()
	all := store.Articles("All")
	if len(all) < 3 {
		t.Fatalf("expected at least 3 curated articles, got %d", len(all))
	}

	tech := store.Articles("Tech")
	if len(tech) == 0 {
		t.Fatal("expected Tech articles")
	}
	for _, a := range tech {
		if a.Category != "Tech" {
			t.Errorf("expected Tech category, got %s", a.Category)
		}
	}

	// Deduplication test
	dup := store.AddArticle(Article{
		ID:          "tech-1",
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
		Title:       "Brand New Article",
		Category:    "Tech",
		Link:        "https://example.com/tech/brand-new",
		PublishedAt: time.Now(),
	})
	if !added {
		t.Fatal("expected unique article to be added")
	}
}
