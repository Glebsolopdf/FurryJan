package config

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Uninstall removes the application (config folder and binary) with sudo
// Prompts for password and performs deletion with elevated privileges
func Uninstall() (string, error) {
	// Get the binary path
	ex, err := os.Executable()
	if err != nil {
		ex = "/usr/bin/furryjan" // fallback
	}

	// Get config path
	configPath, _ := ConfigPath()
	configDir := configPath[:len(configPath)-len("config.json")]
	configDir = configDir[:len(configDir)-1] // Remove trailing slash

	reader := bufio.NewReader(os.Stdin)

	// Request password
	fmt.Print("Введите пароль sudo: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Remove config directory with sudo
	cmdRmConfig := exec.Command("sudo", "-S", "rm", "-rf", configDir)
	cmdRmConfig.Stdin = strings.NewReader(password + "\n")
	err1 := cmdRmConfig.Run()

	// Remove binary with sudo
	cmdRmBin := exec.Command("sudo", "-S", "rm", "-f", ex)
	cmdRmBin.Stdin = strings.NewReader(password + "\n")
	err2 := cmdRmBin.Run()

	// Remove locales with sudo
	cmdRmLocales := exec.Command("sudo", "-S", "rm", "-rf", "/usr/share/furryjan")
	cmdRmLocales.Stdin = strings.NewReader(password + "\n")
	err3 := cmdRmLocales.Run()

	// Remove desktop file with sudo
	cmdRmDesktop := exec.Command("sudo", "-S", "rm", "-f", "/usr/share/applications/furryjan.desktop")
	cmdRmDesktop.Stdin = strings.NewReader(password + "\n")
	err4 := cmdRmDesktop.Run()

	// Remove icon with sudo
	cmdRmIcon := exec.Command("sudo", "-S", "rm", "-f", "/usr/share/pixmaps/furryjan.png")
	cmdRmIcon.Stdin = strings.NewReader(password + "\n")
	err5 := cmdRmIcon.Run()

	// Return the first error if any, otherwise return nil
	if err1 != nil {
		return ex, err1
	}
	if err2 != nil {
		return ex, err2
	}
	if err3 != nil {
		return ex, err3
	}
	if err4 != nil {
		return ex, err4
	}
	if err5 != nil {
		return ex, err5
	}

	return ex, nil
}
