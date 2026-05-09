package parsers

import (
	"fmt"

	"furryjan/internal/db"
)

func ParseDatabase(dbPath string) (*db.DB, error) {
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database at '%s': %w", dbPath, err)
	}
	return database, nil
}
