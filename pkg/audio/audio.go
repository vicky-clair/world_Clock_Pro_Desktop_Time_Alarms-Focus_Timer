package audio

import (
	"math"
	"sync"
)

// Engine 统一协调闹钟响铃、睡眠环境音循环以及手动试听音效的状态同步。
// 遵循 Viva 框架设计：仅在音频状态或音量发生实质变化时才触发底层系统调用，平帧开销为零。
// 采用“双缓冲内存固定 (Double-Buffered Memory Pinning) + 软硬件双重音量控制”架构，
// 确保在任何 Windows 电脑（包括蓝牙耳机、USB 声卡、显示器 HDMI 音频或无独立混音器的声卡驱动）上都能稳定发声且渐弱有效。
type Engine struct {
	mu              sync.Mutex
	activeKey       string    // 当前正在循环播放的音频标识（如 "alarm:Classic Chime" 或 "ambient:Gentle Rain"）
	activeBuf       []byte    // 持有当前正在异步播放的 WAV 内存切片引用，防止 GC 回收
	prevBuf         []byte    // 保留上一帧正在切换中的 WAV 内存切片引用，杜绝 winmm 异步线程在切换瞬间读取已释放内存
	scaledBufs      [2][]byte // 双缓冲软件音量缩放池，复用内存实现零分配 PCM 振幅调节
	bufSlot         int       // 当前使用的双缓冲槽位索引 (0 或 1)
	activeVolBucket int       // 当前循环音频对应的 10% 步进软件音量档位 (1 ~ 10)
	lastVolume      float32   // 上一次设置的硬件音量比例
	previewKey      string    // 当前正在试听的单次音频标识
}

// NewEngine 创建并初始化音频播放引擎。
func NewEngine() *Engine {
	return &Engine{
		lastVolume:      1.0,
		activeVolBucket: 10,
	}
}

// quantizeVolumeBucket 将 [0.0, 1.0] 音量映射为 1 ~ 10 的 10% 整数档位，用于驱动睡眠渐弱阶段的软件 PCM 振幅阶梯衰减。
func quantizeVolumeBucket(vol float32) int {
	b := int(math.Round(float64(vol) * 10.0))
	if b < 1 {
		b = 1
	} else if b > 10 {
		b = 10
	}
	return b
}

// preparePlaybackBufLocked 在互斥锁保护下生成指定软件音量的 WAV 缓冲区，并更新双缓冲引用链以防 GC 回收。
func (e *Engine) preparePlaybackBufLocked(rawWAV []byte, vol float32) []byte {
	e.prevBuf = e.activeBuf
	if math.Abs(float64(vol-1.0)) < 0.005 {
		e.activeBuf = rawWAV
		return rawWAV
	}
	e.bufSlot = 1 - e.bufSlot
	scaled := ScaleWAVVolumeInto(e.scaledBufs[e.bufSlot], rawWAV, vol)
	e.scaledBufs[e.bufSlot] = scaled
	e.activeBuf = scaled
	return scaled
}

// SyncState 在每帧绘制/更新时同步闹钟与睡眠模式的音频播放状态：
// 1. 闹钟响铃优先级最高（全音量循环播放指定闹钟铃声）；
// 2. 其次为睡眠模式处于 Active 或 Fading 状态（按实时衰减音量循环播放所选环境白噪音）；
// 3. 若两者均未激活且无手动试听，则立即停止循环播放。
func (e *Engine) SyncState(alarmRinging bool, alarmSound string, sleepActive bool, sleepSound string, sleepVolume float32) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if alarmRinging {
		if alarmSound == "" {
			alarmSound = "Classic Chime"
		}
		targetKey := "alarm:" + alarmSound
		if e.activeKey != targetKey || e.activeVolBucket != 10 {
			wav := e.preparePlaybackBufLocked(GetAlarmWAV(alarmSound), 1.0)
			e.activeKey = targetKey
			e.activeVolBucket = 10
			e.previewKey = ""
			e.setVolumeLocked(1.0)
			platformPlayMemoryWAV(wav, true)
		} else {
			e.setVolumeLocked(1.0)
		}
		return
	}

	if sleepActive && sleepSound != "" && sleepSound != "Mute (Silent)" && sleepVolume > 0.01 {
		targetKey := "ambient:" + sleepSound
		bucket := quantizeVolumeBucket(sleepVolume)
		if e.activeKey != targetKey || e.activeVolBucket != bucket {
			softVol := float32(bucket) / 10.0
			wav := e.preparePlaybackBufLocked(GetAmbientWAV(sleepSound), softVol)
			e.activeKey = targetKey
			e.activeVolBucket = bucket
			e.previewKey = ""
			e.setVolumeLocked(sleepVolume)
			platformPlayMemoryWAV(wav, true)
		} else {
			e.setVolumeLocked(sleepVolume)
		}
		return
	}

	// 若之前正在循环播放闹钟或助眠音，现已停止则先停止硬件发声，再释放缓冲区引用并恢复默认音量
	if e.activeKey != "" {
		platformStopSound()
		e.activeKey = ""
		e.activeVolBucket = 10
		e.prevBuf = e.activeBuf
		e.activeBuf = nil
		e.setVolumeLocked(1.0)
	}
}

// PreviewAlarm 手动单次试听指定的闹钟铃声。
func (e *Engine) PreviewAlarm(soundName string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if soundName == "" {
		soundName = "Classic Chime"
	}
	wav := e.preparePlaybackBufLocked(GetAlarmWAV(soundName), 0.9)
	e.activeKey = ""
	e.activeVolBucket = 9
	e.previewKey = "alarm:" + soundName
	e.setVolumeLocked(0.9)
	platformPlayMemoryWAV(wav, false)
}

// PreviewAmbient 手动单次试听指定的睡眠环境白噪音。
func (e *Engine) PreviewAmbient(ambientName string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if ambientName == "" || ambientName == "Mute (Silent)" {
		platformStopSound()
		e.activeKey = ""
		e.previewKey = ""
		e.prevBuf = e.activeBuf
		e.activeBuf = nil
		return
	}
	wav := e.preparePlaybackBufLocked(GetAmbientWAV(ambientName), 0.8)
	e.activeKey = ""
	e.activeVolBucket = 8
	e.previewKey = "ambient:" + ambientName
	e.setVolumeLocked(0.8)
	platformPlayMemoryWAV(wav, false)
}

// Stop 立即停止一切正在播放或试听的音频。
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.activeKey != "" || e.previewKey != "" {
		platformStopSound()
		e.activeKey = ""
		e.previewKey = ""
		e.activeVolBucket = 10
		e.prevBuf = e.activeBuf
		e.activeBuf = nil
		e.setVolumeLocked(1.0)
	}
}

// ActiveKey 返回当前正在循环播放的音频键（用于状态查询与单元测试断言）。
func (e *Engine) ActiveKey() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.activeKey
}

// PreviewKey 返回最近一次触发的单次试听音频键。
func (e *Engine) PreviewKey() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.previewKey
}

func (e *Engine) setVolumeLocked(vol float32) {
	if math.Abs(float64(e.lastVolume-vol)) < 0.01 {
		return
	}
	e.lastVolume = vol
	platformSetVolume(vol)
}
