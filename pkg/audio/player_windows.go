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
)

const (
	sndAsync     = 0x0001
	sndNodefault = 0x0002
	sndMemory    = 0x0004
	sndLoop      = 0x0008
)

// platformPlayMemoryWAV 调用 Windows 原生 winmm.dll PlaySoundW 异步播放内存中的 WAV 字节流。
// 在 go test 单元测试环境下自动跳过硬件发声，避免测试时扬声器频繁鸣响。
func platformPlayMemoryWAV(wav []byte, loop bool) {
	if testing.Testing() || len(wav) == 0 {
		return
	}
	flags := uintptr(sndMemory | sndAsync | sndNodefault)
	if loop {
		flags |= sndLoop
	}
	_, _, _ = procPlaySoundW.Call(
		uintptr(unsafe.Pointer(&wav[0])),
		0,
		flags,
	)
}

// platformStopSound 立即停止当前异步播放的内存音频。
func platformStopSound() {
	if testing.Testing() {
		return
	}
	_, _, _ = procPlaySoundW.Call(0, 0, 0)
}

// platformSetVolume 动态设置当前进程音频输出流音量（vol 范围 0.0 ~ 1.0）。
func platformSetVolume(vol float32) {
	if testing.Testing() {
		return
	}
	if vol < 0 {
		vol = 0
	} else if vol > 1 {
		vol = 1
	}
	v16 := uint32(vol * 0xFFFF)
	stereoVol := v16 | (v16 << 16)
	_, _, _ = procWaveOutSetVolume.Call(0, uintptr(stereoVol))
}
