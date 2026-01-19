// Package cmd implements the CLI commands for StackEye.
//
// This package contains the root command and all subcommands for the
// StackEye CLI tool. Commands are organized hierarchically using Cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the base command for the CLI.
var rootCmd = &cobra.Command{
	Use:   "stackeye",
	Short: "StackEye CLI - Eye on your stack",
	Long: `StackEye CLI provides command-line access to the StackEye uptime monitoring platform.

Manage probes, alerts, notification channels, and organizations directly from
your terminal. Integrate monitoring into your CI/CD pipelines and automation
workflows.

Get started:
  stackeye login              Authenticate with your StackEye account
  stackeye probe list         List all monitoring probes
  stackeye alert list         View current alerts

For more information about a command:
  stackeye [command] --help`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command and returns any error.
// This is called by main.main() and handles command execution.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
