package cmd

import "github.com/spf13/cobra"

// NewTelemetryCmd creates and returns the telemetry parent command.
// This command provides access to telemetry configuration.
func NewTelemetryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "telemetry",
		Short: "Manage anonymous usage analytics",
		Long: `Manage anonymous usage analytics for StackEye CLI.

Telemetry helps us improve the CLI by collecting anonymous usage data.
No personal data, API keys, or error messages are collected.

Data collected (when enabled):
  - Command names and exit codes
  - Execution duration
  - CLI version and platform info
  - Anonymized organization ID

Commands:
  status   Show current telemetry status
  enable   Enable telemetry collection
  disable  Disable telemetry collection

Examples:
  # Check current telemetry status
  stackeye telemetry status

  # Enable telemetry
  stackeye telemetry enable

  # Disable telemetry
  stackeye telemetry disable`,
		Aliases: []string{"tele"},
	}

	// Register subcommands
	cmd.AddCommand(NewTelemetryStatusCmd())
	cmd.AddCommand(NewTelemetryEnableCmd())
	cmd.AddCommand(NewTelemetryDisableCmd())

	return cmd
}
