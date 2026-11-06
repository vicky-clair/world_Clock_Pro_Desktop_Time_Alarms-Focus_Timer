package clock

import (
	"fmt"
	"math"
	"time"
)

// ClockStyle defines how the clock is rendered.
type ClockStyle string

const (
	StyleDigital ClockStyle = "digital"
	StyleAnalog  ClockStyle = "analog"
)

// FormatTime formats a time into hours, minutes, and optional seconds based on 24h/12h preference.
func FormatTime(t time.Time, is24H, showSeconds bool) (timeStr string, period string) {
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()

	if !is24H {
		period = "AM"
		if hour >= 12 {
			period = "PM"
		}
		hour = hour % 12
		if hour == 0 {
			hour = 12
		}
	}

	if showSeconds {
		timeStr = fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
	} else {
		timeStr = fmt.Sprintf("%02d:%02d", hour, minute)
	}
	return timeStr, period
}

// FormatDate formats date according to full or short preference.
func FormatDate(t time.Time, full bool) string {
	if full {
		return t.Format("Monday, January 2, 2006")
	}
	return t.Format("Mon, Jan 2, 2006")
}

// NextBoundaryDelay computes the time duration until the next second (or next minute if seconds hidden).
func NextBoundaryDelay(now time.Time, showSeconds bool) time.Duration {
	if showSeconds {
		next := now.Truncate(time.Second).Add(time.Second)
		delay := next.Sub(now)
		if delay <= 0 {
			return 10 * time.Millisecond
		}
		return delay
	}
	next := now.Truncate(time.Minute).Add(time.Minute)
	delay := next.Sub(now)
	if delay <= 0 {
		return 100 * time.Millisecond
	}
	return delay
}

// HandCoord represents an endpoint for an analog clock hand or tick mark.
type HandCoord struct {
	X float32
	Y float32
}

// AnalogGeometry computes coordinates for an analog clock face.
type AnalogGeometry struct {
	CenterX float32
	CenterY float32
	Radius  float32

	HourHand   HandCoord
	MinuteHand HandCoord
	SecondHand HandCoord

	HourTicks   [12]HandCoord
	HourTickOut [12]HandCoord
}

// ComputeAnalogGeometry computes the positions for center, hands, and hour marks.
func ComputeAnalogGeometry(cx, cy, radius float32, t time.Time) AnalogGeometry {
	sec := float32(t.Second()) + float32(t.Nanosecond())/1e9
	min := float32(t.Minute()) + sec/60.0
	hr := float32(t.Hour()%12) + min/60.0

	// Angles in radians, 12 o'clock is -pi/2
	secAngle := (sec/60.0)*2*math.Pi - math.Pi/2
	minAngle := (min/60.0)*2*math.Pi - math.Pi/2
	hrAngle := (hr/12.0)*2*math.Pi - math.Pi/2

	hrLen := radius * 0.50
	minLen := radius * 0.75
	secLen := radius * 0.85

	geom := AnalogGeometry{
		CenterX: cx,
		CenterY: cy,
		Radius:  radius,
		HourHand: HandCoord{
			X: cx + hrLen*float32(math.Cos(float64(hrAngle))),
			Y: cy + hrLen*float32(math.Sin(float64(hrAngle))),
		},
		MinuteHand: HandCoord{
			X: cx + minLen*float32(math.Cos(float64(minAngle))),
			Y: cy + minLen*float32(math.Sin(float64(minAngle))),
		},
		SecondHand: HandCoord{
			X: cx + secLen*float32(math.Cos(float64(secAngle))),
			Y: cy + secLen*float32(math.Sin(float64(secAngle))),
		},
	}

	tickInner := radius * 0.85
	tickOuter := radius * 0.95
	for i := range 12 {
		angle := (float32(i)/12.0)*2*math.Pi - math.Pi/2
		cosA := float32(math.Cos(float64(angle)))
		sinA := float32(math.Sin(float64(angle)))
		geom.HourTicks[i] = HandCoord{
			X: cx + tickInner*cosA,
			Y: cy + tickInner*sinA,
		}
		geom.HourTickOut[i] = HandCoord{
			X: cx + tickOuter*cosA,
			Y: cy + tickOuter*sinA,
		}
	}

	return geom
}
