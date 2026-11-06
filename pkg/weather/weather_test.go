package weather

import (
	"testing"
)

func TestFormatTemperature(t *testing.T) {
	c := float32(20.0)
	cStr := FormatTemperature(c, false)
	if cStr != "20°C" {
		t.Errorf("expected 20°C, got %s", cStr)
	}

	fStr := FormatTemperature(c, true)
	if fStr != "68°F" {
		t.Errorf("expected 68°F, got %s", fStr)
	}

	cZero := float32(0.0)
	fZero := FormatTemperature(cZero, true)
	if fZero != "32°F" {
		t.Errorf("expected 32°F for 0C, got %s", fZero)
	}
}

func TestWeatherService(t *testing.T) {
	svc := NewService()
	snap := svc.GetWeather("beijing")
	if snap.CityID != "beijing" || snap.TempCelsius <= 0 {
		t.Errorf("unexpected weather snapshot: %+v", snap)
	}

	// Test fallback generation for non-preset city
	snapCustom := svc.GetWeather("custom-city-xyz")
	if snapCustom.CityID != "custom-city-xyz" || snapCustom.Icon == "" {
		t.Errorf("unexpected custom snapshot: %+v", snapCustom)
	}
}
