# File Bridge

iPhoneからWindows PCへ、同一Wi-Fi経由でファイルを転送するデスクトップアプリです。

## 機能

- Windowsアプリを起動するとローカルHTTPサーバが自動起動
- 画面にアップロードURLのQRコードを表示
- iPhoneでQRを読み取り → Safariでアップロード画面を開く
- 複数ファイルを選択して送信 → Windows上の指定フォルダに保存
- 画像・動画・PDF・任意のファイルに対応（最大2GB）
- ファイル名衝突時は自動リネーム（例: `name (1).ext`）
- 受信履歴の表示（直近10件）

## 技術スタック

- **Backend**: Go + Wails v2
- **Frontend**: React + TypeScript + Vite
- **QRコード**: qrcode.react

## セットアップ

### 前提条件

- [Go](https://go.dev/dl/) 1.22以上
- [Node.js](https://nodejs.org/) 18以上
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 開発

```bash
# 開発モードで起動（ホットリロード対応）
wails dev

# プロダクションビルド
wails build
```

ビルド成果物は `build/bin/file-bridge.exe` に生成されます。

## 使い方

1. `file-bridge.exe` を起動する
2. 画面にQRコードとURLが表示される
3. iPhoneで QRコードを読み取る（またはURLをSafariで開く）
4. 「Select Files」をタップしてファイルを選択
5. 「Upload」をタップして送信
6. Windows側の指定フォルダにファイルが保存される

## 保存先の変更

アプリ画面の「Change」ボタンからフォルダを選択できます。
設定は `%APPDATA%\FileBridge\config.json` に保存されます。

デフォルトの保存先: `%USERPROFILE%\Downloads\FileBridge`

## トラブルシューティング

### iPhoneから接続できない

1. **同一Wi-Fiか確認**: PCとiPhoneが同じWi-Fiネットワークに接続されていること
2. **Windowsファイアウォール**: アプリがファイアウォールでブロックされている場合があります
   - Windows設定 → プライバシーとセキュリティ → Windowsセキュリティ → ファイアウォールとネットワーク保護
   - 「ファイアウォールによるアプリケーションの許可」から `file-bridge.exe` を許可
3. **ポートの確認**: アプリ画面に表示されているポート番号が正しいか確認
4. **VPN**: VPNが有効な場合、ローカル通信がブロックされることがあります。一時的にOFFにしてください

### QRコードが表示されない

- Wi-Fiに接続されていない場合、LAN IPが取得できずQRが表示されません
- PCがイーサネットのみの場合でもIPが取得できれば動作します

### ファイルが大きすぎる

- 最大2GBまで対応しています
- ネットワーク速度によっては時間がかかる場合があります
