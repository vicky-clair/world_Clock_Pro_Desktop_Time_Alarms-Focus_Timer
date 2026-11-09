//go:build windows

package ui

import (
	"syscall"
	"testing"
	"unsafe"
)

func TestWindowsAutostartRegistryOperations(t *testing.T) {
	testKeyName := "com.xiaoru.clockdesktop.unittest"
	testValStr := `"C:\Program Files\XiaoruClock\XiaoruClock.exe"`

	// 1. 测试写入 REG_SZ 字符串值
	if err := regSetString(runSubKey, testKeyName, testValStr); err != nil {
		t.Fatalf("regSetString failed: %v", err)
	}

	// 验证确实写入成功
	subKeyPtr, _ := syscall.UTF16PtrFromString(runSubKey)
	var hKey syscall.Handle
	ret, _, _ := procRegCreateKeyExW.Call(
		uintptr(syscall.HKEY_CURRENT_USER),
		uintptr(unsafe.Pointer(subKeyPtr)),
		0, 0, 0,
		uintptr(syscall.KEY_QUERY_VALUE),
		0,
		uintptr(unsafe.Pointer(&hKey)),
		0,
	)
	if ret != 0 {
		t.Fatalf("failed to open registry key to verify: %v", syscall.Errno(ret))
	}
	defer procRegCloseKey.Call(uintptr(hKey))

	namePtr, _ := syscall.UTF16PtrFromString(testKeyName)
	var valType uint32
	var bufSize uint32 = 512
	buf := make([]uint16, 256)
	err := syscall.RegQueryValueEx(hKey, namePtr, nil, &valType, (*byte)(unsafe.Pointer(&buf[0])), &bufSize)
	if err != nil {
		t.Fatalf("RegQueryValueEx failed: %v", err)
	}
	if valType != syscall.REG_SZ {
		t.Fatalf("expected REG_SZ type, got %d", valType)
	}
	readStr := syscall.UTF16ToString(buf)
	if readStr != testValStr {
		t.Fatalf("expected %q, got %q", testValStr, readStr)
	}

	// 2. 测试删除键值
	if err := regDeleteValue(runSubKey, testKeyName); err != nil {
		t.Fatalf("regDeleteValue failed: %v", err)
	}

	// 验证删除后再查返回错误
	err = syscall.RegQueryValueEx(hKey, namePtr, nil, &valType, (*byte)(unsafe.Pointer(&buf[0])), &bufSize)
	if err == nil {
		t.Fatal("expected key to be deleted, but still found")
	}

	// 3. 测试删除不存在的键值（静默成功幂等性）
	if err := regDeleteValue(runSubKey, "non.existent.random.key.12345"); err != nil {
		t.Fatalf("expected nil for non-existent key delete, got: %v", err)
	}
}

func TestWindowsAutostartBinaryOperations(t *testing.T) {
	testKeyName := "com.xiaoru.clockdesktop.unittest.binary"
	testBytes := []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	// 测试写入 REG_BINARY
	if err := regSetBinary(startupApprovedSubKey, testKeyName, testBytes); err != nil {
		t.Fatalf("regSetBinary failed: %v", err)
	}

	// 清理
	_ = regDeleteValue(startupApprovedSubKey, testKeyName)
}
