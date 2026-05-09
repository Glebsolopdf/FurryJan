package parsers

import (
	"fmt"
	"os"
)

func ParseDownloadDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("нет прав на запись в %s. Пожалуйста, проверьте права доступа: %w", path, err)
	}
	return nil
}
