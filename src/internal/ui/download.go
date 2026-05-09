package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"furryjan/i18n"
	"furryjan/internal/config"
	"furryjan/internal/db"
	"furryjan/internal/downloader"
	"furryjan/internal/downloader/blob"
)

func RunDownloadFlow(ctx context.Context, cfg *config.Config, database *db.DB) error {
	ClearScreen()
	fmt.Println()
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println(i18n.T("download", "searchType"))
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println("1) " + i18n.T("download", "tags"))
	fmt.Println("2) " + i18n.T("download", "popular"))
	fmt.Println("3) " + i18n.T("download", "latest"))
	fmt.Println("4) " + i18n.T("download", "highRated"))
	fmt.Println("5) " + i18n.T("download", "cancel"))
	fmt.Println("─────────────────────────────────────────────────────")

	choice := Prompt(i18n.T("prompt", "choose"))
	if IsExitInput(choice) {
		return ErrExitRequested
	}

	var tags []string

	switch choice {
	case "1":
		tagsInput := Prompt(i18n.T("download", "tagInput"))
		if IsExitInput(tagsInput) {
			return ErrExitRequested
		}
		if tagsInput == "" {
			fmt.Println(i18n.T("download", "cancel"))
			return nil
		}
		tags = strings.Fields(tagsInput)

	case "2":
		tags = []string{"order:hot"}
	case "3":
		tags = []string{"order:latest"}
	case "4":
		tags = []string{"order:score"}
	case "5":
		return nil
	default:
		PrintError(i18n.T("prompt", "choose"))
		return nil
	}

	// Get limit
	limitStr := Prompt(i18n.T("download", "limit"))
	if IsExitInput(limitStr) {
		return ErrExitRequested
	}
	limit := 0
	if limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	// Get dry-run option
	dryRun := Confirm(i18n.T("download", "dryRun"), false)

	// Check if file filters are configured
	hasTypeFilter := len(cfg.AllowedTypes) < 3 // Not all types allowed
	hasSizeFilter := cfg.MaxSizeMB > 0
	if !hasTypeFilter && !hasSizeFilter {
		fmt.Println()
		PrintError(i18n.T("download", "limitWarning"))
	}

	// Download
	ClearScreen()
	fmt.Println()

	downloadDir := cfg.DownloadDir

	PrintInfo(fmt.Sprintf("Searching posts with tags: %s", strings.Join(tags, " ")))

	opts := downloader.Options{
		Tags:        tags,
		Limit:       limit,
		DryRun:      dryRun,
		Verbose:     false,
		DownloadDir: downloadDir,
	}

	result, err := downloader.Run(ctx, cfg, database, opts)
	if err != nil {
		PrintError(fmt.Sprintf("%s: %v", i18n.T("error", "error"), err))
		return nil
	}

	if cfg.BlobWriterEnabled {
		flushDone := make(chan error, 1)
		go func() {
			flushDone <- blob.FlushDefaultBlobWriter()
		}()

		// Wait for flush to complete
		flushErr := <-flushDone

		if flushErr != nil {
			fmt.Println()
			PrintError(fmt.Sprintf("%s: %v", i18n.T("error", "error"), flushErr))
			return nil
		}
		fmt.Println()
	}

	// Extract blob files if blob writer was used
	if cfg.BlobWriterEnabled {
		blobOut := filepath.Join(downloadDir, "data.blob")
		if _, err := os.Stat(blobOut); err == nil {
			if err := blob.ExtractAndCleanupDefaultBlobWriter(downloadDir); err != nil {
				fmt.Println()
				PrintError(fmt.Sprintf("%s: %v", i18n.T("error", "error"), err))
				return nil
			}
		}
	}

	ClearScreen()
	fmt.Println()
	PrintSuccess(fmt.Sprintf(i18n.T("download", "downloaded"),
		result.Downloaded,
		result.Skipped,
		result.Failed))
	fmt.Println()
	PrintInfo(fmt.Sprintf("Files saved in: %s", result.DownloadDir))
	fmt.Println()

	WaitForEnter(i18n.T("prompt", "enterToContinue"))
	return nil
}
