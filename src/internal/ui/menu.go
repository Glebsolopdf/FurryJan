package ui

import (
	"fmt"
	"log"

	"furryjan/i18n"
	"furryjan/internal/config"
	"furryjan/internal/db"
)

func Run(cfg *config.Config, database *db.DB) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC in UI: %v", r)
		}
	}()

	for {
		ClearScreen()
		fmt.Println()
		fmt.Println("═════════════════════════════════════")
		fmt.Println("        Furryjan  v1.0            ")
		fmt.Println("═══════════════════════════════════")
		fmt.Printf("   1.  %-29s\n", i18n.T("menu", "download"))
		fmt.Printf("   2.  %-29s\n", i18n.T("menu", "history"))
		fmt.Printf("   3.  %-29s\n", i18n.T("menu", "archive"))
		fmt.Printf("   4.  %-29s\n", i18n.T("menu", "settings"))
		fmt.Printf("   5.  %-29s\n", i18n.T("menu", "exit"))
		fmt.Println("═════════════════════════════════════")

		choice := Prompt(i18n.T("prompt", "choose"))

		switch choice {
		case "1":
			err := RunDownloadFlow(cfg, database)
			if err != nil {
				PrintError(fmt.Sprintf("%s: %v", i18n.T("error", "error"), err))
			}

		case "2":
			err := RunHistoryFlow(cfg, database)
			if err != nil {
				PrintError(fmt.Sprintf("%s: %v", i18n.T("error", "error"), err))
			}

		case "3":
			err := RunArchiveFlow(cfg, database)
			if err != nil {
				PrintError(fmt.Sprintf("%s: %v", i18n.T("error", "error"), err))
			}

		case "4":
			err := RunSettingsFlow(cfg)
			if err != nil {
				PrintError(fmt.Sprintf("%s: %v", i18n.T("error", "error"), err))
			}

		case "5":
			return nil

		default:
			PrintError(i18n.T("menu", "choose"))
		}
	}
}
