//go:build !windows

package audio

func platformPlayMemoryWAV(wav []byte, loop bool) {}

func platformStopSound() {}

func platformSetVolume(vol float32) {}
