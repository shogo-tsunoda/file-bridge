param(
  [Parameter(Mandatory = $true)]
  [string]$Version
)

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$binDir = Join-Path $root 'build/bin'
$releaseDir = Join-Path $root 'build/release'
$exeCandidates = @(
  (Join-Path $binDir 'FileBridge.exe'),
  (Join-Path $binDir 'file-bridge.exe')
)
$exe = $exeCandidates | Where-Object { Test-Path $_ } | Select-Object -First 1

if (-not $exe) {
  throw "Not found: FileBridge.exe (or file-bridge.exe). Run 'wails build' first."
}

if ($Version -notmatch '^\d+\.\d+\.\d+$') {
  throw 'Version must be semver without v prefix (example: 0.1.0)'
}

New-Item -ItemType Directory -Force -Path $releaseDir | Out-Null

$zipName = "FileBridge-v$Version-windows-amd64.zip"
$zipPath = Join-Path $releaseDir $zipName
if (Test-Path $zipPath) { Remove-Item $zipPath -Force }

Compress-Archive -Path $exe -DestinationPath $zipPath -Force

$checksums = @()
$zipHash = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToLower()
$checksums += "$zipHash  $zipName"

$installerCandidates = @(
  (Join-Path $binDir 'FileBridge-amd64-installer.exe'),
  (Join-Path $binDir 'file-bridge-amd64-installer.exe')
)
$installer = $installerCandidates | Where-Object { Test-Path $_ } | Select-Object -First 1
if ($installer) {
  $installerName = Split-Path $installer -Leaf
  $installerHash = (Get-FileHash -Path $installer -Algorithm SHA256).Hash.ToLower()
  $checksums += "$installerHash  $installerName"
}

$sumFile = Join-Path $releaseDir 'SHA256SUMS.txt'
$checksums | Set-Content -Path $sumFile -Encoding utf8

Write-Host "Created: $zipPath"
if ($installer) {
  Write-Host "Found installer: $installer"
}
Write-Host "Created: $sumFile"
