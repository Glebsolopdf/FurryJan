package ui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"golang.org/x/term"
)

var (
	globalReader *bufio.Reader
	readerSource *os.File
	readerMu     sync.Mutex
)

const exitInputSentinel = ".exit"

func IsExitInput(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), exitInputSentinel)
}

func getReader() *bufio.Reader {
	readerMu.Lock()
	defer readerMu.Unlock()

	if globalReader != nil {
		return globalReader
	}

	initInputReaderLocked()
	return globalReader
}

func initInputReaderLocked() {
	readerSource = os.Stdin
	globalReader = bufio.NewReader(readerSource)
}

func resetInputReader() {
	readerMu.Lock()
	defer readerMu.Unlock()

	if readerSource != nil && readerSource != os.Stdin {
		_ = readerSource.Close()
	}

	tty, err := os.Open("/dev/tty")
	if err == nil {
		readerSource = tty
		globalReader = bufio.NewReader(readerSource)
		return
	}

	initInputReaderLocked()
}

func readInputLine() (string, error) {
	input, err := getReader().ReadString('\n')
	if err == nil {
		return input, nil
	}

	// Retry once with a refreshed reader, preferring /dev/tty when available.
	resetInputReader()
	return getReader().ReadString('\n')
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
	input, err := readInputLine()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			log.Printf("Error reading input: %v", err)
		}
		return exitInputSentinel
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

	input, err := readInputLine()
	if err != nil {
		return defaultYes
	}
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return defaultYes
	}

	return input == "y" || input == "yes"
}

func WaitForEnter(message string) {
	fmt.Print(message)
	_, _ = readInputLine()
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

// PromptPassword reads password input without echoing
func PromptPassword(message string) string {
	fmt.Print(message)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		// fallback to visible prompt
		p, _ := readInputLine()
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

func RestartApplication(_ context.Context) error {
	return ErrRestartRequested
}
