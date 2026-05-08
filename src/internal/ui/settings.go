package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"furryjan/i18n"
	"furryjan/internal/config"
)

func RunSettingsFlow(cfg *config.Config) error {
	for {
		ClearScreen()
		fmt.Println()
		fmt.Println("═════════════════════════════════════════════════════════════════")
		fmt.Printf("%-67s\n", i18n.T("settings", "title"))
		fmt.Println("═════════════════════════════════════════════════════════════════")
		fmt.Printf("Language                          │ %-27s\n", getLanguageName(cfg.Language))
		fmt.Printf("Username                          │ %-27s\n", Truncate(cfg.Username, 27))
		fmt.Printf("API Key                           │ %-27s\n", Truncate(cfg.MaskAPIKey(), 27))
		fmt.Printf("Download Directory                │ %-27s\n", Truncate(cfg.DownloadDir, 27))
		fmt.Printf("Allowed File Types                │ %-27s\n", Truncate(strings.Join(cfg.AllowedTypes, ","), 27))
		fmt.Printf("Max File Size (MB)                │ %-27d\n", cfg.MaxSizeMB)
		fmt.Printf("Blob Writer                       │ %-27s\n", BoolToString(cfg.BlobWriterEnabled))
		fmt.Printf("Buffer Size (MB)                  │ %-27d\n", cfg.BlobBufferMB)
		fmt.Printf("Auto-delete Blob Files            │ %-27s\n", BoolToString(cfg.BlobAutoCleanup))
		fmt.Printf("Log Level                         │ %-27s\n", string(cfg.LogLevel))
		fmt.Println("═════════════════════════════════════════════════════════════════")
		fmt.Printf("%-67s\n", i18n.T("settings", "menuLabels"))
		fmt.Println("═════════════════════════════════════════════════════════════════")

		choice := Prompt(i18n.T("prompt", "choose"))

		switch choice {
		case "1":
			runSettingsEditor(cfg)
		case "2":
			if Confirm(i18n.T("settings", "resetConfirm"), false) {
				err := config.DeleteConfig()
				if err != nil {
					PrintError(fmt.Sprintf("%s: %v", i18n.T("error", "error"), err))
				} else {
					PrintSuccess(i18n.T("settings", "settingsDeleted"))
					fmt.Println()
					WaitForEnter(i18n.T("prompt", "enterToContinue"))
					return nil
				}
			}
		case "3":
			if Confirm(i18n.T("settings", "uninstallConfirm"), false) {
				_, err := config.Uninstall()
				fmt.Println()
				if err != nil {
					PrintError(fmt.Sprintf("%s: %v", i18n.T("settings", "uninstallFailed"), err))
					fmt.Println()
					WaitForEnter(i18n.T("prompt", "enterToContinue"))
				} else {
					PrintSuccess(i18n.T("settings", "uninstallSuccess"))
					os.Exit(0)
				}
			}
		case "4":
			return nil
		default:
			PrintError(i18n.T("prompt", "choose"))
		}
	}
}

func runSettingsEditor(cfg *config.Config) error {
	ClearScreen()
	fmt.Println()
	fmt.Println(i18n.T("settings", "whichSetting"))
	fmt.Println("1) " + i18n.T("settings", "selectLanguage"))
	fmt.Println("2) " + i18n.T("settings", "selectUsername"))
	fmt.Println("3) " + i18n.T("settings", "selectAPIKey"))
	fmt.Println("4) " + i18n.T("settings", "selectDownloadDir"))
	fmt.Println("5) " + i18n.T("settings", "selectFileTypes"))
	fmt.Println("6) " + i18n.T("settings", "selectMaxSize"))
	fmt.Println("7) " + i18n.T("settings", "blobWriterOption"))
	fmt.Println("8) " + i18n.T("settings", "selectBufferSize"))
	fmt.Println("9) " + i18n.T("settings", "selectAutoCleanup"))
	fmt.Println("10) " + i18n.T("settings", "selectLogLevel"))
	fmt.Println("11) " + i18n.T("settings", "cancel"))

	settingChoice := Prompt(i18n.T("prompt", "choose"))

	switch settingChoice {
	case "1":
		editLanguage(cfg)
	case "2":
		cfg.Username = Prompt(i18n.T("settings", "selectUsername"))
		cfg.Save()
		PrintSuccess(i18n.T("settings", "saved"))
	case "3":
		cfg.APIKey = Prompt(i18n.T("settings", "selectAPIKey"))
		cfg.Save()
		PrintSuccess(i18n.T("settings", "saved"))
	case "4":
		editDownloadDir(cfg)
	case "5":
		editAllowedTypes(cfg)
	case "6":
		editMaxFileSize(cfg)
	case "7":
		editBlobWriter(cfg)
	case "8":
		editBufferSize(cfg)
	case "9":
		editAutoCleanup(cfg)
	case "10":
		editLogLevel(cfg)
	case "11":
		// Cancel
	}
	return nil
}

func editDownloadDir(cfg *config.Config) {
	newDir := Prompt(i18n.T("settings", "selectDownloadDir"))
	if newDir == "" {
		return
	}
	// Expand ~ to home directory
	if strings.HasPrefix(newDir, "~") {
		homeDir, _ := os.UserHomeDir()
		newDir = filepath.Join(homeDir, newDir[1:])
	}
	// Try to create directory with sudo support
	err := CreateDirectoryWithSudo(newDir)
	if err != nil {
		fmt.Println()
		PrintError(i18n.T("settings", "dirNotFound"))
		PrintError(fmt.Sprintf("%s%s", i18n.T("settings", "pathLabel"), newDir))
		PrintError(fmt.Sprintf("%s%v", i18n.T("settings", "reasonLabel"), err))
	} else {
		cfg.DownloadDir = newDir
		cfg.Save()
		PrintSuccess(i18n.T("settings", "created"))
	}
}

func editAllowedTypes(cfg *config.Config) {
	fmt.Println()
	fmt.Println(i18n.T("settings", "selectFileTypes"))
	fmt.Println("1) " + i18n.T("settings", "imagesOnly"))
	fmt.Println("2) " + i18n.T("settings", "imagesAnimations"))
	fmt.Println("3) " + i18n.T("settings", "allTypes"))
	fmt.Println("4) " + i18n.T("settings", "videoOnly"))
	typeChoice := Prompt(i18n.T("prompt", "choose"))

	switch typeChoice {
	case "1":
		cfg.AllowedTypes = []string{"image"}
	case "2":
		cfg.AllowedTypes = []string{"image", "animation", "video"}
	case "3":
		cfg.AllowedTypes = []string{"image", "animation", "video"}
	case "4":
		cfg.AllowedTypes = []string{"video"}
	}
	cfg.Save()
	PrintSuccess(i18n.T("settings", "saved"))
}

func editMaxFileSize(cfg *config.Config) {
	fmt.Println()
	maxStr := Prompt(i18n.T("settings", "selectMaxSize"))
	fmt.Sscanf(maxStr, "%d", &cfg.MaxSizeMB)
	cfg.Save()
	PrintSuccess(i18n.T("settings", "saved"))
}

func editBlobWriter(cfg *config.Config) {
	fmt.Println()
	oldValue := cfg.BlobWriterEnabled
	cfg.BlobWriterEnabled = Confirm(i18n.T("settings", "selectBlobWriter"), cfg.BlobWriterEnabled)
	if oldValue != cfg.BlobWriterEnabled {
		cfg.Save()
		fmt.Println()
		if cfg.BlobWriterEnabled {
			blobPath := filepath.Join(cfg.DownloadDir, "data.blob")
			PrintSuccess(i18n.T("settings", "blobEnabled"))
			fmt.Println()
			fmt.Printf("%s%s\n", i18n.T("settings", "filesWillBe"), blobPath)
			fmt.Println()
			fmt.Println(i18n.T("settings", "restartFurryjan"))
			fmt.Println()
			RestartApplication()
		} else {
			PrintSuccess(i18n.T("settings", "blobDisabled"))
			fmt.Println()
			fmt.Println(i18n.T("settings", "restartFurryjan"))
			fmt.Println()
			RestartApplication()
		}
	}
}

func editBufferSize(cfg *config.Config) {
	fmt.Println()
	bufStr := Prompt(i18n.T("settings", "selectBufferSize"))
	var bufMB int
	fmt.Sscanf(bufStr, "%d", &bufMB)
	if bufMB < 100 {
		bufMB = 100
		PrintError(i18n.T("settings", "minBuffer"))
	} else if bufMB > 2000 {
		bufMB = 2000
		PrintError(i18n.T("settings", "maxBuffer"))
	}
	oldBufMB := cfg.BlobBufferMB
	cfg.BlobBufferMB = bufMB
	cfg.Save()
	PrintSuccess(fmt.Sprintf(i18n.T("settings", "bufferSetTo"), bufMB))
	if oldBufMB != bufMB {
		fmt.Println()
		fmt.Println(i18n.T("settings", "restartFurryjan"))
		fmt.Println()
		RestartApplication()
	}
}

func editAutoCleanup(cfg *config.Config) {
	fmt.Println()
	cfg.BlobAutoCleanup = Confirm(i18n.T("settings", "selectAutoCleanup"), cfg.BlobAutoCleanup)
	cfg.Save()
	PrintSuccess(i18n.T("settings", "saved"))
}

func editLogLevel(cfg *config.Config) {
	fmt.Println()
	fmt.Println(i18n.T("settings", "selectLogLevel"))
	fmt.Println("1) " + i18n.T("settings", "debug"))
	fmt.Println("2) " + i18n.T("settings", "info"))
	fmt.Println("3) " + i18n.T("settings", "warn"))
	fmt.Println("4) " + i18n.T("settings", "logError"))
	levelChoice := Prompt(i18n.T("prompt", "choose"))
	switch levelChoice {
	case "1":
		cfg.LogLevel = config.LogDebug
	case "2":
		cfg.LogLevel = config.LogInfo
	case "3":
		cfg.LogLevel = config.LogWarn
	case "4":
		cfg.LogLevel = config.LogError
	}
	cfg.Save()
	PrintSuccess(fmt.Sprintf("Log level set to %s", cfg.LogLevel))
}

func editLanguage(cfg *config.Config) {
	fmt.Println()
	fmt.Println(i18n.T("settings", "selectLanguage"))
	fmt.Println("1) English")
	fmt.Println("2) Русский")
	langChoice := Prompt(i18n.T("prompt", "choose"))
	oldLang := cfg.Language
	switch langChoice {
	case "1":
		cfg.Language = "en"
	case "2":
		cfg.Language = "ru"
	default:
		PrintError(i18n.T("prompt", "choose"))
		return
	}
	cfg.Save()
	PrintSuccess(fmt.Sprintf(i18n.T("settings", "languageChangedTo"), getLanguageName(cfg.Language)))
	if oldLang != cfg.Language {
		fmt.Println()
		fmt.Println(i18n.T("settings", "restartFurryjan"))
		fmt.Println()
		RestartApplication()
	}
}

func getLanguageName(lang string) string {
	switch lang {
	case "ru":
		return "Русский"
	case "en":
		return "English"
	default:
		return "English"
	}
}
