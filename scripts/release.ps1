param(
  [Parameter(Mandatory = $true)]
  [string]$Version
)

$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$binDir = Join-Path $root 'build/bin'
$releaseDir = Join-Path $root 'build/release'
$exe = Join-Path $binDir 'file-bridge.exe'

if (-not (Test-Path $exe)) {
  throw "Not found: $exe. Run 'wails build' first."
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

$installer = Join-Path $binDir 'file-bridge-amd64-installer.exe'
if (Test-Path $installer) {
  $installerName = Split-Path $installer -Leaf
  $installerHash = (Get-FileHash -Path $installer -Algorithm SHA256).Hash.ToLower()
  $checksums += "$installerHash  $installerName"
}

$sumFile = Join-Path $releaseDir 'SHA256SUMS.txt'
$checksums | Set-Content -Path $sumFile -Encoding utf8

Write-Host "Created: $zipPath"
if (Test-Path $installer) {
  Write-Host "Found installer: $installer"
}
Write-Host "Created: $sumFile"
