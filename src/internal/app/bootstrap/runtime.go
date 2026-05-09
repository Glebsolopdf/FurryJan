package bootstrap

import (
	"log"

	"furryjan/i18n"
	"furryjan/internal/config"
)

func ParseLocalesData() {
	if err := loadLocales(); err != nil {
		log.Printf("Warning: Failed to load translations: %v", err)
		log.Printf("Falling back to en-US keys as translations")
	}
}

func ParseRuntimeData(cfg *config.Config) {
	i18n.SetGlobal(cfg.Language)
	log.Printf("[Config] Language: %s", cfg.Language)
	log.Printf("[Config] BlobWriterEnabled=%v, BlobBufferMB=%d, BlobAutoCleanup=%v, LogLevel=%s",
		cfg.BlobWriterEnabled, cfg.BlobBufferMB, cfg.BlobAutoCleanup, cfg.LogLevel)
	log.Printf("[Startup] Download directory: %s", cfg.DownloadDir)
	log.Printf("[Startup] Database: %s", cfg.DBPath)
}
