<div align="center">

# File Bridge

**The simple AirDrop alternative for Windows.**
Transfer files from iPhone to PC over local Wi-Fi — no cloud, no account, no cables.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%2010%2F11-0078D6)](https://github.com/shogo-tsunoda/file-bridge/releases)
[![Release](https://img.shields.io/github/v/release/shogo-tsunoda/file-bridge?label=Download)](https://github.com/shogo-tsunoda/file-bridge/releases/latest)

[**Download**](https://github.com/shogo-tsunoda/file-bridge/releases/latest) | [How it works](#how-it-works) | [FAQ](#faq)

</div>

---

## What is File Bridge?

File Bridge is a free, open-source Windows desktop app that lets you **transfer files from your iPhone to your Windows PC** using only your local Wi-Fi network.

Just open the app, scan a QR code with your iPhone, and send files instantly. No cloud service, no sign-up, no iTunes, no USB cable. Your files stay on your local network and are never uploaded to any external server.

It works as a **local network file transfer** tool — a lightweight, privacy-first alternative to AirDrop for Windows users.

---

## file-bridge とは？

iPhone から Windows PC にファイルを転送するための無料デスクトップアプリです。

- アプリを起動して QR コードを読み取るだけ
- 同じ Wi-Fi につながっていれば OK
- クラウドなし、アカウント登録なし、ケーブル不要
- データは外部サーバーに一切送信されません

---

## Why File Bridge?

- **No cloud** — Files transfer directly over your local network
- **No account** — No sign-up, no login, no email required
- **No cables** — Wireless transfer over Wi-Fi
- **Private** — Your data never leaves your network
- **Simple** — Scan a QR code and you're connected
- **Free & open source** — MIT licensed, no ads, no tracking

---

## How it works

```mermaid
graph LR
    A["iPhone (Safari)"] -- "Wi-Fi (LAN)" --> B["File Bridge (Windows)"]
    B --> C["Save Folder"]
```

1. **Launch** File Bridge on your Windows PC
2. **Scan** the QR code shown on screen with your iPhone
3. **Select** files on your iPhone and tap Upload
4. **Done** — files are saved to your chosen folder

> Both devices must be connected to the same Wi-Fi network. No internet connection is required.

---

## Screenshots

<!-- Replace with actual screenshots -->

| Windows App | iPhone Upload |
|:-----------:|:------------:|
| ![Windows App](docs/screenshots/app.png) | ![iPhone Upload](docs/screenshots/mobile.png) |

---

## Quick Start

### 1. Download

Go to the [Releases page](https://github.com/shogo-tsunoda/file-bridge/releases/latest) and download:

- **`FileBridge-vX.X.X-windows-amd64.zip`** — Portable version (just unzip and run)
- **`FileBridge-amd64-installer.exe`** — Installer version

### 2. Run

Launch `FileBridge.exe`. A QR code and URL will appear.

> On first launch, Windows Firewall may ask to allow the app. Click **Allow** so your iPhone can connect.

### 3. Scan

Open your iPhone camera and scan the QR code. Safari will open the upload page.

### 4. Upload

Tap **Select Files**, choose your photos/videos/documents, then tap **Upload**.

### 5. Done

Files are saved to the folder shown in the app. You can change the save location anytime.

---

## FAQ

### Can I transfer files from iPhone to Windows without iTunes?

Yes. File Bridge transfers files directly over your local Wi-Fi. No iTunes, no iCloud, no USB cable needed.

### Does this use the internet?

No. File Bridge works entirely on your local network. No data is sent to the internet.

### Is my data uploaded to any server?

No. Files are sent directly from your iPhone to your PC. There is no external server involved.

### Does it work without Wi-Fi?

Both your iPhone and PC need to be on the same Wi-Fi network. File Bridge does not use Bluetooth or mobile data.

### Is it safe?

File Bridge runs a local HTTP server only accessible within your Wi-Fi network. No data leaves your local network. The source code is fully open and auditable.

### What file types are supported?

All file types — photos, videos, PDFs, documents, and any other files. Maximum file size is 2 GB.

### Does it support multiple languages?

Yes. File Bridge supports English and Japanese. You can switch languages in the app.

---

## Security & Privacy

File Bridge is designed with privacy in mind:

- **Local only** — All transfers happen over your local Wi-Fi network
- **No external connections** — The app never contacts any remote server
- **No telemetry** — No usage data, analytics, or tracking of any kind
- **No account** — No personal information is collected
- **Open source** — The complete source code is available for review

---

## Contributing

Contributions are welcome! See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for build instructions.

Bug reports: [GitHub Issues](https://github.com/shogo-tsunoda/file-bridge/issues)

---

## License

[MIT License](LICENSE) — free for personal and commercial use.
