package bootstrap

import (
	"log"
	"path/filepath"

	"furryjan/internal/config"
	"furryjan/internal/db"
	"furryjan/internal/downloader/blob"
)

func ParseBlobData(cfg *config.Config, database *db.DB) bool {
	if !cfg.BlobWriterEnabled {
		log.Printf("Blob writer disabled in settings, using direct filesystem mode")
		return false
	}

	blobOut := filepath.Join(cfg.DownloadDir, "data.blob")
	blobIndex := filepath.Join(cfg.DownloadDir, "data.index.json")
	bufSizeBytes := cfg.BlobBufferMB * 1024 * 1024
	log.Printf("[Startup] Starting blob writer: %s", blobOut)

	if err := blob.StartDefaultBlobWriter(blobOut, blobIndex, bufSizeBytes, cfg.BlobAutoCleanup, string(cfg.LogLevel), database); err != nil {
		log.Printf("ERROR: Failed to start blob writer: %v", err)
		log.Printf("FALLBACK: Using direct filesystem mode instead")
		return false
	}

	log.Printf("Blob writer started successfully: %s, buffer=%dMB, auto-cleanup=%v", blobOut, cfg.BlobBufferMB, cfg.BlobAutoCleanup)
	return true
}
