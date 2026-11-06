package clock

import (
	"testing"
	"time"
)

func BenchmarkFormatTime(b *testing.B) {
	now := time.Now()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = FormatTime(now, true, true)
	}
}

func BenchmarkComputeAnalogGeometry(b *testing.B) {
	now := time.Now()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = ComputeAnalogGeometry(100, 100, 80, now)
	}
}
