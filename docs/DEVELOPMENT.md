# 開発ガイド

## 前提条件

| ツール | バージョン | インストール |
|--------|-----------|-------------|
| Go | 1.22+ | https://go.dev/dl/ |
| Node.js | 18+ | https://nodejs.org/ |
| Wails CLI | v2 | 下記参照 |

```bash
# Wails CLI のインストール
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# インストール確認
wails doctor
```

---

## プロジェクト構成

```
file-bridge/
├── main.go                 # エントリポイント（Wails起動、ウィンドウ設定）
├── app.go                  # App構造体（Wails binding: サーバ情報、フォルダ選択、履歴、言語設定）
├── server.go               # HTTPサーバ管理（LAN IP検出、ポート自動検出、起動/停止）
├── upload_handler.go       # アップロード処理（GET /upload、POST /api/upload、ファイル名サニタイズ）
├── config.go               # 設定の読み書き（%APPDATA%\FileBridge\config.json）
├── wails.json              # Wailsプロジェクト設定
├── go.mod / go.sum         # Goモジュール
├── build/
│   ├── appicon.png         # アプリアイコン（1024x1024 PNG）
│   └── windows/
│       ├── icon.ico        # appicon.png から自動生成
│       └── installer/      # NSIS インストーラー設定
├── frontend/
│   ├── src/
│   │   ├── App.tsx         # メインUI（QR表示、保存先、履歴、警告）
│   │   ├── App.css         # スタイル
│   │   ├── i18n.ts         # 多言語対応（ja / en）
│   │   ├── main.tsx        # Reactエントリポイント
│   │   └── style.css       # グローバルスタイル
│   ├── wailsjs/            # Wails自動生成バインディング（編集不要）
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── docs/                   # ドキュメント
├── LICENSE
└── README.md
```

---

## 開発コマンド

### 開発モード（ホットリロード）

```bash
wails dev
```

- フロントエンド（React）の変更はホットリロードで即座に反映
- Go側の変更は自動で再コンパイル・再起動
- 開発用ブラウザ: http://localhost:34115 でもアクセス可能

### プロダクションビルド

```bash
wails build
```

成果物: `build/bin/file-bridge.exe`

### NSIS インストーラー付きビルド

```bash
wails build --nsis
```

前提: [NSIS](https://nsis.sourceforge.io/Download) がインストール済みであること。

成果物:
- `build/bin/file-bridge.exe`
- `build/bin/file-bridge-amd64-installer.exe`

### Wails バインディングの再生成

Go側のメソッドを追加・変更した場合:

```bash
wails generate module
```

`frontend/wailsjs/go/main/App.js` と `App.d.ts` が再生成されます。

---

## アーキテクチャ

### 全体構成

```
┌─────────────────────────────────┐
│  Wails Desktop App (Windows)    │
│                                 │
│  ┌──────────┐  ┌─────────────┐  │
│  │  Go側    │  │  React側    │  │
│  │          │◄─┤             │  │
│  │ App      │  │ App.tsx     │  │
│  │ (binding)│─►│ (QR/履歴等) │  │
│  └────┬─────┘  └─────────────┘  │
│       │                         │
│  ┌────▼─────────────────────┐   │
│  │  HTTP Server (net/http)  │   │
│  │  0.0.0.0:<自動ポート>     │   │
│  │                          │   │
│  │  GET /upload → HTML配信  │   │
│  │  POST /api/upload → 保存 │   │
│  └──────────────────────────┘   │
└─────────────────────────────────┘
        ▲
        │ 同一Wi-Fi
        ▼
┌───────────────┐
│  iPhone Safari │
│  (アップロード) │
└───────────────┘
```

### データフロー

1. アプリ起動 → `app.startup()` → HTTPサーバ起動（空きポート自動検出）
2. React UI が `GetServerInfo()` を呼び出し → URL + QR表示
3. iPhone が `GET /upload` → スマホ用HTML取得
4. iPhone が `POST /api/upload` → Go側でストリーミング保存
5. 保存完了 → `EventsEmit("upload:completed")` → React側の履歴が自動更新
6. アプリ終了 → `app.shutdown()` → HTTPサーバ graceful shutdown

### 設定ファイル

パス: `%APPDATA%\FileBridge\config.json`

```json
{
  "saveDir": "C:\\Users\\<user>\\Downloads\\FileBridge",
  "lang": "ja"
}
```

---

## 多言語対応（i18n）

### 対象箇所

| 箇所 | 仕組み |
|------|--------|
| React UI | `frontend/src/i18n.ts` の翻訳オブジェクト + `t("key")` 関数 |
| モバイルページ | `upload_handler.go` の `uploadTranslations` map + Go `html/template` |

### 言語の追加方法

1. `frontend/src/i18n.ts` の `translations` に新言語のキーを追加
2. `upload_handler.go` の `uploadTranslations` に同じ言語キーを追加
3. `app.go` の `SetLang()` で新言語を許可
4. `frontend/src/i18n.ts` の `Lang` 型に追加
5. `App.tsx` の言語切替ボタンを追加

---

## セキュリティ対策

| 対策 | 実装箇所 |
|------|---------|
| Path traversal防止 | `sanitizeFilename()` で `filepath.Base()` + 危険文字除去 |
| ファイル名サニタイズ | `<>:"/\|?*` や制御文字を `_` に置換 |
| ファイル名衝突 | `resolveUniquePath()` で `name (1).ext` 形式にリネーム |
| アップロードサイズ制限 | `http.MaxBytesReader` で 2GB上限 |
| ストリーミング保存 | `io.Copy` でメモリに全載せしない |

---

## トラブルシューティング（開発時）

### `wails dev` が起動しない

```bash
# 依存関係の確認
wails doctor

# フロントエンドの依存を再インストール
cd frontend && npm install && cd ..
```

### バインディングが見つからない

Go側のメソッドを追加/変更した後は:

```bash
wails generate module
```

### ポートが使えない

前回のプロセスが残っている場合:

```bash
# 使用中のポートを確認（PowerShell）
netstat -ano | findstr LISTENING | findstr <ポート番号>

# プロセスを終了
taskkill /PID <PID> /F
```

### アイコンが反映されない

`build/windows/icon.ico` を削除してからリビルド:

```bash
rm build/windows/icon.ico
wails build
```
