//go:build windows

package audio

import (
	"syscall"
	"testing"
	"unsafe"
)

var (
	winmm                = syscall.NewLazyDLL("winmm.dll")
	procPlaySoundW       = winmm.NewProc("PlaySoundW")
	procWaveOutSetVolume = winmm.NewProc("waveOutSetVolume")

	user32          = syscall.NewLazyDLL("user32.dll")
	procMessageBeep = user32.NewProc("MessageBeep")
)

const (
	sndAsync     = 0x0001
	sndNodefault = 0x0002
	sndMemory    = 0x0004
	sndLoop      = 0x0008
	waveMapper   = 0xFFFFFFFF // Windows WAVE_MAPPER 默认音频输出设备句柄标识
)

// platformPlayMemoryWAV 调用 Windows 原生 winmm.dll PlaySoundW 异步播放内存中的 WAV 字节流。
// 兼容性保障：
// 1. 调用前执行 Find() 校验，避免在精简版系统上触发异常；
// 2. 若目标电脑未插入任何音频输出设备导致 PlaySoundW 返回 0，自动回退至 MessageBeep 系统级蜂鸣提示，绝不阻塞或崩溃；
// 3. 在 go test 单元测试环境下自动跳过硬件发声，避免测试时扬声器频繁鸣响。
func platformPlayMemoryWAV(wav []byte, loop bool) {
	if testing.Testing() || len(wav) == 0 {
		return
	}
	if procPlaySoundW.Find() != nil {
		return
	}
	flags := uintptr(sndMemory | sndAsync | sndNodefault)
	if loop {
		flags |= sndLoop
	}
	ok, _, _ := procPlaySoundW.Call(
		uintptr(unsafe.Pointer(&wav[0])),
		0,
		flags,
	)
	if ok == 0 && procMessageBeep.Find() == nil {
		_, _, _ = procMessageBeep.Call(0xFFFFFFFF)
	}
}

// platformStopSound 立即停止当前异步播放的内存音频。
func platformStopSound() {
	if testing.Testing() {
		return
	}
	if procPlaySoundW.Find() != nil {
		return
	}
	_, _, _ = procPlaySoundW.Call(0, 0, 0)
}

// platformSetVolume 动态设置当前进程音频输出流音量（vol 范围 0.0 ~ 1.0）。
// 同时向 WAVE_MAPPER (默认输出设备) 与设备 0 发送音量指令，配合软件层 PCM 振幅缩放实现全硬件兼容。
func platformSetVolume(vol float32) {
	if testing.Testing() {
		return
	}
	if procWaveOutSetVolume.Find() != nil {
		return
	}
	if vol < 0 {
		vol = 0
	} else if vol > 1 {
		vol = 1
	}
	v16 := uint32(vol * 0xFFFF)
	stereoVol := uintptr(v16 | (v16 << 16))
	_, _, _ = procWaveOutSetVolume.Call(uintptr(waveMapper), stereoVol)
	_, _, _ = procWaveOutSetVolume.Call(0, stereoVol)
}
