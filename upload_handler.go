package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// UploadRecord represents a single file upload record
type UploadRecord struct {
	FileName     string `json:"fileName"`
	Size         int64  `json:"size"`
	Timestamp    string `json:"timestamp"`
	SavePath     string `json:"savePath"`
	Compressed   bool   `json:"compressed,omitempty"`
	OriginalSize int64  `json:"originalSize,omitempty"`
}

// maxUploadSize is the max upload size (2GB)
const maxUploadSize = 2 << 30

// uploadTexts holds translations for the mobile upload page
type uploadTexts struct {
	PageTitle     string
	Heading       string
	SelectFiles   string
	FileHint      string
	UploadBtn     string
	Uploading     string
	UploadingPct  string
	SuccessSuffix string
	UploadFailed  string
	NetworkError  string
	Cancelled     string
}

var uploadTranslations = map[string]uploadTexts{
	"ja": {
		PageTitle:     "File Bridge - アップロード",
		Heading:       "File Bridge",
		SelectFiles:   "ファイルを選択",
		FileHint:      "画像・動画・PDF など",
		UploadBtn:     "アップロード",
		Uploading:     "アップロード中...",
		UploadingPct:  "アップロード中... ",
		SuccessSuffix: " 件のファイルをアップロードしました",
		UploadFailed:  "アップロードに失敗しました",
		NetworkError:  "ネットワークエラーです。接続を確認してください。",
		Cancelled:     "アップロードがキャンセルされました。",
	},
	"en": {
		PageTitle:     "File Bridge - Upload",
		Heading:       "File Bridge",
		SelectFiles:   "Select Files",
		FileHint:      "Images, videos, PDFs, etc.",
		UploadBtn:     "Upload",
		Uploading:     "Uploading...",
		UploadingPct:  "Uploading... ",
		SuccessSuffix: " file(s) uploaded successfully!",
		UploadFailed:  "Upload failed",
		NetworkError:  "Network error. Please check your connection.",
		Cancelled:     "Upload cancelled.",
	},
}

var uploadPageTemplate = template.Must(template.New("upload").Parse(uploadPageTmpl))

// handleUploadPage serves the mobile upload HTML page
func (fs *FileServer) handleUploadPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lang := r.URL.Query().Get("lang")
	texts, ok := uploadTranslations[lang]
	if !ok {
		texts = uploadTranslations["ja"]
	}

	var buf bytes.Buffer
	if err := uploadPageTemplate.Execute(&buf, texts); err != nil {
		log.Printf("Failed to render upload page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

// handleFileUpload handles multipart file upload
func (fs *FileServer) handleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	saveDir := fs.app.GetSaveDir()
	if saveDir == "" {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Save directory not configured",
		})
		return
	}

	// Ensure save directory exists
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		log.Printf("Failed to create save directory: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to create save directory",
		})
		return
	}

	// Parse multipart form with streaming (32MB buffer, rest goes to disk temp)
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Failed to parse upload. File may be too large (max 2GB).",
		})
		return
	}
	defer r.MultipartForm.RemoveAll()

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "No files uploaded",
		})
		return
	}

	// Get compression settings
	compressEnabled := fs.app.config.CompressImages
	imageQuality := fs.app.config.ImageQuality
	keepOriginal := fs.app.config.KeepOriginal

	var results []UploadRecord

	for _, fh := range files {
		// Sanitize filename
		safeName := sanitizeFilename(fh.Filename)

		// Open uploaded file
		src, err := fh.Open()
		if err != nil {
			log.Printf("Failed to open uploaded file %s: %v", fh.Filename, err)
			continue
		}

		// Check if we should compress this file
		shouldCompress := compressEnabled && IsCompressibleImage(safeName)

		if shouldCompress {
			// Read file into memory for compression
			originalData, compResult, compErr := CompressImageFromReader(src, safeName, imageQuality)
			src.Close()

			if compErr != nil {
				log.Printf("Compression failed for %s: %v, saving original", safeName, compErr)
				// Fallback: save original
				destPath := resolveUniquePath(saveDir, safeName)
				if writeErr := os.WriteFile(destPath, originalData, 0644); writeErr != nil {
					log.Printf("Failed to write fallback file %s: %v", destPath, writeErr)
					continue
				}
				record := UploadRecord{
					FileName:  filepath.Base(destPath),
					Size:      int64(len(originalData)),
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
					SavePath:  destPath,
				}
				results = append(results, record)
				fs.app.addUploadRecord(record)
				log.Printf("File saved (compression failed, original): %s (%d bytes)", destPath, len(originalData))
				continue
			}

			// Save original copy if requested
			if keepOriginal {
				origExt := filepath.Ext(safeName)
				origBase := strings.TrimSuffix(safeName, origExt)
				origName := origBase + "_original" + origExt
				origPath := resolveUniquePath(saveDir, origName)
				if writeErr := os.WriteFile(origPath, originalData, 0644); writeErr != nil {
					log.Printf("Failed to save original copy %s: %v", origPath, writeErr)
				} else {
					log.Printf("Original copy saved: %s (%d bytes)", origPath, len(originalData))
				}
			}

			// Determine output filename (extension may change for webp->jpg)
			outName := safeName
			if compResult.DidCompress {
				origExt := filepath.Ext(safeName)
				if strings.ToLower(origExt) != compResult.Extension {
					outName = strings.TrimSuffix(safeName, origExt) + compResult.Extension
				}
			}

			destPath := resolveUniquePath(saveDir, outName)
			dataToWrite := compResult.Data

			if writeErr := os.WriteFile(destPath, dataToWrite, 0644); writeErr != nil {
				log.Printf("Failed to write compressed file %s: %v", destPath, writeErr)
				continue
			}

			record := UploadRecord{
				FileName:     filepath.Base(destPath),
				Size:         int64(len(dataToWrite)),
				Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
				SavePath:     destPath,
				Compressed:   compResult.DidCompress,
				OriginalSize: compResult.OriginalSize,
			}
			results = append(results, record)
			fs.app.addUploadRecord(record)

			if compResult.DidCompress {
				log.Printf("File saved (compressed): %s (%d bytes → %d bytes)", destPath, compResult.OriginalSize, compResult.NewSize)
			} else {
				log.Printf("File saved (no size reduction): %s (%d bytes)", destPath, len(dataToWrite))
			}
		} else {
			// No compression: stream copy as before
			destPath := resolveUniquePath(saveDir, safeName)

			dst, err := os.Create(destPath)
			if err != nil {
				src.Close()
				log.Printf("Failed to create file %s: %v", destPath, err)
				continue
			}

			written, err := io.Copy(dst, src)
			src.Close()
			dst.Close()

			if err != nil {
				log.Printf("Failed to write file %s: %v", destPath, err)
				os.Remove(destPath)
				continue
			}

			record := UploadRecord{
				FileName:  filepath.Base(destPath),
				Size:      written,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				SavePath:  destPath,
			}
			results = append(results, record)
			fs.app.addUploadRecord(record)

			log.Printf("File saved: %s (%d bytes)", destPath, written)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"count":   len(results),
		"files":   results,
	})
}

// sanitizeFilename removes dangerous characters and path traversal attempts
func sanitizeFilename(name string) string {
	// Get only the base name (prevent path traversal)
	name = filepath.Base(name)

	// Remove null bytes
	name = strings.ReplaceAll(name, "\x00", "")

	// Replace dangerous characters
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	name = re.ReplaceAllString(name, "_")

	// Remove leading/trailing dots and spaces
	name = strings.Trim(name, ". ")

	// Ensure valid UTF-8
	if !utf8.ValidString(name) {
		name = strings.ToValidUTF8(name, "_")
	}

	// If empty after sanitization, use a default name
	if name == "" {
		name = fmt.Sprintf("upload_%s", time.Now().Format("20060102_150405"))
	}

	// Limit length (255 chars for Windows)
	if len(name) > 200 {
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		if len(base) > 200-len(ext) {
			base = base[:200-len(ext)]
		}
		name = base + ext
	}

	return name
}

// resolveUniquePath returns a unique file path, renaming if collision
func resolveUniquePath(dir, name string) string {
	destPath := filepath.Join(dir, name)

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		return destPath
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	for i := 1; i < 10000; i++ {
		newName := fmt.Sprintf("%s (%d)%s", base, i, ext)
		destPath = filepath.Join(dir, newName)
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			return destPath
		}
	}

	// Fallback: use timestamp
	newName := fmt.Sprintf("%s_%s%s", base, time.Now().Format("20060102_150405"), ext)
	return filepath.Join(dir, newName)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// uploadPageTmpl is the Go template for the mobile upload page
const uploadPageTmpl = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
<title>{{.PageTitle}}</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  background: #0f172a;
  color: #e2e8f0;
  min-height: 100vh;
  padding: 20px;
}
.container {
  max-width: 480px;
  margin: 0 auto;
}
h1 {
  font-size: 1.5rem;
  text-align: center;
  margin-bottom: 24px;
  color: #38bdf8;
}
.upload-area {
  border: 2px dashed #475569;
  border-radius: 12px;
  padding: 32px 16px;
  text-align: center;
  margin-bottom: 16px;
  transition: border-color 0.2s;
}
.upload-area.active {
  border-color: #38bdf8;
  background: rgba(56, 189, 248, 0.05);
}
.file-input-label {
  display: inline-block;
  background: #2563eb;
  color: white;
  padding: 14px 28px;
  border-radius: 8px;
  font-size: 1.1rem;
  cursor: pointer;
  margin-bottom: 12px;
  -webkit-tap-highlight-color: transparent;
}
.file-input-label:active {
  background: #1d4ed8;
}
input[type="file"] { display: none; }
.file-list {
  margin: 16px 0;
  text-align: left;
}
.file-item {
  background: #1e293b;
  padding: 10px 14px;
  border-radius: 8px;
  margin-bottom: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.9rem;
}
.file-item .name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 8px;
}
.file-item .size { color: #94a3b8; white-space: nowrap; margin-right: 8px; }
.file-item .remove-btn {
  background: none;
  border: none;
  color: #64748b;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
  -webkit-tap-highlight-color: transparent;
}
.file-item .remove-btn:active { color: #f87171; }
.send-btn {
  display: block;
  width: 100%;
  padding: 16px;
  background: #16a34a;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1.1rem;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}
.send-btn:disabled {
  background: #334155;
  color: #64748b;
}
.send-btn:active:not(:disabled) { background: #15803d; }
.progress-bar {
  width: 100%;
  height: 6px;
  background: #334155;
  border-radius: 3px;
  margin: 16px 0;
  overflow: hidden;
  display: none;
}
.progress-bar .fill {
  height: 100%;
  background: #38bdf8;
  width: 0%;
  transition: width 0.2s;
  border-radius: 3px;
}
.status {
  text-align: center;
  margin: 16px 0;
  font-size: 0.95rem;
  min-height: 1.5em;
}
.status.success { color: #4ade80; }
.status.error { color: #f87171; }
.status.uploading { color: #38bdf8; }
</style>
</head>
<body>
<div class="container">
  <h1>{{.Heading}}</h1>
  <div class="upload-area" id="uploadArea">
    <label class="file-input-label" for="fileInput">{{.SelectFiles}}</label>
    <input type="file" id="fileInput" multiple accept="*/*">
    <p style="color:#94a3b8; margin-top:8px; font-size:0.85rem;">{{.FileHint}}</p>
  </div>
  <div class="file-list" id="fileList"></div>
  <div class="progress-bar" id="progressBar"><div class="fill" id="progressFill"></div></div>
  <div class="status" id="status"></div>
  <button class="send-btn" id="sendBtn" disabled>{{.UploadBtn}}</button>
</div>

<script>
var T = {
  uploading: '{{.Uploading}}',
  uploadingPct: '{{.UploadingPct}}',
  successSuffix: '{{.SuccessSuffix}}',
  uploadFailed: '{{.UploadFailed}}',
  networkError: '{{.NetworkError}}',
  cancelled: '{{.Cancelled}}'
};

var fileInput = document.getElementById('fileInput');
var fileList = document.getElementById('fileList');
var sendBtn = document.getElementById('sendBtn');
var statusEl = document.getElementById('status');
var progressBar = document.getElementById('progressBar');
var progressFill = document.getElementById('progressFill');

var selectedFiles = [];

fileInput.addEventListener('change', function() {
  var newFiles = Array.from(this.files);
  var existingNames = {};
  selectedFiles.forEach(function(f) { existingNames[f.name + '_' + f.size] = true; });
  newFiles.forEach(function(f) {
    if (!existingNames[f.name + '_' + f.size]) {
      selectedFiles.push(f);
    }
  });
  this.value = '';
  renderFileList();
  sendBtn.disabled = selectedFiles.length === 0;
  statusEl.textContent = '';
  statusEl.className = 'status';
});

function renderFileList() {
  fileList.innerHTML = '';
  selectedFiles.forEach(function(f, i) {
    var div = document.createElement('div');
    div.className = 'file-item';
    var nameSpan = document.createElement('span');
    nameSpan.className = 'name';
    nameSpan.textContent = f.name;
    var sizeSpan = document.createElement('span');
    sizeSpan.className = 'size';
    sizeSpan.textContent = formatSize(f.size);
    var removeBtn = document.createElement('button');
    removeBtn.className = 'remove-btn';
    removeBtn.textContent = '\u00d7';
    removeBtn.setAttribute('data-index', i);
    removeBtn.addEventListener('click', function() {
      selectedFiles.splice(parseInt(this.getAttribute('data-index')), 1);
      renderFileList();
      sendBtn.disabled = selectedFiles.length === 0;
    });
    div.appendChild(nameSpan);
    div.appendChild(sizeSpan);
    div.appendChild(removeBtn);
    fileList.appendChild(div);
  });
}

sendBtn.addEventListener('click', function() {
  if (selectedFiles.length === 0) return;

  var formData = new FormData();
  selectedFiles.forEach(function(f) { formData.append('files', f); });

  var xhr = new XMLHttpRequest();
  xhr.open('POST', '/api/upload');

  xhr.upload.addEventListener('progress', function(e) {
    if (e.lengthComputable) {
      var pct = Math.round((e.loaded / e.total) * 100);
      progressFill.style.width = pct + '%';
      statusEl.textContent = T.uploadingPct + pct + '%';
      statusEl.className = 'status uploading';
    }
  });

  xhr.addEventListener('load', function() {
    progressBar.style.display = 'none';
    if (xhr.status === 200) {
      var res = JSON.parse(xhr.responseText);
      statusEl.textContent = res.count + T.successSuffix;
      statusEl.className = 'status success';
      selectedFiles = [];
      fileList.innerHTML = '';
      fileInput.value = '';
      sendBtn.disabled = true;
    } else {
      var msg = T.uploadFailed;
      try { msg = JSON.parse(xhr.responseText).error || msg; } catch(e) {}
      statusEl.textContent = msg;
      statusEl.className = 'status error';
    }
  });

  xhr.addEventListener('error', function() {
    progressBar.style.display = 'none';
    statusEl.textContent = T.networkError;
    statusEl.className = 'status error';
  });

  xhr.addEventListener('abort', function() {
    progressBar.style.display = 'none';
    statusEl.textContent = T.cancelled;
    statusEl.className = 'status error';
  });

  progressBar.style.display = 'block';
  progressFill.style.width = '0%';
  statusEl.textContent = T.uploading;
  statusEl.className = 'status uploading';
  sendBtn.disabled = true;

  xhr.send(formData);
});

function formatSize(bytes) {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
}

function escapeHtml(s) {
  var div = document.createElement('div');
  div.textContent = s;
  return div.innerHTML;
}
</script>
</body>
</html>`
