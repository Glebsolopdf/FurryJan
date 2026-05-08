package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"furryjan/i18n"
	"furryjan/internal/config"
	"furryjan/internal/db"
	"furryjan/internal/downloader/blob"
	"furryjan/internal/ui"
)

func main() {
	exitCode := 0
	defer func() {
		log.Printf("[Shutdown] Cleaning up resources...")

		if err := blob.StopDefaultBlobWriter(); err != nil {
			log.Printf("Warning: Error stopping blob writer: %v", err)
		}
		if err := blob.CleanupDefaultBlobWriter(); err != nil {
			log.Printf("Warning: Error cleaning up blob files: %v", err)
		}

		for i := 0; i < 3; i++ {
			runtime.GC()
		}

		if r := recover(); r != nil {
			log.Printf("❌ Fatal error: %v", r)
			exitCode = 1
		}

		log.Printf("[Shutdown] Exiting with code %d", exitCode)
		os.Exit(exitCode)
	}()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := config.EnsureInstalled()
	if err != nil {
		log.Printf("[Warning] Self-installation check failed: %v", err)
	}

	var cfg *config.Config

	if config.Exists() {
		cfg, err = config.Load()
		if err != nil {
			log.Fatalf("Failed to load config: %v\n\nPlease check your config file or delete it to reconfigure.", err)
		}

		if !cfg.IsComplete() {
			cfg, err = config.RunSetup()
			if err != nil {
				log.Fatalf("Setup failed: %v", err)
			}
		}
	} else {
		cfg, err = config.RunSetup()
		if err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database at '%s': %v\n\nPlease ensure:\n- The path is valid\n- You have write permissions\n- Your antivirus is not blocking the file", cfg.DBPath, err)
	}
	defer database.Close()

	err = ui.CreateDirectoryWithSudo(cfg.DownloadDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create download directory: %v\n", err)
	}

	loadErr := i18n.LoadFromEmbed()
	if loadErr != nil {
		log.Printf("Info: Embedded locales not available: %v, trying filesystem paths", loadErr)

		exePath, err := os.Executable()
		if err != nil {
			log.Printf("Warning: Could not determine executable path: %v", err)
			exePath = "."
		}

		// Try paths in order: exe dir, current dir, project root, /usr/share/furryjan
		possiblePaths := []string{
			filepath.Join(filepath.Dir(exePath), "i18n", "locales"),
			filepath.Join(".", "i18n", "locales"),
			filepath.Join("..", "i18n", "locales"),
			"/usr/share/furryjan/locales",
		}

		for _, path := range possiblePaths {
			loadErr = i18n.Load(path)
			if loadErr == nil {
				log.Printf("[Config] Translations loaded from: %s", path)
				break
			}
		}
	} else {
		log.Printf("[Config] Translations loaded from embedded binary")
	}

	if loadErr != nil {
		log.Printf("Warning: Failed to load translations: %v", loadErr)
		log.Printf("Falling back to en-US keys as translations")
	}

	i18n.SetGlobal(cfg.Language)
	log.Printf("[Config] Language: %s", cfg.Language)

	// Log current configuration
	log.Printf("[Config] BlobWriterEnabled=%v, BlobBufferMB=%d, BlobAutoCleanup=%v, LogLevel=%s",
		cfg.BlobWriterEnabled, cfg.BlobBufferMB, cfg.BlobAutoCleanup, cfg.LogLevel)
	log.Printf("[Startup] Download directory: %s", cfg.DownloadDir)
	log.Printf("[Startup] Database: %s", cfg.DBPath)

	// Start blob writer as the default download backend if enabled
	if cfg.BlobWriterEnabled {
		blobOut := filepath.Join(cfg.DownloadDir, "data.blob")
		blobIndex := filepath.Join(cfg.DownloadDir, "data.index.json")
		bufSizeBytes := cfg.BlobBufferMB * 1024 * 1024
		log.Printf("[Startup] Starting blob writer: %s", blobOut)
		if err := blob.StartDefaultBlobWriter(blobOut, blobIndex, bufSizeBytes, cfg.BlobAutoCleanup, string(cfg.LogLevel)); err != nil {
			log.Printf("ERROR: Failed to start blob writer: %v", err)
			log.Printf("FALLBACK: Using direct filesystem mode instead")
		} else {
			log.Printf("✅ Blob writer started successfully: %s (index: %s), buffer=%dMB, auto-cleanup=%v", blobOut, blobIndex, cfg.BlobBufferMB, cfg.BlobAutoCleanup)
		}
	} else {
		log.Printf("Blob writer disabled in settings, using direct filesystem mode")
	}

	// Run UI
	err = ui.Run(cfg, database)
	if err != nil {
		log.Printf("UI error: %v", err)
		exitCode = 1
		return
	}
}
