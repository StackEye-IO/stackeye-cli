// Package clipboard provides cross-platform clipboard copy utilities for the
// StackEye CLI. It supports macOS (pbcopy), Linux (xclip or xsel), and
// Windows (clip.exe), with a graceful fallback that prints the text when the
// clipboard is not available.
//
// Usage:
//
//	if err := clipboard.Copy("se_abc123..."); err != nil {
//	    // handle error
//	}
//
//	// Or with automatic fallback to printing the text:
//	clipboard.CopyOrPrint("se_abc123...")
package clipboard

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ErrEmptyText is returned when an empty string is passed to Copy.
var ErrEmptyText = fmt.Errorf("clipboard: text must not be empty")

// copier is the function used to copy text to the system clipboard.
// Override this in tests via SetCopier to avoid interacting with the
// real clipboard.
var copier = defaultCopy

// SetCopier replaces the function used to copy text to the clipboard.
// Pass nil to restore the default platform-specific behaviour. Intended
// for testing only.
func SetCopier(fn func(string) error) {
	if fn == nil {
		copier = defaultCopy
		return
	}
	copier = fn
}

// Copy copies the given text to the system clipboard. It returns an error
// if the text is empty, the platform is unsupported, or the clipboard
// command fails.
func Copy(text string) error {
	if err := validateText(text); err != nil {
		return err
	}
	return copier(text)
}

// CopyOrPrint attempts to copy text to the clipboard. If the clipboard
// is unavailable or the copy fails, it prints the text to stderr so
// the user can copy it manually. This never returns an error for
// clipboard failures — only for invalid input.
func CopyOrPrint(text string) error {
	if err := validateText(text); err != nil {
		return err
	}

	if err := copier(text); err != nil {
		fmt.Fprintln(os.Stderr, "Could not copy to clipboard automatically.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Please copy the following value manually:")
		fmt.Fprintf(os.Stderr, "  %s\n", text)
		fmt.Fprintln(os.Stderr)
	}

	return nil
}

// validateText checks that text is non-empty.
func validateText(text string) error {
	if strings.TrimSpace(text) == "" {
		return ErrEmptyText
	}
	return nil
}

// defaultCopy writes text to the system clipboard using the platform's
// native clipboard command.
func defaultCopy(text string) error {
	var name string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		name = "pbcopy"
	case "linux":
		name, args = linuxClipboardCmd()
		if name == "" {
			return fmt.Errorf("clipboard: no clipboard tool found (install xclip or xsel)")
		}
	case "windows":
		name = "clip"
	default:
		return fmt.Errorf("clipboard: unsupported platform: %s", runtime.GOOS)
	}

	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// linuxClipboardCmd returns the command name and args for the first
// available Linux clipboard tool. It prefers xclip over xsel.
func linuxClipboardCmd() (string, []string) {
	if path, err := exec.LookPath("xclip"); err == nil && path != "" {
		return "xclip", []string{"-selection", "clipboard"}
	}
	if path, err := exec.LookPath("xsel"); err == nil && path != "" {
		return "xsel", []string{"--clipboard", "--input"}
	}
	return "", nil
}
