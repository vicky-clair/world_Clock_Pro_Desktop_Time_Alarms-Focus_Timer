# 小瑞综合时钟桌面 - WiX v4 自动化构建 MSI 安装包脚本
$ErrorActionPreference = "Stop"

$RootDir = (Resolve-Path "$PSScriptRoot\..\..").Path
$DistDir = Join-Path $RootDir "dist"
$WxsPath = Join-Path $PSScriptRoot "package.wxs"
$MsiPath = Join-Path $DistDir "小瑞综合时钟桌面.msi"
$ExePath = Join-Path $DistDir "小瑞综合时钟桌面.exe"

if (-not (Test-Path $ExePath)) {
    Write-Host "未找到主程序 EXE，正在执行编译..." -ForegroundColor Yellow
    & "$RootDir\build_windows.ps1"
}

Write-Host "[1/3] 正在使用 WiX Toolset v4 编译 MSI 安装包..." -ForegroundColor Cyan
Set-Location $RootDir
wix build $WxsPath -arch x64 -o $MsiPath

Write-Host "[2/3] 正在应用 Authenticode 数字签名..." -ForegroundColor Cyan
$cert = Get-ChildItem Cert:\CurrentUser\My -CodeSigningCert | Where-Object { $_.Subject -like "*Xiaoru Studio*" } | Select-Object -First 1
if ($cert) {
    Set-AuthenticodeSignature -FilePath $MsiPath -Certificate $cert -HashAlgorithm SHA256 | Out-Null
}

Write-Host "[3/3] MSI 安装包构建完成！" -ForegroundColor Green
$item = Get-Item $MsiPath
Write-Host ("  文件路径 : {0}" -f $item.FullName) -ForegroundColor Green
Write-Host ("  文件大小 : {0:N0} 字节 ({1:N2} MiB)" -f $item.Length, ($item.Length / 1MB)) -ForegroundColor Green
