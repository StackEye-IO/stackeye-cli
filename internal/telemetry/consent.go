package telemetry

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/StackEye-IO/stackeye-cli/internal/config"
)

// ConsentMessage is the message shown to users when prompting for telemetry consent.
const ConsentMessage = `StackEye would like to collect anonymous usage data to improve the CLI.

This includes: command usage frequency, error rates, and feature adoption.
No personal data or API keys are collected.

`

// ConsentPrompt is the prompt shown after the consent message.
const ConsentPrompt = "Enable telemetry? [y/N]: "

// CheckAndPromptConsent checks if the user has been prompted for telemetry consent.
// If not, it prompts them and saves their preference.
// Returns true if telemetry is enabled after the check.
//
// This should be called early in CLI initialization, but only for interactive sessions.
// Non-interactive sessions (piped input, CI/CD) should skip the prompt.
func CheckAndPromptConsent(stdin io.Reader, stdout io.Writer) (bool, error) {
	cfg, err := config.Load()
	if err != nil {
		// Can't load config, skip telemetry
		return false, nil
	}

	// Check if already prompted
	if cfg.Preferences != nil && cfg.Preferences.TelemetryPrompted {
		return cfg.Preferences.TelemetryEnabled, nil
	}

	// Check if running non-interactively
	if !isInteractive() {
		return false, nil
	}

	// Prompt for consent
	enabled, err := promptConsent(stdin, stdout)
	if err != nil {
		// Error prompting, default to disabled
		return false, nil
	}

	// Save preference
	if cfg.Preferences == nil {
		cfg.Preferences = config.NewPreferences()
	}
	cfg.Preferences.TelemetryEnabled = enabled
	cfg.Preferences.TelemetryPrompted = true

	if err := config.Save(); err != nil {
		// Failed to save, but still honor the user's choice for this session
		return enabled, nil
	}

	// Update the global client
	GetClient().SetEnabled(enabled)

	return enabled, nil
}

// promptConsent shows the consent prompt and returns the user's choice.
func promptConsent(stdin io.Reader, stdout io.Writer) (bool, error) {
	fmt.Fprint(stdout, ConsentMessage)
	fmt.Fprint(stdout, ConsentPrompt)

	reader := bufio.NewReader(stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))

	switch response {
	case "y", "yes":
		fmt.Fprintln(stdout, "Telemetry enabled. Thank you for helping improve StackEye!")
		return true, nil
	default:
		fmt.Fprintln(stdout, "Telemetry disabled. No data will be collected.")
		return false, nil
	}
}

// isInteractive returns true if stdin appears to be a terminal.
func isInteractive() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	// Check if it's a character device (terminal)
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// MarkPrompted marks the user as having been prompted for telemetry consent,
// without changing their telemetry preference.
// This is useful when the user explicitly enables/disables via command.
func MarkPrompted() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if cfg.Preferences == nil {
		cfg.Preferences = config.NewPreferences()
	}
	cfg.Preferences.TelemetryPrompted = true

	return config.Save()
}
