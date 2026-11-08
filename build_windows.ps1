# World Clock Pro: Desktop Time, Alarms & Focus Timer - Windows Clean Build & Authenticode Signing Script
# Avoids AV heuristic false-positives (e.g., VHO:Trojan-PSW.Win32.Stealer.gen) by:
# 1. Linking native COFF .syso PE resources (RT_MANIFEST asInvoker, RT_VERSION, RT_GROUP_ICON) via Go linker
# 2. Preserving Go symbol table & BuildID (avoiding -s -w -buildid= obfuscation heuristics)
# 3. Applying an Authenticode digital signature if a local code-signing certificate is available

$ErrorActionPreference = "Stop"

$OutDir = Join-Path $PSScriptRoot "dist"
if (-not (Test-Path $OutDir)) {
    New-Item -ItemType Directory -Path $OutDir | Out-Null
}

$ExePath = Join-Path $OutDir "WorldClockPro.exe"
$RootExePath = Join-Path $PSScriptRoot "WorldClockPro.exe"
$ZipPath = Join-Path $OutDir "WorldClockPro-Windows-x64.zip"

Write-Host "[1/4] Building WorldClockPro.exe with native PE resources (.syso)..." -ForegroundColor Cyan
go build -ldflags="-H=windowsgui" -o $ExePath .\cmd\worldclock

Write-Host "[2/4] Applying Authenticode signature..." -ForegroundColor Cyan
$cert = Get-ChildItem Cert:\CurrentUser\My -CodeSigningCert | Where-Object { $_.Subject -like "*World Clock Pro*" } | Select-Object -First 1
if (-not $cert) {
    $cert = New-SelfSignedCertificate -Type CodeSigningCert `
        -Subject "CN=World Clock Pro Studio, O=World Clock Pro Studio, C=CN" `
        -CertStoreLocation "Cert:\CurrentUser\My" `
        -NotAfter (Get-Date).AddYears(5)
}
Set-AuthenticodeSignature -FilePath $ExePath -Certificate $cert -HashAlgorithm SHA256 | Out-Null
Copy-Item -Force $ExePath $RootExePath

Write-Host "[3/4] Creating release package zip..." -ForegroundColor Cyan
if (Test-Path $ZipPath) {
    Remove-Item -Force $ZipPath
}
Compress-Archive -Path $ExePath, (Join-Path $PSScriptRoot "README.md") -DestinationPath $ZipPath -Force

Write-Host "[4/4] Verifying output binary..." -ForegroundColor Cyan
$item = Get-Item $ExePath
$md5 = (Get-FileHash -Algorithm MD5 $ExePath).Hash
$sha256 = (Get-FileHash -Algorithm SHA256 $ExePath).Hash

Write-Host ("  File      : {0}" -f $item.FullName) -ForegroundColor Green
Write-Host ("  Size      : {0:N0} B ({1:N2} MiB)" -f $item.Length, ($item.Length / 1MB)) -ForegroundColor Green
Write-Host ("  Version   : {0} ({1})" -f $item.VersionInfo.ProductVersion, $item.VersionInfo.CompanyName) -ForegroundColor Green
Write-Host ("  MD5       : {0}" -f $md5) -ForegroundColor Green
Write-Host ("  SHA256    : {0}" -f $sha256) -ForegroundColor Green
