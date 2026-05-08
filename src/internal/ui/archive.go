package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"furryjan/i18n"
	"furryjan/internal/archiver"
	"furryjan/internal/config"
	"furryjan/internal/db"
)

func RunArchiveFlow(cfg *config.Config, database *db.DB) error {
	ClearScreen()
	fmt.Println()
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Printf("%-51s\n", i18n.T("menu", "archive"))
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Printf("%-51s\n", i18n.T("archive", "whatToArchive"))
	fmt.Printf("1) %-49s\n", i18n.T("archive", "allDownloads"))
	fmt.Printf("2) %-49s\n", i18n.T("archive", "selectTags"))
	fmt.Printf("3) %-49s\n", i18n.T("menu", "exit"))
	fmt.Println("─────────────────────────────────────────────────────")

	choice := Prompt(i18n.T("prompt", "choose"))

	includeTags := []string{}
	switch choice {
	case "1":
		// Archive all
	case "2":
		tagsInput := Prompt(i18n.T("archive", "enterTagsFilter"))
		if tagsInput != "" {
			includeTags = strings.Fields(tagsInput)
		}
	case "3":
		return nil
	default:
		PrintError(i18n.T("prompt", "choose"))
		return nil
	}

	// Get exclusions
	excludeInput := Prompt(i18n.T("archive", "excludeTags"))
	excludeTags := []string{}
	if excludeInput != "" {
		excludeTags = strings.Fields(excludeInput)
	}

	// Get output directory
	ClearScreen()
	fmt.Println()
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Printf("%-51s\n", i18n.T("archive", "whereToSave"))
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Printf("1) %-49s\n", i18n.T("archive", "inDownloadFolder"))
	fmt.Printf("2) %-49s\n", i18n.T("archive", "customPath"))
	fmt.Printf("3) %-49s\n", i18n.T("menu", "exit"))
	fmt.Println("─────────────────────────────────────────────────────")

	dirChoice := Prompt(i18n.T("prompt", "choose"))

	var outputDir string
	switch dirChoice {
	case "1":
		outputDir = cfg.DownloadDir
	case "2":
		customPath := Prompt(i18n.T("archive", "enterPath"))
		if customPath == "" {
			return nil
		}
		outputDir = customPath
	case "3":
		return nil
	default:
		PrintError(i18n.T("prompt", "choose"))
		return nil
	}

	// Get output filename
	outputName := Prompt(i18n.T("archive", "archiveName"))
	if outputName == "" {
		outputName = i18n.T("archive", "defaultArchiveName")
	}

	// Full path
	fullPath := filepath.Join(outputDir, outputName)

	// Archive
	ClearScreen()
	fmt.Println()
	PrintInfo(i18n.T("archive", "archivingProgress"))

	// Call archiver
	opts := archiver.Options{
		IncludeTags: includeTags,
		ExcludeTags: excludeTags,
		OutputPath:  fullPath,
		Verbose:     true,
	}

	err := archiver.Run(cfg, opts)
	if err != nil {
		PrintError(fmt.Sprintf("%s: %s", i18n.T("error", "error"), err.Error()))
		return err
	}

	PrintSuccess(i18n.T("archive", "archiveCreated"))

	return nil
}
