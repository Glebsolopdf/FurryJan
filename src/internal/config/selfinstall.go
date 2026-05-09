package config

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"furryjan/assets"
)

func EnsureInstalled(ctx context.Context) error {
	if runtime.GOOS == "windows" {
		return ensureInstalledWindows(ctx)
	}

	targetPath := "/usr/bin/furryjan"

	if fileExists(targetPath) {
		return nil
	}

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine current executable: %w", err)
	}

	if currentExe == targetPath {
		return nil
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║  Installing Furryjan to system...    ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()

	err = copyWithSudo(ctx, currentExe, targetPath)
	if err != nil {
		fmt.Printf("⚠ Warning: could not install to /usr/bin: %v\n", err)
		fmt.Println("The program will continue, but auto-update won't be available.")
		return nil
	}

	exeDir := filepath.Dir(currentExe)
	i18nSrc := filepath.Join(exeDir, "i18n")
	if !fileExists(i18nSrc) {
		for _, relPath := range []string{"./i18n", "../i18n"} {
			candidate := filepath.Join(exeDir, relPath)
			if fileExists(candidate) {
				i18nSrc = candidate
				break
			}
		}
	}

	if fileExists(i18nSrc) {
		i18nDst := "/usr/share/furryjan/i18n"
		_ = copyDirWithSudo(ctx, i18nSrc, i18nDst)
	}

	_ = createDesktopFile(ctx)
	_ = installIcon(ctx)

	fmt.Println("✓ Furryjan installed to /usr/bin/furryjan")
	fmt.Println()

	return nil
}

func ensureInstalledWindows(ctx context.Context) error {
	_ = ctx

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
			localAppData = filepath.Join(userProfile, "AppData", "Local")
		}
	}
	if localAppData == "" {
		return nil
	}

	targetDir := filepath.Join(localAppData, "Programs", "Furryjan")
	targetPath := filepath.Join(targetDir, "furryjan.exe")
	if fileExists(targetPath) {
		return nil
	}

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine current executable: %w", err)
	}

	if strings.EqualFold(currentExe, targetPath) {
		return nil
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("cannot create install directory '%s': %w", targetDir, err)
	}

	if err := copyFile(currentExe, targetPath); err != nil {
		return fmt.Errorf("cannot copy executable to '%s': %w", targetPath, err)
	}

	fmt.Println()
	fmt.Printf("✓ Furryjan installed to %s\n", targetPath)
	fmt.Println()

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyWithSudo(ctx context.Context, src, dst string) error {
	cmd := exec.CommandContext(ctx, "sudo", "install", "-m", "0755", src, dst)
	err := cmd.Run()
	if err == nil {
		return nil
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot open source: %w", err)
	}
	defer srcFile.Close()

	cmd = exec.CommandContext(ctx, "sudo", "tee", dst)
	cmd.Stdin = srcFile
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("sudo copy failed: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func copyDirWithSudo(ctx context.Context, src, dst string) error {
	cmd := exec.CommandContext(ctx, "sudo", "cp", "-r", src, dst)
	return cmd.Run()
}

func createDesktopFile(ctx context.Context) error {
	desktopContent := `[Desktop Entry]
Version=1.0
Type=Application
Name=Furryjan
Comment=e621 Content Downloader
Exec=/usr/bin/furryjan
Terminal=true
Categories=Utility;Network;FileTransfer;
Icon=/usr/share/pixmaps/furryjan.png
StartupNotify=true
`
	cmd := exec.CommandContext(ctx, "sudo", "tee", "/usr/share/applications/furryjan.desktop")
	cmd.Stdin = strings.NewReader(desktopContent)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return err
	}
	exec.CommandContext(ctx, "sudo", "chmod", "644", "/usr/share/applications/furryjan.desktop").Run()
	return nil
}

func installIcon(ctx context.Context) error {
	iconData, err := assets.FS.ReadFile("icon.png")
	if err != nil {
		return nil
	}

	tempFile := filepath.Join(os.TempDir(), "furryjan_icon.png")
	if err := os.WriteFile(tempFile, iconData, 0644); err != nil {
		return err
	}
	defer os.Remove(tempFile)

	cmd := exec.CommandContext(ctx, "sudo", "cp", tempFile, "/usr/share/pixmaps/furryjan.png")
	if err := cmd.Run(); err != nil {
		return err
	}
	exec.CommandContext(ctx, "sudo", "chmod", "644", "/usr/share/pixmaps/furryjan.png").Run()
	return nil
}
