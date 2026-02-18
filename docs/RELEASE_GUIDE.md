# File Bridge リリースガイド

フリーソフトとして配布するまでの全手順をまとめたドキュメントです。

---

## 目次

1. [リリース前の準備](#1-リリース前の準備)
2. [バージョン管理](#2-バージョン管理)
3. [ビルド手順](#3-ビルド手順)
4. [配布パッケージの作成](#4-配布パッケージの作成)
5. [GitHub Releases で公開](#5-github-releases-で公開)
6. [GitHub Actions で自動化（任意）](#6-github-actions-で自動化任意)
7. [リリース後の対応](#7-リリース後の対応)
8. [チェックリスト](#8-チェックリスト)

---

## 1. リリース前の準備

### 1.1 ライセンスファイルの追加

フリーソフトとして配布するため、ライセンスを明記します。
プロジェクトルートに `LICENSE` ファイルを作成してください。

**推奨: MIT License**（制約が少なく、フリーソフトに適している）

```
MIT License

Copyright (c) 2026 shogo-tsunoda

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

### 1.2 wails.json のバージョン情報を設定

`wails.json` に以下を追加します。

```json
{
  "info": {
    "companyName": "shogo-tsunoda",
    "productName": "File Bridge",
    "productVersion": "1.0.0",
    "copyright": "Copyright (c) 2026 shogo-tsunoda",
    "comments": "iPhone to Windows file transfer over Wi-Fi"
  }
}
```

### 1.3 アプリアイコンの準備

`build/appicon.png` を差し替えると、ビルド時にexeのアイコンとして反映されます。

- **サイズ**: 1024x1024px の PNG を推奨
- ビルド時に自動で `build/windows/icon.ico` に変換されます
- アイコンがなければデフォルトのWailsアイコンが使われます

### 1.4 .gitignore の確認

以下が含まれていることを確認します。

```
build/bin
node_modules
frontend/dist
```

---

## 2. バージョン管理

### セマンティックバージョニング

`MAJOR.MINOR.PATCH` の形式で管理します。

| 種類 | 例 | 変更内容 |
|------|------|----------|
| MAJOR | 2.0.0 | 破壊的変更、大幅な機能変更 |
| MINOR | 1.1.0 | 新機能追加（後方互換あり） |
| PATCH | 1.0.1 | バグ修正 |

### バージョンの更新箇所

リリースのたびに以下を更新します。

1. **`wails.json`** → `info.productVersion`
2. **Git タグ** → `v1.0.0` 形式

---

## 3. ビルド手順

### 3.1 前提条件

| ツール | バージョン | インストール |
|--------|-----------|-------------|
| Go | 1.22+ | https://go.dev/dl/ |
| Node.js | 18+ | https://nodejs.org/ |
| Wails CLI | v2 | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| NSIS | 3.x | インストーラーを作る場合のみ（後述） |

### 3.2 通常ビルド（ポータブル版 exe）

```bash
wails build
```

成果物: `build/bin/file-bridge.exe`

このexe単体で動作します（ポータブル版として配布可能）。

### 3.3 NSIS インストーラー版ビルド

```bash
wails build --nsis
```

成果物:
- `build/bin/file-bridge.exe` （アプリ本体）
- `build/bin/file-bridge-amd64-installer.exe` （インストーラー）

**前提**: [NSIS](https://nsis.sourceforge.io/Download) がインストールされ、PATH に通っていること。

インストーラーの動作:
- Program Files にインストール
- スタートメニュー & デスクトップにショートカット作成
- WebView2ランタイムが無い場合は自動インストール
- アンインストーラー付き

---

## 4. 配布パッケージの作成

### 4.1 ポータブル版（ZIP）

一番シンプルな配布形態です。

```bash
# ビルド
wails build

# ZIP作成（PowerShellの場合）
Compress-Archive -Path build\bin\file-bridge.exe -DestinationPath build\bin\FileBridge-v1.0.0-windows-amd64.zip
```

ユーザーはZIPを解凍して exe を実行するだけで使えます。

### 4.2 インストーラー版

```bash
wails build --nsis
```

生成される `file-bridge-amd64-installer.exe` をそのまま配布します。

### 配布形態の比較

| | ポータブル版 (ZIP) | インストーラー版 (NSIS) |
|---|---|---|
| **手軽さ** | 解凍するだけ | ウィザードでインストール |
| **ショートカット** | なし（自分で作る） | 自動作成 |
| **アンインストール** | フォルダ削除 | アンインストーラーあり |
| **推奨** | 試用・持ち運び | 常用 |

**推奨**: 両方を GitHub Releases に置く。

---

## 5. GitHub Releases で公開

### 5.1 リポジトリの準備

```bash
# GitHubにリポジトリを作成済みの場合
git remote add origin https://github.com/<ユーザー名>/file-bridge.git
git push -u origin main
```

### 5.2 タグの作成とプッシュ

```bash
git tag -a v1.0.0 -m "v1.0.0 初回リリース"
git push origin v1.0.0
```

### 5.3 GitHub Releases の作成

```bash
# gh CLI を使う場合
gh release create v1.0.0 \
  build/bin/FileBridge-v1.0.0-windows-amd64.zip \
  build/bin/file-bridge-amd64-installer.exe \
  --title "File Bridge v1.0.0" \
  --notes "$(cat <<'EOF'
## File Bridge v1.0.0

iPhoneからWindows PCへ、同一Wi-Fi経由でファイルを転送するアプリです。

### ダウンロード
- **FileBridge-v1.0.0-windows-amd64.zip** - ポータブル版（解凍して実行）
- **file-bridge-amd64-installer.exe** - インストーラー版

### 主な機能
- QRコードでiPhoneから簡単にアクセス
- 画像・動画・PDF・任意のファイルに対応（最大2GB）
- 日本語 / English 対応
- ファイル名の衝突を自動リネーム
- 受信履歴の表示

### 動作要件
- Windows 10 / 11 (64-bit)
- WebView2 Runtime（通常はプリインストール済み）
- iPhoneと同一Wi-Fiネットワーク

### 注意事項
- 初回起動時にWindowsファイアウォールの確認ダイアログが表示されることがあります。「許可」を選択してください。
- Windows SmartScreen の警告が出る場合は「詳細情報」→「実行」で起動できます。
EOF
)"
```

### 5.4 gh CLI がない場合

1. https://github.com/<ユーザー名>/file-bridge/releases/new を開く
2. タグ `v1.0.0` を選択
3. リリースタイトルとノートを記入
4. ビルド成果物をドラッグ＆ドロップでアップロード
5. 「Publish release」をクリック

---

## 6. GitHub Actions で自動化（任意）

タグをプッシュしたら自動でビルド＆リリースする設定です。

ファイル: `.github/workflows/release.yml`

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  build:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install Wails CLI
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest

      - name: Install NSIS
        run: choco install nsis -y

      - name: Build (portable)
        run: wails build

      - name: Build (NSIS installer)
        run: wails build --nsis

      - name: Create ZIP
        run: |
          Compress-Archive -Path build\bin\file-bridge.exe -DestinationPath build\bin\FileBridge-${{ github.ref_name }}-windows-amd64.zip

      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            build/bin/FileBridge-${{ github.ref_name }}-windows-amd64.zip
            build/bin/file-bridge-amd64-installer.exe
          generate_release_notes: true
```

### 使い方

```bash
# タグをプッシュするだけで自動リリース
git tag -a v1.0.0 -m "v1.0.0"
git push origin v1.0.0
```

---

## 7. リリース後の対応

### 7.1 Windows SmartScreen 警告について

コード署名していないexeは、初回実行時にSmartScreenの警告が出ます。

**ユーザーへの案内文（READMEやリリースノートに記載）**:

> 初回起動時に「WindowsによってPCが保護されました」と表示される場合があります。
> 「詳細情報」をクリック → 「実行」をクリックすると起動できます。
> これはコード署名がないためで、ウイルスではありません。

### 7.2 コード署名（任意・有料）

SmartScreen警告を回避するにはコード署名証明書が必要です。

| 種類 | 費用目安 | 説明 |
|------|---------|------|
| OV証明書 | 年 $200〜400 | 組織認証。SmartScreen警告が減る |
| EV証明書 | 年 $300〜600 | 拡張認証。SmartScreen即時信頼 |

フリーソフトの初期段階では不要です。ダウンロード数が増えるとSmartScreenの「評判」が蓄積され、警告が出にくくなります。

### 7.3 バグ報告の受付

GitHub Issues を有効にしておきます。
リリースノートやREADMEに以下を記載:

```
バグ報告・要望: https://github.com/<ユーザー名>/file-bridge/issues
```

---

## 8. チェックリスト

### 初回リリース前

- [ ] `LICENSE` ファイルをリポジトリルートに追加
- [ ] `wails.json` に `info`（バージョン、著作権など）を記入
- [ ] アプリアイコン (`build/appicon.png`) を差し替え（任意）
- [ ] `README.md` にダウンロードリンク、使い方、注意事項を記載
- [ ] ローカルでビルド確認 (`wails build`)
- [ ] 実際にexeを起動し、全機能を手動テスト
  - [ ] QRコード表示 → iPhoneでアクセスできる
  - [ ] ファイルアップロード（画像・動画・PDF）
  - [ ] 大きめのファイル（200MB程度）
  - [ ] ファイル名衝突 → リネームされる
  - [ ] 保存先フォルダの変更
  - [ ] 言語切替（JA / EN）
  - [ ] アプリ再起動後、設定が保持されている
- [ ] GitHub リポジトリを作成・プッシュ
- [ ] タグ作成 (`v1.0.0`)
- [ ] GitHub Release を作成し、ビルド成果物をアップロード

### 各リリース時

- [ ] `wails.json` の `productVersion` を更新
- [ ] 変更内容をコミット
- [ ] タグ作成 (`vX.Y.Z`)
- [ ] ビルド & テスト
- [ ] GitHub Release 作成（または CI で自動）
