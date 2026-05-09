package config

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrSetupCancelled = errors.New("setup cancelled by user")

// setupStrings contains localized strings for setup wizard
type setupStrings struct {
	welcome          string
	langPrompt       string
	langEnglish      string
	langRussian      string
	step1Prompt      string
	step2Prompt      string
	step3Prompt      string
	step3Default     string
	errorUsername    string
	errorAPIKey      string
	errorConfigDir   string
	errorDownloadDir string
	errorSaveConfig  string
	successSaved     string
}

var setupLocales = map[string]setupStrings{
	"en": {
		welcome:          "Welcome to Furryjan!",
		langPrompt:       "Choose your language:",
		langEnglish:      "English",
		langRussian:      "Русский",
		step1Prompt:      "Step 1/4  Enter your username on e621:",
		step2Prompt:      "Step 2/4  Enter your API key (Open Profile Settings on e621 → Manage API Access):",
		step3Prompt:      "Step 3/4  Download folder (Enter = default):",
		step3Default:     "",
		errorUsername:    "username cannot be empty",
		errorAPIKey:      "API key cannot be empty",
		errorConfigDir:   "cannot create config directory '%s': %w",
		errorDownloadDir: "cannot create download directory '%s': %w",
		errorSaveConfig:  "cannot save config: %w",
		successSaved:     "✓ Config saved: %s",
	},
	"ru": {
		welcome:          "Добро пожаловать в Furryjan!",
		langPrompt:       "Выбор языка:",
		langEnglish:      "English",
		langRussian:      "Русский",
		step1Prompt:      "Шаг 1/4  Введите ваш username на e621:",
		step2Prompt:      "Шаг 2/4  Введите API-ключ (Откройте Настройки профиля на e621 → Manage API Access):",
		step3Prompt:      "Шаг 3/4  Папка для загрузок (Enter = по умолчанию):",
		step3Default:     "",
		errorUsername:    "username cannot be empty",
		errorAPIKey:      "API key cannot be empty",
		errorConfigDir:   "cannot create config directory '%s': %w",
		errorDownloadDir: "cannot create download directory '%s': %w",
		errorSaveConfig:  "cannot save config: %w",
		successSaved:     "✓ Конфиг сохранён: %s",
	},
}

// RunSetup runs the initial setup wizard
func RunSetup() (*Config, error) {
	fmt.Println()
	fmt.Println("╔═════════════════════════════════════════════════╗")
	fmt.Println("║                                                 ║")
	fmt.Println("║              Welcome to Furryjan!               ║")
	fmt.Println("║                                                 ║")
	fmt.Println("║  Before we start, please provide the necessary  ║")
	fmt.Println("║   data for the downloader to work properly      ║")
	fmt.Println("║                                                 ║")
	fmt.Println("║                (Ctrl+C to exit)                 ║")
	fmt.Println("║                                                 ║")
	fmt.Println("╚═════════════════════════════════════════════════╝")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	homeDir, _ := os.UserHomeDir()
	defaultDownloadDir := filepath.Join(homeDir, "Downloads", "Furryjan")

	// Step 0: Language
	fmt.Println(" Step 0/4  Choose your language:")
	fmt.Println("  1) English")
	fmt.Println("  2) Русский")
	fmt.Println("  (.delete to uninstall)")
	fmt.Print(" > ")
	langChoice, _ := reader.ReadString('\n')
	langChoice = strings.TrimSpace(langChoice)

	if langChoice == ".delete" {
		_, err := Uninstall(context.Background())
		if err != nil {
			fmt.Printf("Error during uninstallation: %v\n", err)
			return nil, err
		}
		return nil, ErrSetupCancelled
	}

	language := "en"
	if langChoice == "2" {
		language = "ru"
	}

	loc := setupLocales[language]

	// Step 1: Username
	if language == "ru" {
		fmt.Printf(" Шаг 1/4  Введите ваш username на e621:\n > ")
	} else {
		fmt.Printf(" Step 1/4  Enter your username on e621:\n > ")
	}
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf(loc.errorUsername)
	}

	// Step 2: API Key
	if language == "ru" {
		fmt.Printf(" Шаг 2/4  Введите API-ключ (Откройте Настройки профиля на e621 → Manage API Access):\n > ")
	} else {
		fmt.Printf(" Step 2/4  Enter your API key (Open Profile Settings on e621 → Manage API Access):\n > ")
	}
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf(loc.errorAPIKey)
	}

	// Step 3: Download directory
	if language == "ru" {
		fmt.Printf(" Шаг 3/4  Папка для загрузок (Enter = %s):\n > ", defaultDownloadDir)
	} else {
		fmt.Printf(" Step 3/4  Download folder (Enter = %s):\n > ", defaultDownloadDir)
	}
	downloadDir, _ := reader.ReadString('\n')
	downloadDir = strings.TrimSpace(downloadDir)
	if downloadDir == "" {
		downloadDir = defaultDownloadDir
	}

	cfg := Default()
	cfg.Username = username
	cfg.APIKey = apiKey
	cfg.DownloadDir = downloadDir
	cfg.Language = language

	// Ensure config directory exists
	configPath, _ := ConfigPath()
	configDir := filepath.Dir(configPath)
	err := os.MkdirAll(configDir, 0700)
	if err != nil {
		return nil, fmt.Errorf(loc.errorConfigDir, configDir, err)
	}

	// Create download directory
	err = os.MkdirAll(cfg.DownloadDir, 0755)
	if err != nil {
		return nil, fmt.Errorf(loc.errorDownloadDir, cfg.DownloadDir, err)
	}

	err = cfg.Save()
	if err != nil {
		return nil, fmt.Errorf(loc.errorSaveConfig, err)
	}

	fmt.Println()
	fmt.Printf(loc.successSaved+"\n", configPath)
	fmt.Println()

	return cfg, nil
}

// DeleteConfig removes the configuration directory and all its contents
func DeleteConfig() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(path)
	return os.RemoveAll(configDir)
}
