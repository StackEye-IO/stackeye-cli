package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/StackEye-IO/stackeye-cli/internal/config"
	"github.com/StackEye-IO/stackeye-cli/internal/telemetry"
)

// NewTelemetryStatusCmd creates and returns the telemetry status command.
func NewTelemetryStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current telemetry status",
		Long: `Show the current telemetry status and configuration.

Displays whether telemetry is enabled or disabled, and shows
any environment variable overrides that may be in effect.

Examples:
  # Check telemetry status
  stackeye telemetry status`,
		RunE: runTelemetryStatus,
	}

	return cmd
}

func runTelemetryStatus(_ *cobra.Command, _ []string) error {
	client := telemetry.GetClient()

	// Check environment override
	envOverride := os.Getenv(telemetry.EnvTelemetry)
	hasEnvOverride := envOverride != ""

	// Load config to show stored preference
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var configEnabled bool
	var prompted bool
	if cfg.Preferences != nil {
		configEnabled = cfg.Preferences.TelemetryEnabled
		prompted = cfg.Preferences.TelemetryPrompted
	}

	// Show status
	if client.IsEnabled() {
		fmt.Println("Telemetry is enabled")
	} else {
		fmt.Println("Telemetry is disabled")
	}

	fmt.Println()

	// Show details
	fmt.Printf("Config setting:    %s\n", boolToStatus(configEnabled))
	fmt.Printf("Consent prompted:  %s\n", boolToYesNo(prompted))

	if hasEnvOverride {
		fmt.Printf("Environment override (%s): %s\n", telemetry.EnvTelemetry, envOverride)
	}

	fmt.Println()
	fmt.Println("Use 'stackeye telemetry enable' or 'stackeye telemetry disable' to change.")

	return nil
}

func boolToStatus(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
