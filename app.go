package main

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	config     *Config
	fileServer *FileServer
	history    []UploadRecord
	historyMu  sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	cfg := LoadConfig()
	app := &App{
		config:  cfg,
		history: make([]UploadRecord, 0),
	}
	app.fileServer = NewFileServer(app)
	return app
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Ensure save directory exists
	os.MkdirAll(a.config.SaveDir, 0755)

	// Start the HTTP server
	if err := a.fileServer.Start(); err != nil {
		log.Printf("Failed to start HTTP server: %v", err)
	}
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.fileServer != nil {
		a.fileServer.Stop()
	}
}

// GetServerInfo returns server information for the frontend
func (a *App) GetServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"running":   a.fileServer.IsRunning(),
		"url":       a.fileServer.GetUploadURL(),
		"port":      a.fileServer.GetPort(),
		"lanIP":     a.fileServer.GetLANIP(),
		"saveDir":   a.config.SaveDir,
		"lang":      a.config.Lang,
	}
}

// GetSaveDir returns the current save directory
func (a *App) GetSaveDir() string {
	return a.config.SaveDir
}

// GetLang returns the current language setting
func (a *App) GetLang() string {
	return a.config.Lang
}

// SetLang changes the language and saves config
func (a *App) SetLang(lang string) error {
	if lang != "ja" && lang != "en" {
		lang = "ja"
	}
	a.config.Lang = lang
	return SaveConfig(a.config)
}

// SelectSaveDir opens a folder selection dialog
func (a *App) SelectSaveDir() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Select Save Folder",
		DefaultDirectory: a.config.SaveDir,
	})
	if err != nil {
		return "", err
	}
	if dir == "" {
		// User cancelled
		return a.config.SaveDir, nil
	}

	a.config.SaveDir = dir
	if err := SaveConfig(a.config); err != nil {
		log.Printf("Failed to save config: %v", err)
	}

	return dir, nil
}

// GetCompressSettings returns the current compression settings
func (a *App) GetCompressSettings() map[string]interface{} {
	return map[string]interface{}{
		"compressImages": a.config.CompressImages,
		"imageQuality":   a.config.ImageQuality,
		"keepOriginal":   a.config.KeepOriginal,
	}
}

// SetCompressSettings updates the compression settings and saves config
func (a *App) SetCompressSettings(compressImages bool, imageQuality int, keepOriginal bool) error {
	if imageQuality < 30 {
		imageQuality = 30
	}
	if imageQuality > 95 {
		imageQuality = 95
	}
	a.config.CompressImages = compressImages
	a.config.ImageQuality = imageQuality
	a.config.KeepOriginal = keepOriginal
	return SaveConfig(a.config)
}

// GetUploadHistory returns the recent upload history
func (a *App) GetUploadHistory() []UploadRecord {
	a.historyMu.Lock()
	defer a.historyMu.Unlock()

	// Return a copy
	result := make([]UploadRecord, len(a.history))
	copy(result, a.history)
	return result
}

// addUploadRecord adds a record to the upload history (max 10)
func (a *App) addUploadRecord(record UploadRecord) {
	a.historyMu.Lock()
	defer a.historyMu.Unlock()

	// Prepend
	a.history = append([]UploadRecord{record}, a.history...)

	// Keep max 10
	if len(a.history) > 10 {
		a.history = a.history[:10]
	}

	// Emit event to frontend
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "upload:completed", record)
	}
}
