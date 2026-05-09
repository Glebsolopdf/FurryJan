package parsers

import (
	"context"
	"log"

	"furryjan/internal/app/bootstrap"
	"furryjan/internal/config"
)

func ParseData(ctx context.Context) (*bootstrap.Data, error) {
	if err := config.EnsureInstalled(ctx); err != nil {
		log.Printf("[Warning] Self-installation check failed: %v", err)
	}

	cfg, err := parseConfigData()
	if err != nil {
		return nil, err
	}

	database, err := parseDatabaseData(cfg)
	if err != nil {
		return nil, err
	}

	if err := parseDownloadDirData(cfg); err != nil {
		database.Close()
		return nil, err
	}

	bootstrap.ParseLocalesData()
	bootstrap.ParseRuntimeData(cfg)

	blobOn := bootstrap.ParseBlobData(cfg, database)

	return &bootstrap.Data{
		Ctx:      ctx,
		Config:   cfg,
		Database: database,
		BlobOn:   blobOn,
	}, nil
}
