# 小瑞综合时钟桌面 - Themida 保护与加壳自动化脚本
# 针对 Go 语言特性的 PE 加壳保护与签名重建

$ErrorActionPreference = "Stop"

$ThemidaDir = "C:\BLWJ\Themida_3.2.6.0_nCh2CZZRNLW"
$ThemidaExe = Join-Path $ThemidaDir "Themida.exe"

if (-not (Test-Path $ThemidaExe)) {
    Write-Error "未找到 Themida 可执行文件，请确认路径是否存在：$ThemidaExe"
}

$RootDir = (Resolve-Path "$PSScriptRoot\..\..").Path
$DistDir = Join-Path $RootDir "dist"
$InputExe = Join-Path $DistDir "小瑞综合时钟桌面.exe"
$OutputExe = Join-Path $DistDir "小瑞综合时钟桌面_Protected.exe"

if (-not (Test-Path $InputExe)) {
    Write-Host "[0/3] 未找到待保护的源程序，正在执行原始编译..." -ForegroundColor Yellow
    & "$RootDir\build_windows.ps1"
}

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host " [Themida PE 保护与代码混淆处理]" -ForegroundColor Cyan
Write-Host " 输入文件: $InputExe" -ForegroundColor Gray
Write-Host " 输出文件: $OutputExe" -ForegroundColor Gray
Write-Host "==========================================================" -ForegroundColor Cyan

# 提示杀毒软件敏感风险
Write-Host "【重要安全提醒】:" -ForegroundColor Yellow
Write-Host "  1. Themida 加壳会向 PE 头部注入反调试与虚拟化外壳，极易触发卡巴斯基、Defender 等" -ForegroundColor Yellow
Write-Host "     杀软的启发式报毒（如 Heur.Packer.Themida / VHO:HackTool 等）。" -ForegroundColor Yellow
Write-Host "  2. Go 语言自身调度器依赖特殊内存段，在 Themida GUI 中请避免勾选针对整个 Runtime 的激进虚拟化。" -ForegroundColor Yellow
Write-Host ""

# 启动 Themida
Write-Host "[1/3] 正在启动 Themida 进行加壳处理..." -ForegroundColor Cyan
# 若命令行加壳，可通过 /protect 或启动 GUI
Write-Host "提示: 如果未在 GUI 中预先保存 .tmd 项目文件，将自动启动 Themida 界面加载当前文件。" -ForegroundColor DarkGray
Start-Process -FilePath $ThemidaExe -ArgumentList "`"$InputExe`""

Write-Host "[2/3] 完成加壳后，请将生成的文件保存为: $OutputExe" -ForegroundColor Green
Write-Host ""
Write-Host "[3/3] 正在准备数字签名恢复工具..." -ForegroundColor Cyan
Write-Host "  因为 Themida 加壳会修改 PE 头，加壳后的程序必须重新执行 Authenticode 签名：" -ForegroundColor Gray
Write-Host "  执行签名命令: Set-AuthenticodeSignature -FilePath `"$OutputExe`" -Certificate (Get-ChildItem Cert:\CurrentUser\My -CodeSigningCert | Select-Object -First 1)" -ForegroundColor White
