// Package browser provides cross-platform browser-opening utilities for the
// StackEye CLI. It supports macOS (open), Linux (xdg-open), and Windows
// (rundll32), with a graceful fallback that prints the URL when the browser
// cannot be opened.
//
// Usage:
//
//	if err := browser.Open("https://app.stackeye.io"); err != nil {
//	    // handle error
//	}
//
//	// Or with automatic fallback to printing the URL:
//	browser.OpenOrPrint("https://app.stackeye.io")
package browser

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ErrEmptyURL is returned when an empty URL is passed to Open.
var ErrEmptyURL = fmt.Errorf("browser: URL must not be empty")

// opener is the function used to launch a URL in the default browser.
// Override this in tests via SetOpener to avoid launching real browsers.
var opener = defaultOpen

// SetOpener replaces the function used to launch URLs. Pass nil to restore
// the default platform-specific behaviour. Intended for testing only.
func SetOpener(fn func(string) error) {
	if fn == nil {
		opener = defaultOpen
		return
	}
	opener = fn
}

// Open opens the given URL in the user's default browser. It validates the
// URL and returns an error if the URL is empty, unparseable, has a
// disallowed scheme, or the browser command fails to start.
// Only http and https URLs are permitted.
func Open(rawURL string) error {
	if err := validateURL(rawURL); err != nil {
		return err
	}
	return opener(rawURL)
}

// OpenOrPrint attempts to open the URL in a browser. If the browser fails
// to open, it prints the URL to stderr so the user can open it manually.
// This never returns an error for browser launch failures - only for
// invalid input. Only http and https URLs are permitted.
func OpenOrPrint(rawURL string) error {
	if err := validateURL(rawURL); err != nil {
		return err
	}

	if err := opener(rawURL); err != nil {
		fmt.Fprintln(os.Stderr, "Could not open browser automatically.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Please open this URL manually:")
		fmt.Fprintf(os.Stderr, "  %s\n", rawURL)
		fmt.Fprintln(os.Stderr)
	}

	return nil
}

// validateURL checks that rawURL is a non-empty, parseable URL with an
// http or https scheme.
func validateURL(rawURL string) error {
	if rawURL == "" {
		return ErrEmptyURL
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("browser: invalid URL %q: %w", rawURL, err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("browser: unsupported URL scheme %q (only http and https are allowed)", u.Scheme)
	}

	return nil
}

// defaultOpen launches the URL using the platform's default browser command.
func defaultOpen(rawURL string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "linux":
		cmd = exec.Command("xdg-open", rawURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		return fmt.Errorf("browser: unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
