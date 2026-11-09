# 小瑞综合时钟桌面 - Inno Setup 自动化构建 Setup.exe 安装程序脚本
$ErrorActionPreference = "Stop"

$RootDir = (Resolve-Path "$PSScriptRoot\..\..").Path
$DistDir = Join-Path $RootDir "dist"
$IssPath = Join-Path $PSScriptRoot "installer.iss"
$SetupPath = Join-Path $DistDir "小瑞综合时钟桌面_Setup.exe"
$ExePath = Join-Path $DistDir "小瑞综合时钟桌面.exe"

if (-not (Test-Path $ExePath)) {
    Write-Host "未找到主程序 EXE，正在执行编译..." -ForegroundColor Yellow
    & "$RootDir\build_windows.ps1"
}

Write-Host "[1/3] 正在使用 Inno Setup 6 编译安装向导程序..." -ForegroundColor Cyan
Set-Location $RootDir
iscc $IssPath

Write-Host "[2/3] 正在应用 Authenticode 数字签名..." -ForegroundColor Cyan
$cert = Get-ChildItem Cert:\CurrentUser\My -CodeSigningCert | Where-Object { $_.Subject -like "*Xiaoru Studio*" } | Select-Object -First 1
if ($cert) {
    Set-AuthenticodeSignature -FilePath $SetupPath -Certificate $cert -HashAlgorithm SHA256 | Out-Null
}

Write-Host "[3/3] 安装向导程序构建完成！" -ForegroundColor Green
$item = Get-Item $SetupPath
Write-Host ("  文件路径 : {0}" -f $item.FullName) -ForegroundColor Green
Write-Host ("  文件大小 : {0:N0} 字节 ({1:N2} MiB)" -f $item.Length, ($item.Length / 1MB)) -ForegroundColor Green
