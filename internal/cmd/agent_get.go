// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// agentGetTimeout is the maximum time to wait for the API response.
const agentGetTimeout = 30 * time.Second

// NewAgentGetCmd creates and returns the agent get command.
func NewAgentGetCmd() *cobra.Command {
	var agentID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details and status of an agent",
		Long: `Get details and current status of a specific agent.

Shows the agent's ID, name, description, status, hostname, IP address,
agent binary version, API key prefix, and last heartbeat time.

Examples:
  # Get agent details
  stackeye agent get --id <uuid>

  # Get in JSON format for scripting
  stackeye agent get --id <uuid> -o json`,
		Aliases: []string{"show", "status"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAgentGet(cmd.Context(), agentID)
		},
	}

	cmd.Flags().StringVarP(&agentID, "id", "i", "", "Agent UUID (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// runAgentGet executes the agent get command logic.
func runAgentGet(ctx context.Context, agentID string) error {
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, agentGetTimeout)
	defer cancel()

	response, err := client.GetAgent(reqCtx, apiClient, agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(response)
		}
	}

	printAgentDetail(&response.Data)
	return nil
}

// printAgentDetail formats and prints agent details in a human-friendly format.
func printAgentDetail(a *client.Agent) {
	status := agentStatus(a)

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                      AGENT DETAILS                         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Name:        %s\n", a.Name)
	fmt.Printf("  ID:          %s\n", a.ID)
	fmt.Printf("  Status:      %s\n", status)
	fmt.Printf("  Key Prefix:  %s…\n", a.KeyPrefix)

	if a.Description != nil {
		fmt.Printf("  Description: %s\n", *a.Description)
	}
	if a.Hostname != nil {
		fmt.Printf("  Hostname:    %s\n", *a.Hostname)
	}
	if a.IPAddress != nil {
		fmt.Printf("  IP Address:  %s\n", *a.IPAddress)
	}
	if a.AgentVersion != nil {
		fmt.Printf("  Version:     %s\n", *a.AgentVersion)
	}
	if a.LastSeenAt != nil {
		fmt.Printf("  Last Seen:   %s\n", formatAgentDate(*a.LastSeenAt))
	} else {
		fmt.Println("  Last Seen:   (never)")
	}

	fmt.Printf("  Registered:  %s\n", formatAgentDate(a.CreatedAt))
	fmt.Println()

	if !a.IsActive {
		fmt.Println("  ⚠ This agent is deactivated and will not authenticate.")
		fmt.Println()
	} else if status == "offline" {
		fmt.Println("  ⚠ Agent has not reported in over 5 minutes — check that the binary is running.")
		fmt.Println()
	} else if status == "pending" {
		fmt.Println("  ℹ Agent registered but not yet seen — start the binary with the API key.")
		fmt.Println()
	}
}
