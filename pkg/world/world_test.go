package world

import (
	"testing"
	"time"
)

func TestBuiltinCities(t *testing.T) {
	cities := BuiltinCities()
	if len(cities) < 10 {
		t.Errorf("expected at least 10 builtin cities, got %d", len(cities))
	}
	bj, ok := FindCityByID("beijing")
	if !ok || bj.IanaTimeZoneID != "Asia/Shanghai" {
		t.Errorf("failed to find beijing: %+v", bj)
	}
}

func TestManagerSnapshots(t *testing.T) {
	mgr := NewManager(nil)
	utc := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	local := utc

	snaps := mgr.Snapshots(utc, local, true, true, "en")
	if len(snaps) != 4 {
		t.Fatalf("expected 4 default snapshots, got %d", len(snaps))
	}

	for _, s := range snaps {
		if s.TimeDisplay == "" {
			t.Errorf("city %s has empty time display", s.DisplayName)
		}
		if s.OffsetStr == "" {
			t.Errorf("city %s has empty offset", s.DisplayName)
		}
	}

	// Test Chinese localization
	snapsZH := mgr.Snapshots(utc, local, true, true, "zh")
	if len(snapsZH) != 4 {
		t.Fatalf("expected 4 snapshots, got %d", len(snapsZH))
	}
	if snapsZH[0].DisplayName != "北京" || snapsZH[0].DisplayCountry != "中国" {
		t.Errorf("expected 北京, 中国, got %s, %s", snapsZH[0].DisplayName, snapsZH[0].DisplayCountry)
	}
}

func TestManagerCityOperations(t *testing.T) {
	mgr := NewManager(nil)
	initialLen := len(mgr.Cities())

	added := mgr.AddCity("paris")
	if !added || len(mgr.Cities()) != initialLen+1 {
		t.Errorf("failed to add paris")
	}

	dup := mgr.AddCity("paris")
	if dup {
		t.Errorf("should not allow duplicate city addition")
	}

	mgr.ToggleFavorite("paris")
	found := false
	for _, c := range mgr.Cities() {
		if c.CityID == "paris" && c.IsFavorite {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("paris should be favorite")
	}

	mgr.MoveUp("paris")
	lastIdx := len(mgr.Cities()) - 1
	if mgr.Cities()[lastIdx].CityID == "paris" {
		t.Errorf("paris should have moved up from last position")
	}

	removed := mgr.RemoveCity("paris")
	if !removed || len(mgr.Cities()) != initialLen {
		t.Errorf("failed to remove paris")
	}

	// Verify original default cities can be removed both by legacy ID ("beijing") and canonical ID ("china")
	if !mgr.RemoveCity("china") {
		t.Errorf("expected to remove default city 'beijing' via canonical ID 'china'")
	}
	if !mgr.RemoveCity("tokyo") {
		t.Errorf("expected to remove default city 'tokyo'")
	}
	if !mgr.RemoveCity("london") {
		t.Errorf("expected to remove default city 'london'")
	}
	if !mgr.RemoveCity("newyork") {
		t.Errorf("expected to remove default city 'newyork'")
	}
	if len(mgr.Cities()) != 0 {
		t.Errorf("expected 0 cities after removing all defaults, got %d", len(mgr.Cities()))
	}

	// Verify empty non-nil slice does not re-populate default cities
	mgrEmpty := NewManager(mgr.Cities())
	if len(mgrEmpty.Cities()) != 0 {
		t.Errorf("expected empty manager to remain empty, got %d", len(mgrEmpty.Cities()))
	}
}

func TestSearchCatalog(t *testing.T) {
	// Search English
	results := SearchCatalog("Lon", "")
	if len(results) == 0 {
		t.Errorf("search for 'Lon' should return results")
	}

	// Search Chinese country
	resultsZH := SearchCatalog("法国", "")
	if len(resultsZH) == 0 || resultsZH[0].ID != "france" {
		t.Errorf("search for '法国' should return France, got %v", resultsZH)
	}

	// Filter continent Asia
	asia := SearchCatalog("", "亚洲")
	if len(asia) != 46 {
		t.Errorf("expected 46 countries in Asia, got %d", len(asia))
	}

	// Verify 193 global countries
	g193 := Global193Countries()
	if len(g193) != 193 {
		t.Errorf("expected 193 global countries, got %d", len(g193))
	}

	all := SearchCatalog("", "")
	if len(all) < 193 {
		t.Errorf("expected at least 193 entries, got %d", len(all))
	}

	// Search major non-capital cities in Chinese, English, and Japanese
	for _, tc := range []struct {
		query string
		want  string
	}{
		{"纽约", "newyork"},
		{"New York", "newyork"},
		{"上海", "shanghai"},
		{"Shanghai", "shanghai"},
		{"洛杉矶", "losangeles"},
		{"悉尼", "sydney"},
		{"多伦多", "toronto"},
		{"東京", "japan"},
	} {
		found := false
		for _, c := range SearchCatalog(tc.query, "") {
			if c.ID == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SearchCatalog(%q) did not find city ID %q", tc.query, tc.want)
		}
	}
}
