// Package update provides update checking and notification functionality for the CLI.
// It wraps the stackeye-go-sdk update package to provide CLI-specific features
// like colored output and background checking during command execution.
package update

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	sdkupdate "github.com/StackEye-IO/stackeye-go-sdk/update"
)

const (
	// GitHubRepo is the repository to check for updates.
	GitHubRepo = "stackeye-io/stackeye-cli"

	// BackgroundTimeout is how long to wait for a background check to complete
	// when printing the notification.
	BackgroundTimeout = 500 * time.Millisecond
)

// Notifier manages update checks and notifications for the CLI.
// It wraps the SDK's Checker to provide CLI-specific functionality.
type Notifier struct {
	checker        *sdkupdate.Checker
	currentVersion string
	writer         io.Writer
	colorEnabled   bool
}

// NotifierOption is a functional option for configuring the Notifier.
type NotifierOption func(*Notifier)

// WithWriter sets the output writer for notifications.
func WithWriter(w io.Writer) NotifierOption {
	return func(n *Notifier) {
		n.writer = w
	}
}

// WithColor enables or disables colored output.
func WithColor(enabled bool) NotifierOption {
	return func(n *Notifier) {
		n.colorEnabled = enabled
	}
}

// WithCacheDir sets the directory for storing the update check cache.
func WithCacheDir(dir string) NotifierOption {
	return func(n *Notifier) {
		// Recreate checker with new cache dir if one exists
		if n.checker != nil {
			updater := sdkupdate.NewUpdater(GitHubRepo, n.currentVersion)
			n.checker = sdkupdate.NewChecker(updater, sdkupdate.WithCacheDir(dir))
		}
	}
}

// NewNotifier creates a new update notifier for the CLI.
//
// The currentVersion should be the version currently installed (e.g., "1.0.0").
// Pass version.Version from the internal/version package.
//
// Example:
//
//	notifier := update.NewNotifier(version.Version)
//	notifier.StartCheck(ctx)
//	// ... run command ...
//	notifier.PrintNotification()
func NewNotifier(currentVersion string, opts ...NotifierOption) *Notifier {
	updater := sdkupdate.NewUpdater(GitHubRepo, currentVersion)
	checker := sdkupdate.NewChecker(updater)

	n := &Notifier{
		checker:        checker,
		currentVersion: currentVersion,
		writer:         os.Stderr,
		colorEnabled:   true,
	}

	for _, opt := range opts {
		opt(n)
	}

	return n
}

// StartCheck initiates a non-blocking update check in the background.
// Call this early in command execution, then call PrintNotification() after
// the command completes to show any available updates.
func (n *Notifier) StartCheck(ctx context.Context) {
	n.checker.CheckInBackground(ctx)
}

// PrintNotification displays an update notification if an update is available.
// This waits briefly for the background check to complete, then prints
// a notification if an update is found.
//
// Returns true if a notification was printed, false otherwise.
func (n *Notifier) PrintNotification() bool {
	result, err := n.checker.WaitForBackgroundResult(BackgroundTimeout)
	if err != nil || result == nil || !result.HasUpdate {
		return false
	}

	n.printUpdateMessage(result)
	return true
}

// GetResult returns the result of the background check without printing.
// This is useful for testing or when custom notification handling is needed.
func (n *Notifier) GetResult() (*sdkupdate.CachedCheckResult, error) {
	return n.checker.WaitForBackgroundResult(BackgroundTimeout)
}

// printUpdateMessage formats and prints the update notification.
func (n *Notifier) printUpdateMessage(result *sdkupdate.CachedCheckResult) {
	fmt.Fprintln(n.writer)

	if n.colorEnabled {
		n.printColoredMessage(result)
	} else {
		n.printPlainMessage(result)
	}
}

// printColoredMessage prints the update notification with ANSI colors.
func (n *Notifier) printColoredMessage(result *sdkupdate.CachedCheckResult) {
	// ANSI color codes
	const (
		yellow = "\033[33m"
		cyan   = "\033[36m"
		bold   = "\033[1m"
		reset  = "\033[0m"
	)

	fmt.Fprintf(n.writer, "%s%sA new version of stackeye is available!%s %s%s%s -> %s%s%s\n",
		bold, yellow, reset,
		cyan, n.currentVersion, reset,
		bold, result.LatestVersion, reset,
	)
	fmt.Fprintf(n.writer, "Run %sstackeye upgrade%s to update.\n", bold, reset)

	if result.ReleaseURL != "" {
		fmt.Fprintf(n.writer, "Release notes: %s%s%s\n", cyan, result.ReleaseURL, reset)
	}
}

// printPlainMessage prints the update notification without colors.
func (n *Notifier) printPlainMessage(result *sdkupdate.CachedCheckResult) {
	fmt.Fprintf(n.writer, "A new version of stackeye is available! %s -> %s\n",
		n.currentVersion, result.LatestVersion,
	)
	fmt.Fprintln(n.writer, "Run 'stackeye upgrade' to update.")

	if result.ReleaseURL != "" {
		fmt.Fprintf(n.writer, "Release notes: %s\n", result.ReleaseURL)
	}
}

// ShouldCheck returns true if update checking should be performed.
// This is false for dev builds or when explicitly disabled.
func ShouldCheck(currentVersion string, disabled bool) bool {
	if disabled {
		return false
	}

	// Skip for dev builds
	if currentVersion == "dev" || currentVersion == "" {
		return false
	}

	return true
}
