<#
.SYNOPSIS
    Build FileBridge as an MSIX package for Microsoft Store or sideloading.

.DESCRIPTION
    1. Builds the Wails app (optional)
    2. Generates MSIX asset PNGs from appicon.png using ImageMagick
    3. Stamps version into AppxManifest.xml
    4. Packs everything into an MSIX with makeappx
    5. Optionally signs with a certificate (self-signed or provided)

.PARAMETER Version
    App version in X.Y.Z or X.Y.Z.W format. Defaults to 1.0.0.0.

.PARAMETER Publisher
    Publisher CN for the manifest Identity. For Store submission, use the CN
    assigned by Microsoft Partner Center. Defaults to "CN=shogo-tsunoda".

.PARAMETER SkipBuild
    Skip the Wails build step (use existing binary).

.PARAMETER Sign
    Sign the MSIX after packing.

.PARAMETER CertPath
    Path to an existing .pfx certificate. If omitted and -Sign is set,
    a new self-signed certificate is created.

.PARAMETER CertPassword
    Password for the .pfx certificate.

.EXAMPLE
    # Basic build
    .\build-msix.ps1 -Version 1.2.0

    # Build + sign with self-signed cert
    .\build-msix.ps1 -Version 1.2.0 -Sign

    # Build for Store with specific publisher
    .\build-msix.ps1 -Version 1.2.0 -Publisher "CN=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"

    # Skip build, sign with existing cert
    .\build-msix.ps1 -SkipBuild -Sign -CertPath .\cert.pfx -CertPassword "pass"
#>
param(
    [string]$Version = "1.0.0.0",
    [string]$Publisher = "CN=EB801A8D-2401-45BC-9A63-9D1DDEBC6283",
    [switch]$SkipBuild,
    [switch]$Sign,
    [string]$CertPath = "",
    [string]$CertPassword = ""
)

$ErrorActionPreference = "Stop"

# ---- Normalize version to 4-part ----
$versionParts = $Version -split "\."
while ($versionParts.Count -lt 4) { $versionParts += "0" }
$Version4 = ($versionParts[0..3]) -join "."
Write-Host "Version: $Version4" -ForegroundColor Cyan

# ---- Paths ----
$ProjectRoot = Resolve-Path "$PSScriptRoot\..\..\..\"
$MsixDir     = $PSScriptRoot
$StagingDir  = Join-Path $MsixDir "staging"
$AssetsDir   = Join-Path $StagingDir "Assets"
$OutputMsix  = Join-Path $ProjectRoot "build\bin\FileBridge.msix"
$IconPng     = Join-Path $ProjectRoot "build\appicon.png"
$IconIco     = Join-Path $ProjectRoot "build\windows\icon.ico"
$WailsBin    = Join-Path $ProjectRoot "build\bin\FileBridge.exe"
$ManifestSrc = Join-Path $MsixDir "AppxManifest.xml"

# ---- Step 1: Build Wails app ----
if (-not $SkipBuild) {
    Write-Host "`n==> Step 1/5: Building Wails app..." -ForegroundColor Cyan
    Push-Location $ProjectRoot
    try {
        wails build
        if ($LASTEXITCODE -ne 0) { throw "Wails build failed (exit code: $LASTEXITCODE)" }
    } finally {
        Pop-Location
    }
} else {
    Write-Host "`n==> Step 1/5: Skipping Wails build (-SkipBuild)" -ForegroundColor DarkGray
}

if (-not (Test-Path $WailsBin)) {
    throw "Binary not found at $WailsBin. Run without -SkipBuild first."
}

# ---- Step 2: Prepare staging directory ----
Write-Host "`n==> Step 2/5: Preparing staging directory..." -ForegroundColor Cyan
if (Test-Path $StagingDir) { Remove-Item $StagingDir -Recurse -Force }
New-Item -ItemType Directory -Path $AssetsDir -Force | Out-Null

# Copy binary
Copy-Item $WailsBin -Destination $StagingDir
Write-Host "   Copied FileBridge.exe" -ForegroundColor DarkGray

# Process manifest: replace placeholders
$manifestContent = Get-Content $ManifestSrc -Raw
$manifestContent = $manifestContent -replace '\{\{VERSION\}\}', $Version4
$manifestContent = $manifestContent -replace 'Publisher="CN=[^"]*"', "Publisher=`"$Publisher`""
$manifestDest = Join-Path $StagingDir "AppxManifest.xml"
Set-Content -Path $manifestDest -Value $manifestContent -Encoding UTF8
Write-Host "   Generated AppxManifest.xml (Version=$Version4, Publisher=$Publisher)" -ForegroundColor DarkGray

# ---- Step 3: Generate asset PNGs ----
Write-Host "`n==> Step 3/5: Generating MSIX asset images..." -ForegroundColor Cyan

$SourceImage = if (Test-Path $IconPng) { $IconPng } else { $IconIco }
if (-not (Test-Path $SourceImage)) {
    throw "No source icon found. Expected $IconPng or $IconIco"
}

$HasMagick = $null -ne (Get-Command "magick" -ErrorAction SilentlyContinue)

# Asset definitions: name => "WxH" or size (square)
$AssetSizes = [ordered]@{
    "StoreLogo.png"         = "50x50"
    "Square44x44Logo.png"   = "44x44"
    "SmallTile.png"         = "71x71"
    "Square150x150Logo.png" = "150x150"
    "Wide310x150Logo.png"   = "310x150"
    "Square310x310Logo.png" = "310x310"
    "SplashScreen.png"      = "620x300"
}

if ($HasMagick) {
    foreach ($name in $AssetSizes.Keys) {
        $size = $AssetSizes[$name]
        $outPath = Join-Path $AssetsDir $name
        $dims = $size -split "x"
        $w = $dims[0]; $h = $dims[1]

        if ($w -eq $h) {
            # Square: resize to fill
            magick $SourceImage -resize "${w}x${h}" -background "#0F172A" -gravity center -extent "${size}" $outPath
        } else {
            # Non-square (wide tile, splash): fit icon inside with padding
            $iconSize = [math]::Min([int]$w, [int]$h) - 20
            magick $SourceImage -resize "${iconSize}x${iconSize}" -background "#0F172A" -gravity center -extent "${size}" $outPath
        }

        if ($LASTEXITCODE -ne 0) { throw "ImageMagick failed for $name" }
        Write-Host "   Created $name ($size)" -ForegroundColor DarkGray
    }
} else {
    Write-Host "   WARNING: ImageMagick (magick) not found in PATH." -ForegroundColor Yellow
    Write-Host "   Install: winget install ImageMagick.ImageMagick" -ForegroundColor Yellow
    Write-Host "   Falling back to copying source image (replace with correct sizes later)." -ForegroundColor Yellow

    foreach ($name in $AssetSizes.Keys) {
        Copy-Item $SourceImage -Destination (Join-Path $AssetsDir $name)
    }
}

# ---- Step 4: Pack MSIX ----
Write-Host "`n==> Step 4/5: Packing MSIX..." -ForegroundColor Cyan

# Find makeappx.exe from Windows SDK
$makeappx = Get-ChildItem "C:\Program Files (x86)\Windows Kits\10\bin\*\x64\makeappx.exe" -ErrorAction SilentlyContinue |
    Sort-Object FullName -Descending |
    Select-Object -First 1

if (-not $makeappx) {
    $makeappx = Get-Command "makeappx" -ErrorAction SilentlyContinue
}

if (-not $makeappx) {
    throw @"
makeappx.exe not found. Install Windows 10/11 SDK:
  winget install Microsoft.WindowsSDK.10.0.26100
  -- or --
  https://developer.microsoft.com/en-us/windows/downloads/windows-sdk/
"@
}

$makeappxPath = if ($makeappx -is [System.IO.FileInfo]) { $makeappx.FullName } else { $makeappx.Source }

# Ensure output directory exists
$outputDir = Split-Path $OutputMsix -Parent
if (-not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
}

& $makeappxPath pack /d $StagingDir /p $OutputMsix /o
if ($LASTEXITCODE -ne 0) { throw "makeappx pack failed (exit code: $LASTEXITCODE)" }

Write-Host "   Created: $OutputMsix" -ForegroundColor Green

# ---- Step 5: Sign (optional) ----
if ($Sign) {
    Write-Host "`n==> Step 5/5: Signing MSIX..." -ForegroundColor Cyan

    $signtool = Get-ChildItem "C:\Program Files (x86)\Windows Kits\10\bin\*\x64\signtool.exe" -ErrorAction SilentlyContinue |
        Sort-Object FullName -Descending |
        Select-Object -First 1

    if (-not $signtool) {
        $signtool = Get-Command "signtool" -ErrorAction SilentlyContinue
    }

    if (-not $signtool) {
        throw "signtool.exe not found. Install Windows 10/11 SDK."
    }

    $signtoolPath = if ($signtool -is [System.IO.FileInfo]) { $signtool.FullName } else { $signtool.Source }

    if ($CertPath -eq "") {
        # Create self-signed certificate
        $CertPath = Join-Path $MsixDir "FileBridge-dev.pfx"
        $CertPassword = "FileBridge"

        Write-Host "   Creating self-signed certificate (Subject=$Publisher)..." -ForegroundColor DarkGray
        $cert = New-SelfSignedCertificate `
            -Type Custom `
            -Subject $Publisher `
            -KeyUsage DigitalSignature `
            -FriendlyName "FileBridge Dev Certificate" `
            -CertStoreLocation "Cert:\CurrentUser\My" `
            -TextExtension @("2.5.29.37={text}1.3.6.1.5.5.7.3.3", "2.5.29.19={text}")

        $securePassword = ConvertTo-SecureString -String $CertPassword -Force -AsPlainText
        Export-PfxCertificate -Cert "Cert:\CurrentUser\My\$($cert.Thumbprint)" -FilePath $CertPath -Password $securePassword | Out-Null

        Write-Host "   Certificate: $CertPath (password: $CertPassword)" -ForegroundColor DarkGray
        Write-Host "   NOTE: Install this certificate to 'Trusted People' store for sideloading:" -ForegroundColor Yellow
        Write-Host "         certutil -addstore TrustedPeople $CertPath" -ForegroundColor Yellow
    }

    $signArgs = @("sign", "/fd", "SHA256", "/a", "/f", $CertPath)
    if ($CertPassword -ne "") {
        $signArgs += @("/p", $CertPassword)
    }
    $signArgs += $OutputMsix

    & $signtoolPath @signArgs
    if ($LASTEXITCODE -ne 0) { throw "Signing failed (exit code: $LASTEXITCODE)" }

    Write-Host "   MSIX signed successfully." -ForegroundColor Green
} else {
    Write-Host "`n==> Step 5/5: Skipping signing (pass -Sign to enable)" -ForegroundColor DarkGray
}

# ---- Summary ----
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host " MSIX build complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host " Output:    $OutputMsix" -ForegroundColor White
Write-Host " Version:   $Version4" -ForegroundColor White
Write-Host " Publisher: $Publisher" -ForegroundColor White
Write-Host ""

if (-not $Sign) {
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "  - For local testing:  .\build-msix.ps1 -Version $Version -Sign" -ForegroundColor White
    Write-Host "  - For Store upload:   Upload $OutputMsix to Partner Center" -ForegroundColor White
    Write-Host "    (Store signs automatically; no -Sign needed)" -ForegroundColor DarkGray
} else {
    Write-Host "For local testing:" -ForegroundColor Cyan
    Write-Host "  1. Install cert:  certutil -addstore TrustedPeople `"$CertPath`"" -ForegroundColor White
    Write-Host "  2. Install app:   Double-click $OutputMsix" -ForegroundColor White
}
