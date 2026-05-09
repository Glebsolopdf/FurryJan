package parsers

import (
	"fmt"

	"furryjan/internal/config"
)

func parseConfigData() (*config.Config, error) {
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
