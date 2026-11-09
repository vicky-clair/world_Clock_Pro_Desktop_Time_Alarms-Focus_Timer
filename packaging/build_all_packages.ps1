# 小瑞综合时钟桌面 - 一键构建全套 Windows 安装包与发布产物
$ErrorActionPreference = "Stop"

$RootDir = (Resolve-Path "$PSScriptRoot\..").Path
$DistDir = Join-Path $RootDir "dist"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host " [小瑞综合时钟桌面] 全套 Windows 分发包自动化编译" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. 编译带原生 PE 资源与数字签名的绿色版 EXE 及 ZIP
Write-Host "`n>>> [1/3] 编译 Windows 原生可执行程序及绿色 ZIP 包..." -ForegroundColor Yellow
& "$RootDir\build_windows.ps1"

# 2. 编译 WiX v4 标准 MSI 安装包
Write-Host "`n>>> [2/3] 使用 WiX Toolset v4 编译企业级 .msi 安装包..." -ForegroundColor Yellow
& "$PSScriptRoot\wix\build_msi.ps1"

# 3. 编译 Inno Setup 向导安装包 (Setup.exe)
Write-Host "`n>>> [3/3] 使用 Inno Setup 6 编译 Windows Setup 安装程序..." -ForegroundColor Yellow
& "$PSScriptRoot\innosetup\build_installer.ps1"

Write-Host "`n==========================================================" -ForegroundColor Green
Write-Host "  全套 Windows 安装包已全部生成至 dist 目录：" -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Green

Get-ChildItem -Path $DistDir -Filter "*小瑞*" | ForEach-Object {
    Write-Host ("  - {0,-36} ({1,8:N2} MiB)" -f $_.Name, ($_.Length / 1MB)) -ForegroundColor White
}
