package i18n

import (
	"testing"
)

func TestTranslations(t *testing.T) {
	// English check
	if s := T(LangEN, "nav.clock"); s != "Clock" {
		t.Errorf("expected Clock, got %s", s)
	}

	// Chinese check
	if s := T(LangZH, "nav.clock"); s != "时钟" {
		t.Errorf("expected 时钟, got %s", s)
	}

	// Japanese check
	if s := T(LangJA, "nav.clock"); s != "時計" {
		t.Errorf("expected 時計, got %s", s)
	}

	// Fallback to English for missing key in JA
	if s := T(LangJA, "nonexistent_key_xyz"); s != "nonexistent_key_xyz" {
		t.Errorf("expected fallback to key name, got %s", s)
	}
}

func TestFormatDateLocalized(t *testing.T) {
	// Thursday, Nov 5, 2026 -> weekday = 4 (Thursday)
	en := FormatDateLocalized(LangEN, 11, 5, 2026, 4, true)
	if en != "Thursday, November 5, 2026" {
		t.Errorf("unexpected en date: %s", en)
	}

	zh := FormatDateLocalized(LangZH, 11, 5, 2026, 4, true)
	if zh != "星期四，2026年11月5日" {
		t.Errorf("unexpected zh date: %s", zh)
	}

	ja := FormatDateLocalized(LangJA, 11, 5, 2026, 4, true)
	if ja != "2026年11月5日 (木曜日)" {
		t.Errorf("unexpected ja date: %s", ja)
	}

	jaShort := FormatDateLocalized(LangJA, 11, 5, 2026, 4, false)
	if jaShort != "2026/11/5 (木)" {
		t.Errorf("unexpected ja short date: %s", jaShort)
	}
}
