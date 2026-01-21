// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewRegionCmd creates and returns the region parent command.
// This command provides information about monitoring regions.
func NewRegionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "region",
		Short: "Manage monitoring regions",
		Long: `Manage monitoring regions for StackEye probes.

Regions are geographic locations where StackEye runs probe checks. Each probe
can be configured to run from multiple regions, providing geographic coverage
and redundancy for uptime monitoring.

StackEye maintains monitoring infrastructure across multiple global regions
including North America, Europe, Asia-Pacific, and South America. When a probe
is created, you can select which regions should perform the health checks.

Key Concepts:
  - Each region has a unique code (e.g., us-east, eu-west, ap-south)
  - Probes can run from one or more regions simultaneously
  - Results from multiple regions provide consensus-based status detection
  - Regional diversity helps detect geo-specific outages

Use 'stackeye region [command] --help' for information about available subcommands.`,
		Aliases: []string{"regions"},
	}

	// Register subcommands
	cmd.AddCommand(NewRegionListCmd())

	return cmd
}
