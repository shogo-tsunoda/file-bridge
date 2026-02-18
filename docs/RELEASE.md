# Release 手順（Windows）

## 事前チェック

- [ ] `go test ./...` が通る
- [ ] `frontend` の `npm run build` が通る
- [ ] `wails build` で `build/bin/file-bridge.exe` が生成される
- [ ] （任意）NSIS導入済みなら `wails build --nsis` で installer を生成

## バージョンタグ

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

## リリースアセット作成

PowerShell:

```powershell
./scripts/release.ps1 -Version X.Y.Z
```

出力:
- `build/release/FileBridge-vX.Y.Z-windows-amd64.zip`
- `build/release/SHA256SUMS.txt`

※ `file-bridge-amd64-installer.exe` は NSIS が入っている場合のみ `build/bin/` に出ます。

## GitHub Releases

1. GitHub の Releases で `vX.Y.Z` の新規リリースを作成
2. 以下をアップロード
   - `FileBridge-vX.Y.Z-windows-amd64.zip`
   - （ある場合）`file-bridge-amd64-installer.exe`
   - `SHA256SUMS.txt`
3. Release notes を記載して公開
