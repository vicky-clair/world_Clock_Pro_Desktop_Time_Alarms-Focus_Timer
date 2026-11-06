package audio

import (
	"math"
	"sync"
)

// Engine 统一协调闹钟响铃、睡眠环境音循环以及手动试听音效的状态同步。
// 遵循 Viva 框架设计：仅在音频状态或音量发生实质变化时才触发底层系统调用，平帧开销为零。
type Engine struct {
	mu         sync.Mutex
	activeKey  string  // 当前正在循环播放的音频标识（如 "alarm:Classic Chime" 或 "ambient:Gentle Rain"）
	activeBuf  []byte  // 持有当前正在异步播放的 WAV 内存切片引用，防止 GC 回收
	lastVolume float32 // 上一次设置的硬件音量比例
	previewKey string  // 当前正在试听的单次音频标识
}

// NewEngine 创建并初始化音频播放引擎。
func NewEngine() *Engine {
	return &Engine{
		lastVolume: 1.0,
	}
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
		if e.activeKey != targetKey {
			wav := GetAlarmWAV(alarmSound)
			e.activeBuf = wav
			e.activeKey = targetKey
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
		if e.activeKey != targetKey {
			wav := GetAmbientWAV(sleepSound)
			e.activeBuf = wav
			e.activeKey = targetKey
			e.previewKey = ""
			e.setVolumeLocked(sleepVolume)
			platformPlayMemoryWAV(wav, true)
		} else {
			e.setVolumeLocked(sleepVolume)
		}
		return
	}

	// 若之前正在循环播放闹钟或助眠音，现已停止则关闭硬件播放並恢复默认音量
	if e.activeKey != "" {
		e.activeKey = ""
		e.activeBuf = nil
		platformStopSound()
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
	wav := GetAlarmWAV(soundName)
	e.activeKey = ""
	e.previewKey = "alarm:" + soundName
	e.activeBuf = wav
	e.setVolumeLocked(0.9)
	platformPlayMemoryWAV(wav, false)
}

// PreviewAmbient 手动单次试听指定的睡眠环境白噪音。
func (e *Engine) PreviewAmbient(ambientName string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if ambientName == "" || ambientName == "Mute (Silent)" {
		e.activeKey = ""
		e.previewKey = ""
		e.activeBuf = nil
		platformStopSound()
		return
	}
	wav := GetAmbientWAV(ambientName)
	e.activeKey = ""
	e.previewKey = "ambient:" + ambientName
	e.activeBuf = wav
	e.setVolumeLocked(0.8)
	platformPlayMemoryWAV(wav, false)
}

// Stop 立即停止一切正在播放或试听的音频。
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.activeKey != "" || e.previewKey != "" {
		e.activeKey = ""
		e.previewKey = ""
		e.activeBuf = nil
		platformStopSound()
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
