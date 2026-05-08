package ui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/term"
)

var (
	globalReader *bufio.Reader
	readerOnce   sync.Once
)

func getReader() *bufio.Reader {
	readerOnce.Do(func() {
		globalReader = bufio.NewReader(os.Stdin)
	})
	return globalReader
}

func ClearScreen() {
	// Unix/Linux: ANSI escape codes
	fmt.Print("\033[2J\033[H")
}

func printBoxLine(left, fill, right string, width int) {
	fmt.Print(left)
	for i := 0; i < width-2; i++ {
		fmt.Print(fill)
	}
	fmt.Println(right)
}

func PrintBoxTop(width int) {
	for i := 0; i < width; i++ {
		fmt.Print("═")
	}
	fmt.Println()
}

func PrintBoxMiddle(width int) {
	for i := 0; i < width; i++ {
		fmt.Print("─")
	}
	fmt.Println()
}

func PrintBoxBottom(width int) {
	for i := 0; i < width; i++ {
		fmt.Print("═")
	}
	fmt.Println()
}

func Prompt(message string) string {
	fmt.Print(message)
	input, err := getReader().ReadString('\n')
	if err != nil {
		log.Printf("Error reading input: %v", err)
	}
	return strings.TrimSpace(input)
}

func Confirm(message string, defaultYes bool) bool {
	defaultStr := "[y/N]"
	if defaultYes {
		defaultStr = "[Y/n]"
	}

	prompt := fmt.Sprintf("%s %s: ", message, defaultStr)
	fmt.Print(prompt)

	input, _ := getReader().ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return defaultYes
	}

	return input == "y" || input == "yes"
}

func WaitForEnter(message string) {
	fmt.Print(message)
	getReader().ReadString('\n')
}

func PrintSuccess(message string) {
	fmt.Printf("✓ %s\n", message)
}

func PrintError(message string) {
	fmt.Printf("✗ %s\n", message)
}

func PrintInfo(message string) {
	fmt.Printf("→ %s\n", message)
}

func Truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	// Pad with spaces to align columns
	return s + strings.Repeat(" ", maxLen-len(s))
}

func CreateDirectoryWithSudo(path string) error {
	// First, try to create normally
	err := os.MkdirAll(path, 0755)
	if err == nil {
		return nil
	}

	// If it failed and we're on Linux, offer sudo
	if runtime.GOOS == "linux" && strings.Contains(err.Error(), "permission denied") {
		fmt.Println()
		PrintError("Ошибка: нет разрешения на создание папки")
		if Confirm("Попробовать с sudo?", false) {
			fmt.Println()
			password := PromptPassword("Введите пароль: ")

			cmd := exec.Command("sudo", "-S", "mkdir", "-p", path)
			cmd.Stdin = strings.NewReader(password + "\n")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err = cmd.Run()
			if err != nil {
				return fmt.Errorf("не удалось создать папку: %w", err)
			}

			// Change ownership to current user
			currentUser := os.Getenv("USER")
			if currentUser != "" {
				chownCmd := exec.Command("sudo", "-S", "chown", currentUser+":"+currentUser, path)
				chownCmd.Stdin = strings.NewReader(password + "\n")
				chownCmd.Run() // Ignore errors
			}

			return nil
		}
	}

	return err
}

// PromptPassword reads password input without echoing
func PromptPassword(message string) string {
	fmt.Print(message)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		// fallback to visible prompt
		p, _ := getReader().ReadString('\n')
		return strings.TrimSpace(p)
	}
	return strings.TrimSpace(string(bytePassword))
}

func BoolToString(b bool) string {
	if b {
		return "Да"
	}
	return "Нет"
}

func RestartApplication() {
	exePath, err := os.Executable()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка при перезапуске: %v", err))
		return
	}

	// Start new process and exit current one immediately to free resources
	err = exec.Command(exePath).Start()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка при перезапуске: %v", err))
		return
	}

	// Kill current process to free all resources
	os.Exit(0)
}
