# MSIX パッケージング & Microsoft Store 提出ガイド

## 概要

FileBridge を MSIX パッケージとしてビルドし、Microsoft Store に提出するための手順書。

Wails v2 には公式の MSIX サポートがないため、`wails build` → EXE → `makeappx pack` の方式で MSIX 化している。

## 必要ツール

| ツール | 用途 | インストール |
|---|---|---|
| Go (1.24+) | Wails ビルド | `winget install GoLang.Go` |
| Node.js (18+) | フロントエンドビルド | `winget install OpenJS.NodeJS.LTS` |
| Wails CLI v2 | アプリビルド | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| Windows 10/11 SDK | makeappx, signtool | `winget install Microsoft.WindowsSDK.10.0.26100` |
| ImageMagick | アセット画像生成 | `winget install ImageMagick.ImageMagick` |

## ディレクトリ構成

```
build/
├── appicon.png                          # ソースアイコン (512x512以上推奨)
├── bin/
│   ├── FileBridge.exe                   # wails build で生成
│   └── FileBridge.msix                  # build-msix.ps1 で生成
└── windows/
    ├── icon.ico
    └── msix/
        ├── AppxManifest.xml             # テンプレート (Git管理)
        ├── build-msix.ps1               # ビルドスクリプト (Git管理)
        ├── FileBridge-dev.pfx           # 自己署名証明書 (gitignore)
        └── staging/                     # ビルド時生成 (gitignore)
            ├── AppxManifest.xml         # バージョン/Publisher 展開済み
            ├── FileBridge.exe
            └── Assets/
                ├── StoreLogo.png         (50x50)
                ├── Square44x44Logo.png   (44x44)
                ├── SmallTile.png         (71x71)
                ├── Square150x150Logo.png (150x150)
                ├── Wide310x150Logo.png   (310x150)
                ├── Square310x310Logo.png (310x310)
                └── SplashScreen.png      (620x300)
```

## ビルド手順

### 基本ビルド（署名なし）

```powershell
cd build\windows\msix
.\build-msix.ps1 -Version 1.0.0
```

### ビルド + 自己署名（ローカルテスト用）

```powershell
.\build-msix.ps1 -Version 1.0.0 -Sign
```

### Wails ビルドをスキップ（EXE が既にある場合）

```powershell
.\build-msix.ps1 -Version 1.0.0 -SkipBuild -Sign
```

### 既存証明書で署名

```powershell
.\build-msix.ps1 -Version 1.0.0 -Sign -CertPath "C:\path\to\cert.pfx" -CertPassword "password"
```

### Microsoft Store 提出用（Publisher CN を指定、署名不要）

```powershell
.\build-msix.ps1 -Version 1.0.0 -Publisher "CN=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
```

## コード署名

### A. 自己署名証明書（ローカルテスト用）

`-Sign` を付けるだけで自動作成される。手動で作る場合:

```powershell
# 1. 証明書作成
$cert = New-SelfSignedCertificate `
    -Type Custom `
    -Subject "CN=shogo-tsunoda" `
    -KeyUsage DigitalSignature `
    -FriendlyName "FileBridge Dev" `
    -CertStoreLocation "Cert:\CurrentUser\My" `
    -TextExtension @("2.5.29.37={text}1.3.6.1.5.5.7.3.3", "2.5.29.19={text}")

# 2. PFX エクスポート
$pw = ConvertTo-SecureString -String "FileBridge" -Force -AsPlainText
Export-PfxCertificate -Cert "Cert:\CurrentUser\My\$($cert.Thumbprint)" `
    -FilePath .\FileBridge-dev.pfx -Password $pw

# 3. 信頼ストアにインストール (管理者権限)
certutil -addstore TrustedPeople .\FileBridge-dev.pfx
```

### B. Microsoft Store 提出用

Store 提出時は署名不要。Partner Center がアップロード時に Microsoft の証明書で自動署名する。`-Sign` を付けずにビルドし、生成された `.msix` をそのままアップロードする。

### C. Store 外配布用（本番証明書）

EV コード署名証明書を使用する場合:

```powershell
signtool sign /fd SHA256 /a /f "cert.pfx" /p "password" `
    /tr http://timestamp.digicert.com /td SHA256 `
    .\build\bin\FileBridge.msix
```

## Microsoft Store 提出手順

1. [Partner Center](https://partner.microsoft.com) でアプリを登録
2. 「製品」→「アプリ ID」から以下を取得:
   - `Name` (例: `12345ShogoTsunoda.FileBridge`)
   - `Publisher` (例: `CN=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX`)
3. 取得した Publisher を指定してビルド:
   ```powershell
   .\build-msix.ps1 -Version 1.0.0 -Publisher "CN=取得したCN"
   ```
4. `build\bin\FileBridge.msix` を Partner Center にアップロード
5. Store のスクリーンショット、説明文、プライバシーポリシー URL を設定
6. 認定申請を提出

## AppxManifest.xml について

テンプレート (`build/windows/msix/AppxManifest.xml`) 内のプレースホルダ:

- `{{VERSION}}` → `-Version` パラメータで置換される（4桁: `X.Y.Z.W`）
- `Publisher="CN=shogo-tsunoda"` → `-Publisher` パラメータで置換される

### Capabilities

| Capability | 理由 |
|---|---|
| `internetClientServer` | Wi-Fi 経由でファイル転送を受信するため |
| `privateNetworkClientServer` | LAN 内の通信を許可するため |
| `runFullTrust` | デスクトップアプリ (Win32) として動作するため |

## Wails v2 固有の注意点

- Wails v2 の `FileBridge.exe` は単一バイナリ（WebView2 使用）で、追加 DLL は不要
- WebView2 Runtime は Windows 10/11 に標準搭載されており依存関係の問題はない
- MinVersion `10.0.17763.0` (Windows 10 1809) 以上が対象

## トラブルシュート

| 問題 | 原因 | 解決策 |
|---|---|---|
| `makeappx.exe not found` | Windows SDK 未インストール | `winget install Microsoft.WindowsSDK.10.0.26100` |
| `magick not found` | ImageMagick 未インストール | `winget install ImageMagick.ImageMagick` → ターミナル再起動 |
| `Wails build failed` | Go/Node/Wails CLI の問題 | `go version`, `node -v`, `wails doctor` で確認 |
| MSIX インストール時「信頼されていない」 | 自己署名証明書が未登録 | `certutil -addstore TrustedPeople .\FileBridge-dev.pfx` (管理者) |
| Store で「Publisher が一致しない」 | マニフェストの Publisher CN が Partner Center と不一致 | `-Publisher` に Partner Center の正確な CN を指定 |
| `signtool: SignerSign() error` | 証明書の Subject とマニフェストの Publisher が不一致 | 両方を完全に同じ CN にする |
| アセット画像が Store で拒否 | サイズが規格外 | ImageMagick で再生成、背景色 `#0F172A` を確認 |
| `0x80080204` パッケージ検証エラー | MinVersion が高すぎる / Capability 不正 | `MinVersion="10.0.17763.0"` を確認 |
