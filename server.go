package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)

// FileServer manages the HTTP server for file uploads
type FileServer struct {
	server   *http.Server
	port     int
	lanIP    string
	running  bool
	app      *App
}

// NewFileServer creates a new FileServer instance
func NewFileServer(app *App) *FileServer {
	return &FileServer{app: app}
}

// Start starts the HTTP server on an available port
func (fs *FileServer) Start() error {
	if fs.running {
		return fmt.Errorf("server is already running")
	}

	// Find LAN IP
	ip, err := getLANIP()
	if err != nil {
		return fmt.Errorf("failed to get LAN IP: %w", err)
	}
	fs.lanIP = ip

	// Find available port
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}
	fs.port = listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", fs.handleUploadPage)
	mux.HandleFunc("/api/upload", fs.handleFileUpload)

	fs.server = &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", fs.port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		// No ReadTimeout to allow large file uploads
		WriteTimeout:   0,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	go func() {
		log.Printf("HTTP server starting on 0.0.0.0:%d", fs.port)
		if err := fs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	fs.running = true
	log.Printf("Upload URL: http://%s:%d/upload", fs.lanIP, fs.port)
	return nil
}

// Stop gracefully stops the HTTP server
func (fs *FileServer) Stop() error {
	if !fs.running || fs.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := fs.server.Shutdown(ctx)
	fs.running = false
	log.Println("HTTP server stopped")
	return err
}

// GetUploadURL returns the upload URL with language parameter
func (fs *FileServer) GetUploadURL() string {
	if !fs.running {
		return ""
	}
	lang := fs.app.GetLang()
	return fmt.Sprintf("http://%s:%d/upload?lang=%s", fs.lanIP, fs.port, lang)
}

// GetPort returns the server port
func (fs *FileServer) GetPort() int {
	return fs.port
}

// GetLANIP returns the LAN IP address
func (fs *FileServer) GetLANIP() string {
	return fs.lanIP
}

// IsRunning returns whether the server is running
func (fs *FileServer) IsRunning() bool {
	return fs.running
}

// getLANIP finds the best LAN IP address
func getLANIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	type ipCandidate struct {
		ip       string
		priority int
	}

	var candidates []ipCandidate

	for _, iface := range ifaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Only IPv4
			if ip == nil || ip.To4() == nil {
				continue
			}

			ipStr := ip.String()
			priority := 0

			// Prioritize private IP ranges
			if strings.HasPrefix(ipStr, "192.168.") {
				priority = 3
			} else if strings.HasPrefix(ipStr, "10.") {
				priority = 2
			} else if strings.HasPrefix(ipStr, "172.") {
				// Check 172.16.0.0 - 172.31.255.255
				parts := strings.Split(ipStr, ".")
				if len(parts) >= 2 {
					secondOctet := 0
					fmt.Sscanf(parts[1], "%d", &secondOctet)
					if secondOctet >= 16 && secondOctet <= 31 {
						priority = 1
					}
				}
			}

			if priority > 0 {
				// Prefer Wi-Fi interfaces
				nameLower := strings.ToLower(iface.Name)
				if strings.Contains(nameLower, "wi-fi") || strings.Contains(nameLower, "wifi") || strings.Contains(nameLower, "wlan") || strings.Contains(nameLower, "wireless") {
					priority += 10
				}
				candidates = append(candidates, ipCandidate{ip: ipStr, priority: priority})
			}
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no LAN IP address found")
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].priority > candidates[j].priority
	})

	return candidates[0].ip, nil
}
