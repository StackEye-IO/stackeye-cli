package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/StackEye-IO/stackeye-cli/internal/config"
	"github.com/StackEye-IO/stackeye-cli/internal/telemetry"
)

// NewTelemetryDisableCmd creates and returns the telemetry disable command.
func NewTelemetryDisableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable telemetry collection",
		Long: `Disable anonymous usage analytics for StackEye CLI.

When disabled, no usage data will be collected or transmitted.
This setting is stored in your config file at ~/.config/stackeye/config.yaml.

You can also set the STACKEYE_TELEMETRY=0 environment variable to
disable telemetry regardless of the config file setting.

Examples:
  # Disable telemetry
  stackeye telemetry disable

  # Disable via environment variable
  export STACKEYE_TELEMETRY=0`,
		RunE: runTelemetryDisable,
	}

	return cmd
}

func runTelemetryDisable(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Preferences == nil {
		cfg.Preferences = config.NewPreferences()
	}

	cfg.Preferences.TelemetryEnabled = false
	cfg.Preferences.TelemetryPrompted = true

	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Update the global client
	telemetry.GetClient().SetEnabled(false)

	fmt.Println("Telemetry disabled. No data will be collected.")
	return nil
}
