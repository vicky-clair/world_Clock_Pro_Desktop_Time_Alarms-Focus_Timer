//go:build vivahead

// 当使用 -tags vivahead 编译时，自动注册 Viva 无头测试与命令行交互协议驱动（供自动化 UI 验证使用，正式发行版不包含此文件）。
package main

import (
	_ "github.com/vivaui/viva/headless/auto"
)
