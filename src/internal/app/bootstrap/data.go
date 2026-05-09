package bootstrap

import (
	"context"
	"log"

	"furryjan/internal/config"
	"furryjan/internal/db"
	"furryjan/internal/downloader/blob"
)

type Data struct {
	Ctx      context.Context
	Config   *config.Config
	Database *db.DB
	BlobOn   bool
}

func (d *Data) Close() {
	if d.BlobOn {
		if err := blob.StopDefaultBlobWriter(); err != nil {
			log.Printf("Warning: Error stopping blob writer: %v", err)
		}
	}

	if d.Database != nil {
		_ = d.Database.Close()
	}
}
