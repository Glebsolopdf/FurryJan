package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"furryjan/internal/app/parsers"
	"furryjan/internal/config"
	"furryjan/internal/ui"
)

func Start() {
	if err := run(); err != nil {
		log.Printf("Application failed: %v", err)
		os.Exit(1)
	}
}

func run() error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	data, err := parsers.ParseData(ctx)
	if err != nil {
		if errors.Is(err, config.ErrSetupCancelled) {
			return nil
		}
		return err
	}
	defer data.Close()

	if err := ui.Run(data.Ctx, data.Config, data.Database); err != nil {
		if errors.Is(err, ui.ErrRestartRequested) || errors.Is(err, ui.ErrExitRequested) {
			return nil
		}
		return fmt.Errorf("ui error: %w", err)
	}

	return nil
}
