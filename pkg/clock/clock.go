// Package clock 提供核心时间格式化、整秒/整分精确重绘边界计算以及模拟表盘指针几何坐标计算。
package clock

import (
	"math"
	"time"
)

// ClockStyle 定义主时钟的渲染样式（数字时钟或模拟表盘）。
// ClockStyle defines how the clock is rendered.
type ClockStyle string

const (
	StyleDigital ClockStyle = "digital" // 数字时钟模式
	StyleAnalog  ClockStyle = "analog"  // 模拟表盘指针模式
)

// FormatTime 根据 24 小时制/12 小时制及是否显示秒数，使用栈上定长字节缓冲格式化给定时间（避免 fmt.Sprintf 逃逸分配）。
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

	var buf [8]byte
	buf[0] = byte('0' + (hour/10)%10)
	buf[1] = byte('0' + hour%10)
	buf[2] = ':'
	buf[3] = byte('0' + (minute/10)%10)
	buf[4] = byte('0' + minute%10)
	if showSeconds {
		buf[5] = ':'
		buf[6] = byte('0' + (second/10)%10)
		buf[7] = byte('0' + second%10)
		return string(buf[:8]), period
	}
	return string(buf[:5]), period
}

// FormatDate 按照完整格式或简短格式输出英文日期字符串。
// FormatDate formats date according to full or short preference.
func FormatDate(t time.Time, full bool) string {
	if full {
		return t.Format("Monday, January 2, 2006")
	}
	return t.Format("Mon, Jan 2, 2006")
}

// NextBoundaryDelay 计算距离下一个整秒（若不显示秒数则为下一个整分）的精确等待时长。
// 用于遵守 Viva 框架的“空闲零重绘”准则，无需后台轮询定时器。
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

// HandCoord 表示模拟表盘指针端点或刻度线的二维坐标。
// HandCoord represents an endpoint for an analog clock hand or tick mark.
type HandCoord struct {
	X float32
	Y float32
}

// AnalogGeometry 保存模拟时钟表盘的圆心、半径、时/分/秒针端点以及 12 个整点刻度线坐标。
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

// ComputeAnalogGeometry 根据圆心 (cx, cy)、半径 radius 和当前时间 t 计算表盘指针与刻度坐标。
// ComputeAnalogGeometry computes the positions for center, hands, and hour marks.
func ComputeAnalogGeometry(cx, cy, radius float32, t time.Time) AnalogGeometry {
	sec := float32(t.Second()) + float32(t.Nanosecond())/1e9
	min := float32(t.Minute()) + sec/60.0
	hr := float32(t.Hour()%12) + min/60.0

	// 角度转换为弧度，12 点钟方向对应 -pi/2
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

	// 计算 12 个整点刻度线的内外端点
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
