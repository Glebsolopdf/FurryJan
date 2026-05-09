package config

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Uninstall removes the application files and user data.
func Uninstall(ctx context.Context) (string, error) {
	ex, err := os.Executable()
	if err != nil {
		if runtime.GOOS == "windows" {
			ex = "furryjan.exe"
		} else {
			ex = "/usr/bin/furryjan"
		}
	}

	configPath, _ := ConfigPath()
	configDir := filepath.Dir(configPath)

	if runtime.GOOS == "windows" {
		if err := uninstallWindows(ctx, ex, configDir); err != nil {
			return ex, err
		}
		return ex, nil
	}

	return ex, uninstallLinux(ctx, ex, configDir)
}

func uninstallWindows(ctx context.Context, exePath, configDir string) error {
	_ = ctx
	_ = os.RemoveAll(configDir)

	installDir := filepath.Dir(exePath)
	cleanupScript := filepath.Join(os.TempDir(), "furryjan_uninstall.cmd")
	script := strings.Join([]string{
		"@echo off",
		"timeout /t 2 /nobreak >nul",
		fmt.Sprintf("rmdir /s /q \"%s\"", configDir),
		fmt.Sprintf("del /f /q \"%s\"", exePath),
		fmt.Sprintf("rmdir /s /q \"%s\"", installDir),
		"del /f /q \"%~f0\"",
	}, "\r\n") + "\r\n"

	if err := os.WriteFile(cleanupScript, []byte(script), 0600); err != nil {
		return fmt.Errorf("cannot create uninstall script: %w", err)
	}

	cmd := exec.Command("cmd", "/C", "start", "", "/min", cleanupScript)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot start uninstall script: %w", err)
	}

	return nil
}

func uninstallLinux(ctx context.Context, ex, configDir string) error {

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Введите пароль sudo: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	cmdRmConfig := exec.CommandContext(ctx, "sudo", "-S", "rm", "-rf", configDir)
	cmdRmConfig.Stdin = strings.NewReader(password + "\n")
	err1 := cmdRmConfig.Run()

	cmdRmBin := exec.CommandContext(ctx, "sudo", "-S", "rm", "-f", ex)
	cmdRmBin.Stdin = strings.NewReader(password + "\n")
	err2 := cmdRmBin.Run()

	cmdRmLocales := exec.CommandContext(ctx, "sudo", "-S", "rm", "-rf", "/usr/share/furryjan")
	cmdRmLocales.Stdin = strings.NewReader(password + "\n")
	err3 := cmdRmLocales.Run()

	cmdRmDesktop := exec.CommandContext(ctx, "sudo", "-S", "rm", "-f", "/usr/share/applications/furryjan.desktop")
	cmdRmDesktop.Stdin = strings.NewReader(password + "\n")
	err4 := cmdRmDesktop.Run()

	cmdRmIcon := exec.CommandContext(ctx, "sudo", "-S", "rm", "-f", "/usr/share/pixmaps/furryjan.png")
	cmdRmIcon.Stdin = strings.NewReader(password + "\n")
	err5 := cmdRmIcon.Run()

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	if err3 != nil {
		return err3
	}
	if err4 != nil {
		return err4
	}
	if err5 != nil {
		return err5
	}

	return nil
}
