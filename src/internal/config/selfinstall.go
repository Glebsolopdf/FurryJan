package config

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func EnsureInstalled() error {
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

	err = copyWithSudo(currentExe, targetPath)
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
		_ = copyDirWithSudo(i18nSrc, i18nDst)
	}

	fmt.Println("✓ Furryjan installed to /usr/bin/furryjan")
	fmt.Println()

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyWithSudo(src, dst string) error {
	cmd := exec.Command("sudo", "install", "-m", "0755", src, dst)
	err := cmd.Run()
	if err == nil {
		return nil
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot open source: %w", err)
	}
	defer srcFile.Close()

	cmd = exec.Command("sudo", "tee", dst)
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

func copyDirWithSudo(src, dst string) error {
	cmd := exec.Command("sudo", "cp", "-r", src, dst)
	return cmd.Run()
}
