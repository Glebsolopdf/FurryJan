package bootstrap

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"furryjan/i18n"
)

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
