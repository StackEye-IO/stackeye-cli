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

// agentListTimeout is the maximum time to wait for the API response.
const agentListTimeout = 30 * time.Second

// NewAgentListCmd creates and returns the agent list command.
func NewAgentListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all registered agents",
		Long: `List all active (non-deactivated) agents for the current organization.

Displays each agent's ID, name, status, hostname, version, and last-seen time.
Agents are ordered by name.

Examples:
  # List all agents in table format (default)
  stackeye agent list

  # List in JSON format for scripting
  stackeye agent list -o json

  # List in YAML format
  stackeye agent list -o yaml`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAgentList(cmd.Context())
		},
	}

	return cmd
}

// runAgentList executes the agent list command logic.
func runAgentList(ctx context.Context) error {
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, agentListTimeout)
	defer cancel()

	response, err := client.ListAgents(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to list agents: %w", err)
	}

	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(response)
		}
	}

	printAgentList(response)
	return nil
}

// printAgentList formats and prints the agent list in a human-friendly format.
func printAgentList(response *client.AgentListResponse) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                        AGENTS                              ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(response.Data) == 0 {
		fmt.Println("  No agents registered.")
		fmt.Println()
		fmt.Println("  Register one with: stackeye agent register --name <name>")
		fmt.Println()
		return
	}

	fmt.Printf("  Total: %d agent(s)\n\n", response.Meta.Total)

	fmt.Println("  ┌─────────────────────────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("  │ %-36s │ %-20s │ %-8s │ %-15s │ %-12s │\n",
		"ID", "NAME", "STATUS", "HOSTNAME", "LAST SEEN")
	fmt.Println("  ├─────────────────────────────────────────────────────────────────────────────────────────────────────────┤")

	for _, a := range response.Data {
		status := agentStatus(&a)
		hostname := "(unknown)"
		if a.Hostname != nil {
			hostname = *a.Hostname
		}
		lastSeen := "(never)"
		if a.LastSeenAt != nil {
			lastSeen = formatAgentDate(*a.LastSeenAt)
		}

		fmt.Printf("  │ %-36s │ %-20s │ %-8s │ %-15s │ %-12s │\n",
			a.ID,
			truncateAgentField(a.Name, 20),
			status,
			truncateAgentField(hostname, 15),
			truncateAgentField(lastSeen, 12),
		)
	}

	fmt.Println("  └─────────────────────────────────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
}

// agentStatus derives the display status string from agent fields.
// Matches the 4-state logic in AgentStatusBadge on the web UI.
func agentStatus(a *client.Agent) string {
	if !a.IsActive {
		return "inactive"
	}
	if a.LastSeenAt == nil {
		return "pending"
	}
	t, err := time.Parse(time.RFC3339, *a.LastSeenAt)
	if err != nil {
		return "unknown"
	}
	if time.Since(t) > 5*time.Minute {
		return "offline"
	}
	return "online"
}

// truncateAgentField truncates a string to fit within maxLen characters.
func truncateAgentField(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatAgentDate formats an ISO 8601 timestamp for compact display.
func formatAgentDate(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	return t.Format("Jan 02 15:04")
}
