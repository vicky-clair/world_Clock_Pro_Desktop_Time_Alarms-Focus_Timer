//go:build windows

package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	advapi32            = syscall.NewLazyDLL("advapi32.dll")
	procRegCreateKeyExW = advapi32.NewProc("RegCreateKeyExW")
	procRegSetValueExW  = advapi32.NewProc("RegSetValueExW")
	procRegDeleteValueW = advapi32.NewProc("RegDeleteValueW")
	procRegCloseKey     = advapi32.NewProc("RegCloseKey")
)

const (
	runSubKey             = `Software\Microsoft\Windows\CurrentVersion\Run`
	startupApprovedSubKey = `Software\Microsoft\Windows\CurrentVersion\Explorer\StartupApproved\Run`
)

// regSetString 向指定的注册表子项中写入 REG_SZ 字符串值。
func regSetString(subKey, name, value string) error {
	subKeyPtr, err := syscall.UTF16PtrFromString(subKey)
	if err != nil {
		return err
	}
	var hKey syscall.Handle
	var disposition uint32
	ret, _, _ := procRegCreateKeyExW.Call(
		uintptr(syscall.HKEY_CURRENT_USER),
		uintptr(unsafe.Pointer(subKeyPtr)),
		0, 0, 0,
		uintptr(syscall.KEY_SET_VALUE),
		0,
		uintptr(unsafe.Pointer(&hKey)),
		uintptr(unsafe.Pointer(&disposition)),
	)
	if ret != 0 {
		return syscall.Errno(ret)
	}
	defer procRegCloseKey.Call(uintptr(hKey))

	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	valUTF16, err := syscall.UTF16FromString(value)
	if err != nil {
		return err
	}
	ret, _, _ = procRegSetValueExW.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(namePtr)),
		0,
		syscall.REG_SZ,
		uintptr(unsafe.Pointer(&valUTF16[0])),
		uintptr(len(valUTF16)*2),
	)
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}

// regSetBinary 向指定的注册表子项中写入 REG_BINARY 二进制数据。
func regSetBinary(subKey, name string, data []byte) error {
	subKeyPtr, err := syscall.UTF16PtrFromString(subKey)
	if err != nil {
		return err
	}
	var hKey syscall.Handle
	var disposition uint32
	ret, _, _ := procRegCreateKeyExW.Call(
		uintptr(syscall.HKEY_CURRENT_USER),
		uintptr(unsafe.Pointer(subKeyPtr)),
		0, 0, 0,
		uintptr(syscall.KEY_SET_VALUE),
		0,
		uintptr(unsafe.Pointer(&hKey)),
		uintptr(unsafe.Pointer(&disposition)),
	)
	if ret != 0 {
		return syscall.Errno(ret)
	}
	defer procRegCloseKey.Call(uintptr(hKey))

	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	var ptr uintptr
	if len(data) > 0 {
		ptr = uintptr(unsafe.Pointer(&data[0]))
	}
	ret, _, _ = procRegSetValueExW.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(namePtr)),
		0,
		syscall.REG_BINARY,
		ptr,
		uintptr(len(data)),
	)
	if ret != 0 {
		return syscall.Errno(ret)
	}
	return nil
}

// regDeleteValue 删除指定注册表子项下的键值；若键值不存在则静默忽略。
func regDeleteValue(subKey, name string) error {
	subKeyPtr, err := syscall.UTF16PtrFromString(subKey)
	if err != nil {
		return err
	}
	var hKey syscall.Handle
	var disposition uint32
	ret, _, _ := procRegCreateKeyExW.Call(
		uintptr(syscall.HKEY_CURRENT_USER),
		uintptr(unsafe.Pointer(subKeyPtr)),
		0, 0, 0,
		uintptr(syscall.KEY_SET_VALUE),
		0,
		uintptr(unsafe.Pointer(&hKey)),
		uintptr(unsafe.Pointer(&disposition)),
	)
	if ret != 0 {
		return syscall.Errno(ret)
	}
	defer procRegCloseKey.Call(uintptr(hKey))

	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	ret, _, _ = procRegDeleteValueW.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(namePtr)),
	)
	// ERROR_FILE_NOT_FOUND (2) 表示键值原本就不存在，视为删除成功
	if ret != 0 && ret != 2 {
		return syscall.Errno(ret)
	}
	return nil
}

// syncSystemAutostartPlatform 针对 Windows 平台实现高可靠开机自动启动配置：
// 1. 自动清理历史旧版本遗留的废弃启动项键值（如 com.worldclockpro.desktop、OliveClock 等），杜绝开机报错或死锁；
// 2. 强制使用半角双引号包裹可执行文件绝对路径（"C:\...\小瑞综合时钟桌面.exe"），彻底解决包含中文字符或空格时 Windows Logon 解析失败的问题；
// 3. 同步写入 Windows 10/11 任务管理器及系统“启动应用”批准状态 (StartupApproved\Run)，标记为 0x02 (Enabled)，防止被系统静默拦截或推迟。
func syncSystemAutostartPlatform(enabled bool) error {
	// 无论启用还是禁用，先自动清理历史遗留的旧版启动项名称
	legacyNames := []string{"com.worldclockpro.desktop", "OliveClock", "WorldClock"}
	for _, oldName := range legacyNames {
		_ = regDeleteValue(runSubKey, oldName)
		_ = regDeleteValue(startupApprovedSubKey, oldName)
	}

	if !enabled {
		_ = regDeleteValue(runSubKey, AutostartAppID)
		_ = regDeleteValue(startupApprovedSubKey, AutostartAppID)
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return err
	}

	// 核心修复：Windows Run 注册表值强制添加半角双引号包裹，保证包含中文路径或空格时系统 100% 正确调用
	quotedCmd := fmt.Sprintf("\"%s\"", exe)
	if err := regSetString(runSubKey, AutostartAppID, quotedCmd); err != nil {
		return err
	}

	// 同步写入 Windows 10/11 任务管理器及设置中“启动应用”的已启用状态 (0x02, 12 bytes)
	enabledBytes := []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	_ = regSetBinary(startupApprovedSubKey, AutostartAppID, enabledBytes)

	return nil
}
