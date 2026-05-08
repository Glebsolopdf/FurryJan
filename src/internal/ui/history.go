package ui

import (
	"fmt"

	"furryjan/i18n"
	"furryjan/internal/config"
	"furryjan/internal/db"
)

func RunHistoryFlow(cfg *config.Config, database *db.DB) error {
	for {
		ClearScreen()
		fmt.Println()
		fmt.Println("─────────────────────────────────────────────────────")
		fmt.Printf("%-51s\n", i18n.T("menu", "history"))
		fmt.Println("─────────────────────────────────────────────────────")
		fmt.Printf("1) %-49s\n", i18n.T("history", "latest50"))
		fmt.Printf("2) %-49s\n", i18n.T("history", "filterByTag"))
		fmt.Printf("3) %-49s\n", i18n.T("history", "statistics"))
		fmt.Printf("4) %-49s\n", i18n.T("menu", "exit"))
		fmt.Println("─────────────────────────────────────────────────────")

		choice := Prompt(i18n.T("prompt", "choose"))

		switch choice {
		case "1":
			downloads, err := database.QueryHistory(50)
			if err != nil {
				PrintError(fmt.Sprintf("%s: %v", i18n.T("error", "error"), err))
				break
			}

			fmt.Println()
			fmt.Println(i18n.T("history", "latest50Downloads"))
			for i, dl := range downloads {
				fmt.Printf("%3d. [%s] %d - %s (%s)\n", i+1, dl.DownloadedAt.Format("2006-01-02 15:04:05"), dl.PostID, dl.FilePath, db.FormatBytes(dl.FileSize))
			}
			fmt.Println()
			WaitForEnter(i18n.T("prompt", "enterToContinue"))

		case "2":
			tag := Prompt(i18n.T("history", "enterTagFilter"))
			if tag == "" {
				break
			}

			downloads, err := database.QueryHistoryByTag(tag, 50)
			if err != nil {
				PrintError(fmt.Sprintf("%s: %v", i18n.T("error", "error"), err))
				break
			}

			fmt.Println()
			fmt.Printf(i18n.T("history", "downloadsWithTag")+"\n", tag)
			for i, dl := range downloads {
				fmt.Printf("%3d. [%s] %d - %s (%s)\n", i+1, dl.DownloadedAt.Format("2006-01-02 15:04:05"), dl.PostID, dl.FilePath, db.FormatBytes(dl.FileSize))
			}
			fmt.Println()
			WaitForEnter(i18n.T("prompt", "enterToContinue"))

		case "3":
			stats, err := database.GetStats()
			if err != nil {
				PrintError(fmt.Sprintf("%s: %v", i18n.T("error", "error"), err))
				break
			}

			fmt.Println()
			fmt.Println(i18n.T("history", "statisticsLabel"))
			fmt.Printf(i18n.T("history", "totalFiles")+" %d\n", stats.TotalFiles)
			fmt.Printf(i18n.T("history", "totalSize")+" %s\n", db.FormatBytes(stats.TotalSize))
			fmt.Printf(i18n.T("history", "tagsFolders")+" %d\n", stats.TagCount)
			fmt.Printf(i18n.T("history", "firstDownload")+" %s\n", stats.FirstDownload)
			fmt.Printf(i18n.T("history", "latestDownload")+" %s\n", stats.LastDownload)
			fmt.Println()
			WaitForEnter(i18n.T("prompt", "enterToContinue"))

		case "4":
			return nil

		default:
			PrintError(i18n.T("prompt", "choose"))
		}
	}
}
