package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"path/filepath"
	"strings"

	"golang.org/x/image/webp"
)

// CompressResult holds the result of an image compression attempt
type CompressResult struct {
	Data         []byte
	Extension    string
	DidCompress  bool
	OriginalSize int64
	NewSize      int64
}

// compressibleExts lists extensions that can be compressed
var compressibleExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// IsCompressibleImage checks if a filename has a compressible image extension
func IsCompressibleImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return compressibleExts[ext]
}

// CompressImage compresses image data based on the file extension and quality setting.
// Returns the compressed data or an error. On error, caller should fall back to original.
func CompressImage(data []byte, filename string, quality int) (*CompressResult, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	if !compressibleExts[ext] {
		return &CompressResult{Data: data, Extension: ext, DidCompress: false, OriginalSize: int64(len(data))}, nil
	}

	if quality < 1 || quality > 100 {
		quality = 80
	}

	reader := bytes.NewReader(data)
	originalSize := int64(len(data))

	var img image.Image
	var err error

	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(reader)
	case ".png":
		img, err = png.Decode(reader)
	case ".webp":
		img, err = webp.Decode(reader)
	default:
		return &CompressResult{Data: data, Extension: ext, DidCompress: false, OriginalSize: originalSize}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode %s image: %w", ext, err)
	}

	var buf bytes.Buffer
	var outExt string

	switch ext {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		outExt = ".jpg"
	case ".webp":
		// WebP decode is supported but Go stdlib can't encode WebP,
		// so re-encode as JPEG with the specified quality
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		outExt = ".jpg"
	case ".png":
		// PNG compression: use BestCompression encoder
		encoder := &png.Encoder{CompressionLevel: png.BestCompression}
		err = encoder.Encode(&buf, img)
		outExt = ".png"
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	newSize := int64(buf.Len())

	// If compressed is larger, keep original
	if newSize >= originalSize && ext != ".webp" {
		log.Printf("Compression did not reduce size for %s (%d >= %d), keeping original", filename, newSize, originalSize)
		return &CompressResult{Data: data, Extension: ext, DidCompress: false, OriginalSize: originalSize, NewSize: originalSize}, nil
	}

	return &CompressResult{
		Data:         buf.Bytes(),
		Extension:    outExt,
		DidCompress:  true,
		OriginalSize: originalSize,
		NewSize:      newSize,
	}, nil
}

// CompressImageFromReader reads all data from reader and compresses it.
// This is a convenience wrapper for the upload handler.
func CompressImageFromReader(r io.Reader, filename string, quality int) ([]byte, *CompressResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read image data: %w", err)
	}

	result, err := CompressImage(data, filename, quality)
	if err != nil {
		// Return original data so caller can fall back
		return data, nil, err
	}

	return data, result, nil
}
