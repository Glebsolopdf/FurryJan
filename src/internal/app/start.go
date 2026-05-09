package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"furryjan/internal/app/bootstrap"
	"furryjan/internal/app/parsers"
	"furryjan/internal/config"
	"furryjan/internal/ui"
)

func Run() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	processStartError(runLifecycle())
}

func runLifecycle() error {
	for {
		err := start()
		if errors.Is(err, ui.ErrRestartRequested) {
			continue
		}
		return err
	}
}

func processStartError(err error) {
	if err == nil {
		return
	}

	log.Printf("Application failed: %v", err)
	os.Exit(1)
}

func start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startupData, cleanup, err := parsers.ParseData(ctx)
	if err != nil {
		if errors.Is(err, config.ErrSetupCancelled) {
			return nil
		}
		return err
	}
	defer cleanup()

	bootstrap.ApplyRuntime(startupData.Config)

	stopBlobWriter := bootstrap.StartBlobWriter(startupData.Config, startupData.Database)
	defer stopBlobWriter()

	if err := ui.Run(ctx, startupData.Config, startupData.Database); err != nil {
		if errors.Is(err, ui.ErrRestartRequested) {
			return err
		}
		if errors.Is(err, ui.ErrExitRequested) {
			return nil
		}
		return fmt.Errorf("ui error: %w", err)
	}

	return nil
}
