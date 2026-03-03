// Package cmd implements the CLI commands for StackEye.
package cmd

import "github.com/spf13/cobra"

// NewAgentCmd creates and returns the agent parent command.
// This command provides access to self-hosted agent lifecycle operations.
//
// Task #10547: Add CLI agent register/list/status commands (F-841)
func NewAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage self-hosted agents",
		Long: `Manage self-hosted agents that collect host telemetry.

Agents are small binaries installed on your servers. They authenticate
with a per-agent API key and ship host metrics (CPU, memory, disk,
network, temperature, power) to StackEye.

Commands:
  list      List all registered agents
  get       Get details of a specific agent
  register  Register a new agent and obtain its API key

Examples:
  # List all agents
  stackeye agent list

  # Register a new agent
  stackeye agent register --name prod-web-01 --description "Production web server"

  # Get agent details and status
  stackeye agent get --id <uuid>

  # Machine-readable output
  stackeye agent list -o json`,
		Aliases: []string{"agents"},
	}

	// Register subcommands
	cmd.AddCommand(NewAgentListCmd())
	cmd.AddCommand(NewAgentGetCmd())
	cmd.AddCommand(NewAgentRegisterCmd())

	return cmd
}
