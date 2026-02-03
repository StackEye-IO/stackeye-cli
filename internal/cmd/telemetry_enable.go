package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/StackEye-IO/stackeye-cli/internal/config"
	"github.com/StackEye-IO/stackeye-cli/internal/telemetry"
)

// NewTelemetryEnableCmd creates and returns the telemetry enable command.
func NewTelemetryEnableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable telemetry collection",
		Long: `Enable anonymous usage analytics for StackEye CLI.

When enabled, the CLI collects anonymous usage data to help improve
the product. No personal data, API keys, or error messages are collected.

Data collected:
  - Command names and exit codes
  - Execution duration
  - CLI version and platform (OS/arch)
  - Anonymized organization ID (hash prefix)

This setting is stored in your config file at ~/.config/stackeye/config.yaml.

Note: The STACKEYE_TELEMETRY environment variable can override this setting.

Examples:
  # Enable telemetry
  stackeye telemetry enable`,
		RunE: runTelemetryEnable,
	}

	return cmd
}

func runTelemetryEnable(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Preferences == nil {
		cfg.Preferences = config.NewPreferences()
	}

	cfg.Preferences.TelemetryEnabled = true
	cfg.Preferences.TelemetryPrompted = true

	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Update the global client
	telemetry.GetClient().SetEnabled(true)

	fmt.Println("Telemetry enabled. Thank you for helping improve StackEye!")
	return nil
}
