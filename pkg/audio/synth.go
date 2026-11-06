// Package audio 提供纯 Go 零外部依赖的程序化音频合成器（WAV PCM）与原生多媒体播放引擎，
// 用于驱动闹钟响铃、倒计时结束提示音以及睡眠模式的白噪音/雨声/海浪循环播放与音量渐弱。
package audio

import (
	"encoding/binary"
	"math"
	"sync"
)

const (
	// SampleRate 标准合成采样率（22050 Hz 单声道 16-bit PCM，兼顾温暖音质与低内存占用）。
	SampleRate = 22050
)

var (
	cacheMu  sync.Mutex
	wavCache = make(map[string][]byte)
)

// GetAlarmWAV 获取或合成指定名称的闹钟铃声 WAV 字节流（结果缓存在内存中，后续调用零分配）。
func GetAlarmWAV(soundName string) []byte {
	key := "alarm:" + soundName
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if buf, ok := wavCache[key]; ok {
		return buf
	}
	var samples []int16
	switch soundName {
	case "Gentle Chime":
		samples = synthGentleChime()
	case "Digital Bell":
		samples = synthDigitalBell()
	case "Classic Chime":
		fallthrough
	default:
		samples = synthClassicChime()
	}
	wav := encodeWAV(samples, SampleRate)
	wavCache[key] = wav
	return wav
}

// GetAmbientWAV 获取或合成指定名称的助眠环境音无缝循环 WAV 字节流（结果缓存在内存中）。
func GetAmbientWAV(ambientName string) []byte {
	key := "ambient:" + ambientName
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if buf, ok := wavCache[key]; ok {
		return buf
	}
	var samples []int16
	switch ambientName {
	case "Ocean Waves":
		samples = synthOceanWaves()
	case "Forest Stream":
		samples = synthForestStream()
	case "White Noise":
		samples = synthWhiteNoise()
	case "Gentle Rain":
		fallthrough
	default:
		samples = synthGentleRain()
	}
	wav := encodeWAV(samples, SampleRate)
	wavCache[key] = wav
	return wav
}

// encodeWAV 将 16-bit 单声道 PCM 采样封装为标准 RIFF WAVE 内存字节流。
func encodeWAV(samples []int16, sampleRate int) []byte {
	dataSize := len(samples) * 2
	totalSize := 44 + dataSize
	buf := make([]byte, totalSize)

	// RIFF Header
	copy(buf[0:4], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:8], uint32(36+dataSize))
	copy(buf[8:12], "WAVE")

	// fmt Subchunk
	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16)                   // PCM chunk size
	binary.LittleEndian.PutUint16(buf[20:22], 1)                    // AudioFormat = 1 (PCM)
	binary.LittleEndian.PutUint16(buf[22:24], 1)                    // NumChannels = 1 (Mono)
	binary.LittleEndian.PutUint32(buf[24:28], uint32(sampleRate))   // SampleRate
	binary.LittleEndian.PutUint32(buf[28:32], uint32(sampleRate*2)) // ByteRate
	binary.LittleEndian.PutUint16(buf[32:34], 2)                    // BlockAlign
	binary.LittleEndian.PutUint16(buf[34:36], 16)                   // BitsPerSample = 16

	// data Subchunk
	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataSize))

	offset := 44
	for _, s := range samples {
		binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(s))
		offset += 2
	}
	return buf
}

// synthClassicChime 合成 3.6 秒的经典钟声琶音（C5 - E5 - G5 - C6 温暖泛音列）。
func synthClassicChime() []int16 {
	totalSec := 3.6
	n := int(totalSec * SampleRate)
	out := make([]float64, n)

	type note struct {
		startSec float64
		durSec   float64
		freq     float64
		amp      float64
	}
	notes := []note{
		{0.00, 1.20, 523.25, 0.45},  // C5
		{0.35, 1.20, 659.25, 0.45},  // E5
		{0.70, 1.20, 783.99, 0.45},  // G5
		{1.05, 1.90, 1046.50, 0.52}, // C6
		{1.05, 1.90, 523.25, 0.25},  // C5 根音共鸣
	}

	for _, nt := range notes {
		startIdx := int(nt.startSec * SampleRate)
		endIdx := int((nt.startSec + nt.durSec) * SampleRate)
		if endIdx > n {
			endIdx = n
		}
		for i := startIdx; i < endIdx; i++ {
			t := float64(i-startIdx) / SampleRate
			// 快速柔和起音 (15ms) + 自然指数衰减
			attack := 1.0
			if t < 0.015 {
				attack = t / 0.015
			}
			decay := math.Exp(-2.4 * t)
			// 钟琴泛音叠加：基频 + 二倍频 + 三倍频
			w := math.Sin(2*math.Pi*nt.freq*t) +
				0.35*math.Sin(2*math.Pi*nt.freq*2.0*t) +
				0.12*math.Sin(2*math.Pi*nt.freq*3.0*t)
			out[i] += nt.amp * attack * decay * w
		}
	}
	return floatToPCM16(out, 0.85)
}

// synthGentleChime 合成 4.2 秒的轻柔风铃/颂钵和弦（Fmaj9 渐进舒缓唤醒音）。
func synthGentleChime() []int16 {
	totalSec := 4.2
	n := int(totalSec * SampleRate)
	out := make([]float64, n)

	freqs := []struct {
		start float64
		freq  float64
		amp   float64
	}{
		{0.00, 349.23, 0.35}, // F4
		{0.45, 440.00, 0.35}, // A4
		{0.90, 523.25, 0.35}, // C5
		{1.35, 659.25, 0.38}, // E5
		{1.80, 783.99, 0.40}, // G5
	}

	for _, f := range freqs {
		startIdx := int(f.start * SampleRate)
		for i := startIdx; i < n; i++ {
			t := float64(i-startIdx) / SampleRate
			attack := 1.0
			if t < 0.08 {
				attack = t / 0.08 // 80ms 柔和渐入
			}
			decay := math.Exp(-1.35 * t)
			vibrato := 1.0 + 0.0015*math.Sin(2*math.Pi*4.5*t)
			w := math.Sin(2*math.Pi*f.freq*vibrato*t) + 0.20*math.Sin(2*math.Pi*f.freq*2.0*t)
			out[i] += f.amp * attack * decay * w
		}
	}
	return floatToPCM16(out, 0.80)
}

// synthDigitalBell 合成 2.4 秒的清脆现代电子双脉冲提示音（880Hz + 1760Hz）。
func synthDigitalBell() []int16 {
	totalSec := 2.4
	n := int(totalSec * SampleRate)
	out := make([]float64, n)

	pulses := []float64{0.05, 0.22, 0.40, 1.15, 1.32, 1.50}
	pulseDur := 0.11

	for _, pStart := range pulses {
		sIdx := int(pStart * SampleRate)
		eIdx := int((pStart + pulseDur) * SampleRate)
		if eIdx > n {
			eIdx = n
		}
		for i := sIdx; i < eIdx; i++ {
			t := float64(i-sIdx) / SampleRate
			env := 1.0
			if t < 0.008 {
				env = t / 0.008
			} else if t > pulseDur-0.025 {
				env = (pulseDur - t) / 0.025
			}
			w := 0.70*math.Sin(2*math.Pi*880.0*t) + 0.30*math.Sin(2*math.Pi*1760.0*t)
			out[i] += 0.65 * env * w
		}
	}
	return floatToPCM16(out, 0.85)
}

// synthGentleRain 合成 6.0 秒无缝循环的轻柔细雨环境声（粉红/褐噪滤波 + 细微水滴包络）。
func synthGentleRain() []int16 {
	totalSec := 6.0
	n := int(totalSec * SampleRate)
	raw := make([]float64, n)

	var lp1, lp2 float64
	var rng uint64 = 0x9E3779B97F4A7C15
	for i := range n {
		white := nextWhite(&rng)
		// 双级一阶低通滤波模拟柔和雨幕声
		lp1 += 0.16 * (white - lp1)
		lp2 += 0.22 * (lp1 - lp2)
		rain := lp2

		// 叠加稀疏柔和的细小雨滴声
		if (rng & 0x3FF) == 0 {
			rain += 0.18 * white
		}
		raw[i] = rain
	}
	seamLoopCrossfade(raw, SampleRate/4)
	return floatToPCM16(raw, 0.55)
}

// synthOceanWaves 合成 8.0 秒无缝循环的海岸波浪声（余弦低频包络调制海浪冲刷与退潮）。
func synthOceanWaves() []int16 {
	totalSec := 8.0
	n := int(totalSec * SampleRate)
	raw := make([]float64, n)

	var lp1, lp2 float64
	var rng uint64 = 0xBF58476D1CE4E5B9
	for i := range n {
		t := float64(i) / float64(n) // 0.0 -> 1.0 完整周期，首尾相位完全一致
		white := nextWhite(&rng)

		// 海浪涌起与退去的平滑余弦包络 (0.22 ~ 1.0)
		waveEnv := 0.61 - 0.39*math.Cos(2*math.Pi*t)
		// 浪尖到来时低通截止频率升高（浪花声更丰富），退潮时更深沉
		alpha := 0.04 + 0.18*waveEnv
		lp1 += alpha * (white - lp1)
		lp2 += (alpha * 0.8) * (lp1 - lp2)

		raw[i] = lp2 * waveEnv
	}
	seamLoopCrossfade(raw, SampleRate/4)
	return floatToPCM16(raw, 0.65)
}

// synthForestStream 合成 6.0 秒无缝循环的森林潺潺溪流声。
func synthForestStream() []int16 {
	totalSec := 6.0
	n := int(totalSec * SampleRate)
	raw := make([]float64, n)

	var lp1, lp2 float64
	var rng uint64 = 0x94D049BB133111EB
	for i := range n {
		t := float64(i) / float64(n)
		white := nextWhite(&rng)
		mod := 0.5 + 0.5*math.Sin(2*math.Pi*3.0*t+math.Sin(2*math.Pi*t))
		alpha := 0.12 + 0.14*mod
		lp1 += alpha * (white - lp1)
		lp2 += 0.18 * (lp1 - lp2)
		raw[i] = (lp1 - 0.45*lp2) * (0.75 + 0.25*mod)
	}
	seamLoopCrossfade(raw, SampleRate/4)
	return floatToPCM16(raw, 0.55)
}

// synthWhiteNoise 合成 6.0 秒无缝循环的温润助眠白噪音（经柔和低通平滑以消除刺耳高频毛刺）。
func synthWhiteNoise() []int16 {
	totalSec := 6.0
	n := int(totalSec * SampleRate)
	raw := make([]float64, n)

	var lp float64
	var rng uint64 = 0xD6E8FEB86659FD93
	for i := range n {
		white := nextWhite(&rng)
		lp += 0.32 * (white - lp)
		raw[i] = lp
	}
	seamLoopCrossfade(raw, SampleRate/4)
	return floatToPCM16(raw, 0.45)
}

// nextWhite 使用极速 SplitMix64 生成 [-1.0, 1.0] 均匀白噪声采样。
func nextWhite(state *uint64) float64 {
	*state += 0x9E3779B97F4A7C15
	z := *state
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	z = z ^ (z >> 31)
	return (float64(z>>11)*(1.0/float64(uint64(1)<<53)))*2.0 - 1.0
}

// seamLoopCrossfade 对循环音频的首尾窗口做平滑交叉淡入淡出，彻底消除循环播放接缝处的爆音。
func seamLoopCrossfade(buf []float64, fadeSamples int) {
	n := len(buf)
	if fadeSamples <= 0 || fadeSamples*2 >= n {
		return
	}
	for i := range fadeSamples {
		ratio := float64(i) / float64(fadeSamples)
		tailIdx := n - fadeSamples + i
		blended := buf[tailIdx]*(1.0-ratio) + buf[i]*ratio
		buf[tailIdx] = blended
	}
}

// floatToPCM16 将归一化浮点波形转换为 16-bit 有符号整型 PCM。
func floatToPCM16(samples []float64, targetPeak float64) []int16 {
	maxAbs := 0.0
	for _, v := range samples {
		a := math.Abs(v)
		if a > maxAbs {
			maxAbs = a
		}
	}
	scale := 1.0
	if maxAbs > 1e-6 {
		scale = (targetPeak * 32767.0) / maxAbs
	}
	out := make([]int16, len(samples))
	for i, v := range samples {
		s := v * scale
		if s > 32767.0 {
			s = 32767.0
		} else if s < -32768.0 {
			s = -32768.0
		}
		out[i] = int16(s)
	}
	return out
}
