//go:build !windows

package audio

// platformPlayMemoryWAV 非 Windows 平台下的内存 WAV 音频播放空实现桩（保证跨平台编译与单元测试兼容）。
func platformPlayMemoryWAV(wav []byte, loop bool) {}

// platformStopSound 非 Windows 平台下的停止音频播放空实现桩。
func platformStopSound() {}

// platformSetVolume 非 Windows 平台下的硬件音量调节空实现桩。
func platformSetVolume(vol float32) {}
