package parsers

import (
	"fmt"

	"furryjan/internal/config"
	"furryjan/internal/db"
)

func parseDatabaseData(cfg *config.Config) (*db.DB, error) {
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database at '%s': %w", cfg.DBPath, err)
	}
	return database, nil
}
