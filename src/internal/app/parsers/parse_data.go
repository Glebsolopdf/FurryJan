package parsers

import (
	"context"
	"log"

	"furryjan/internal/app/bootstrap"
	"furryjan/internal/config"
)

func ParseData(ctx context.Context) (*bootstrap.Data, func(), error) {
	if err := config.EnsureInstalled(ctx); err != nil {
		log.Printf("[Warning] Self-installation check failed: %v", err)
	}

	startupData := &bootstrap.Data{}
	cleanup := bootstrap.NewCleanupStack()

	cfg, err := ParseConfig()
	if err != nil {
		return nil, nil, err
	}
	startupData.Config = cfg

	database, err := ParseDatabase(cfg.DBPath)
	if err != nil {
		return nil, nil, err
	}
	startupData.Database = database
	cleanup.Add(func() {
		_ = database.Close()
	})

	if err := ParseDownloadDir(cfg.DownloadDir); err != nil {
		return nil, nil, err
	}

	if err := bootstrap.LoadLocales(); err != nil {
		log.Printf("Warning: Failed to load translations: %v", err)
		log.Printf("Falling back to en-US keys as translations")
	}

	return startupData, cleanup.Run, nil
}
