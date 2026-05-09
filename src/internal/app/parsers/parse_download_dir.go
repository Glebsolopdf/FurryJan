package parsers

import (
	"fmt"
	"os"

	"furryjan/internal/config"
)

func parseDownloadDirData(cfg *config.Config) error {
	if err := os.MkdirAll(cfg.DownloadDir, 0755); err != nil {
		return fmt.Errorf("нет прав на запись в %s. Пожалуйста, проверьте права доступа: %w", cfg.DownloadDir, err)
	}
	return nil
}
