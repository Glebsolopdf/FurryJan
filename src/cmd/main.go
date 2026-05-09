package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"furryjan/i18n"
	"furryjan/internal/config"
	"furryjan/internal/db"
	"furryjan/internal/downloader/blob"
	"furryjan/internal/ui"
)

func main() {
	if err := run(); err != nil {
		log.Printf("Application failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := config.EnsureInstalled(ctx); err != nil {
		log.Printf("[Warning] Self-installation check failed: %v", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		if errors.Is(err, config.ErrSetupCancelled) {
			return nil
		}
		return err
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("failed to open database at '%s': %w", cfg.DBPath, err)
	}
	defer database.Close()

	if err := os.MkdirAll(cfg.DownloadDir, 0755); err != nil {
		return fmt.Errorf("нет прав на запись в %s. Пожалуйста, проверьте права доступа: %w", cfg.DownloadDir, err)
	}

	if err := loadLocales(); err != nil {
		log.Printf("Warning: Failed to load translations: %v", err)
		log.Printf("Falling back to en-US keys as translations")
	}

	i18n.SetGlobal(cfg.Language)
	log.Printf("[Config] Language: %s", cfg.Language)
	log.Printf("[Config] BlobWriterEnabled=%v, BlobBufferMB=%d, BlobAutoCleanup=%v, LogLevel=%s",
		cfg.BlobWriterEnabled, cfg.BlobBufferMB, cfg.BlobAutoCleanup, cfg.LogLevel)
	log.Printf("[Startup] Download directory: %s", cfg.DownloadDir)
	log.Printf("[Startup] Database: %s", cfg.DBPath)

	if cfg.BlobWriterEnabled {
		blobOut := filepath.Join(cfg.DownloadDir, "data.blob")
		blobIndex := filepath.Join(cfg.DownloadDir, "data.index.json")
		bufSizeBytes := cfg.BlobBufferMB * 1024 * 1024
		log.Printf("[Startup] Starting blob writer: %s", blobOut)
		if err := blob.StartDefaultBlobWriter(blobOut, blobIndex, bufSizeBytes, cfg.BlobAutoCleanup, string(cfg.LogLevel), database); err != nil {
			log.Printf("ERROR: Failed to start blob writer: %v", err)
			log.Printf("FALLBACK: Using direct filesystem mode instead")
		} else {
			defer func() {
				if err := blob.StopDefaultBlobWriter(); err != nil {
					log.Printf("Warning: Error stopping blob writer: %v", err)
				}
			}()
			log.Printf("Blob writer started successfully: %s, buffer=%dMB, auto-cleanup=%v", blobOut, cfg.BlobBufferMB, cfg.BlobAutoCleanup)
		}
	} else {
		log.Printf("Blob writer disabled in settings, using direct filesystem mode")
	}

	if err := ui.Run(ctx, cfg, database); err != nil {
		if errors.Is(err, ui.ErrRestartRequested) || errors.Is(err, ui.ErrExitRequested) {
			return nil
		}
		return fmt.Errorf("ui error: %w", err)
	}

	return nil
}

func loadConfig() (*config.Config, error) {
	if config.Exists() {
		cfg, err := config.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		if cfg.IsComplete() {
			return cfg, nil
		}
	}

	cfg, err := config.RunSetup()
	if err != nil {
		return nil, fmt.Errorf("setup failed: %w", err)
	}
	return cfg, nil
}

func loadLocales() error {
	if err := i18n.LoadFromEmbed(); err == nil {
		log.Printf("[Config] Translations loaded from embedded binary")
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("embedded locales unavailable and executable path not available: %w", err)
	}

	adjacentPath := filepath.Join(filepath.Dir(exePath), "i18n", "locales")
	if err := i18n.Load(adjacentPath); err == nil {
		log.Printf("[Config] Translations loaded from: %s", adjacentPath)
		return nil
	}

	return fmt.Errorf("embedded locales unavailable and adjacent locales path failed: %s", adjacentPath)
}
