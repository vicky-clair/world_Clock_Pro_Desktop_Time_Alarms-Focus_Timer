#!/usr/bin/env bash
# 小瑞综合时钟桌面 - macOS 原生 Bundle (.app) 与苹果商店 (.pkg / .dmg) 打包脚本
# 运行环境：macOS (Darwin) 或 GitHub Actions macOS Runner

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
APP_NAME="小瑞综合时钟桌面"
BUNDLE_DIR="${DIST_DIR}/${APP_NAME}.app"

echo "==> [1/4] 创建 macOS App Bundle 目录结构..."
mkdir -p "${BUNDLE_DIR}/Contents/MacOS"
mkdir -p "${BUNDLE_DIR}/Contents/Resources"

echo "==> [2/4] 编译通用二进制 (Universal Binary: Apple Silicon arm64 + Intel x86_64)..."
cd "${ROOT_DIR}"

# 编译 arm64 (M1/M2/M3/M4)
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "${DIST_DIR}/bin_arm64" ./cmd/worldclock
# 编译 amd64 (Intel)
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "${DIST_DIR}/bin_amd64" ./cmd/worldclock

# 合并为通用二进制文件
lipo -create -output "${BUNDLE_DIR}/Contents/MacOS/XiaoruIntegratedClockDesktop" "${DIST_DIR}/bin_arm64" "${DIST_DIR}/bin_amd64"
chmod +x "${BUNDLE_DIR}/Contents/MacOS/XiaoruIntegratedClockDesktop"
rm -f "${DIST_DIR}/bin_arm64" "${DIST_DIR}/bin_amd64"

echo "==> [3/4] 注入 Info.plist 与图标资源..."
cp "${SCRIPT_DIR}/Info.plist" "${BUNDLE_DIR}/Contents/Info.plist"
echo "APPL????" > "${BUNDLE_DIR}/Contents/PkgInfo"

# 若存在 assets/icon.icns，则复制
if [ -f "${ROOT_DIR}/assets/icon.icns" ]; then
    cp "${ROOT_DIR}/assets/icon.icns" "${BUNDLE_DIR}/Contents/Resources/icon.icns"
fi

echo "==> [4/4] 验证 .app 打包结果..."
ls -lh "${BUNDLE_DIR}/Contents/MacOS"
echo "✅ macOS App Bundle 打包完成：${BUNDLE_DIR}"

# 如果在真实 macOS 环境且配置了开发者证书，可执行签名与商店 .pkg 构建：
if command -v codesign &> /dev/null; then
    echo "提示：可使用以下命令执行 Mac App Store 签名与打包："
    echo "  codesign --deep --force --verify --verbose --options runtime --entitlements packaging/macos/entitlements.plist --sign \"Apple Distribution: Your Name (ID)\" \"${BUNDLE_DIR}\""
    echo "  productbuild --component \"${BUNDLE_DIR}\" /Applications --sign \"3rd Party Mac Developer Installer: Your Name (ID)\" \"${DIST_DIR}/${APP_NAME}.pkg\""
fi
