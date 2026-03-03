// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// agentRegisterTimeout is the maximum time to wait for the API response.
const agentRegisterTimeout = 30 * time.Second

// NewAgentRegisterCmd creates and returns the agent register command.
func NewAgentRegisterCmd() *cobra.Command {
	var name string
	var description string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new self-hosted agent",
		Long: `Register a new self-hosted agent and obtain its one-time API key.

The API key is returned exactly once — it cannot be retrieved after this
response. Copy it immediately and configure the agent binary with it.

Name rules:
  - 1 to 100 characters
  - Any printable characters

After registration, install and start the agent binary on your host:
  STACKEYE_API_KEY=<key> stackeye-agent

Examples:
  # Register a new agent
  stackeye agent register --name prod-web-01

  # Register with a description
  stackeye agent register --name prod-web-01 --description "Production web server"

  # Preview without creating
  stackeye agent register --name prod-web-01 --dry-run

  # Output in JSON format
  stackeye agent register --name prod-web-01 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var descPtr *string
			if cmd.Flags().Changed("description") && description != "" {
				descPtr = &description
			}
			return runAgentRegister(cmd.Context(), name, descPtr)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Agent name (1-100 chars) (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Optional description (max 1000 chars)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// runAgentRegister executes the agent register command logic.
func runAgentRegister(ctx context.Context, name string, description *string) error {
	if GetDryRun() {
		desc := "(none)"
		if description != nil {
			desc = *description
		}
		dryrun.PrintAction("register", "agent",
			"Name", name,
			"Description", desc,
		)
		return nil
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	req := client.RegisterAgentRequest{
		Name:        name,
		Description: description,
	}

	reqCtx, cancel := context.WithTimeout(ctx, agentRegisterTimeout)
	defer cancel()

	response, err := client.RegisterAgent(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(response)
		}
	}

	printAgentRegistered(response)
	return nil
}

// printAgentRegistered formats and prints a newly registered agent with its one-time key.
func printAgentRegistered(response *client.AgentRegisterResponse) {
	a := response.Data.Agent

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    AGENT REGISTERED                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Name: %s\n", a.Name)
	fmt.Printf("  ID:   %s\n", a.ID)
	fmt.Println()

	fmt.Println("  API Key (save this now — shown only once):")
	fmt.Println("  ┌──────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("    %s\n", response.Data.APIKey)
	fmt.Println("  └──────────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  ⚠ WARNING: Copy this key now — it cannot be retrieved later!")
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Printf("    Install the agent binary on %s and start it with:\n", a.Name)
	fmt.Println("    STACKEYE_API_KEY=<key> stackeye-agent")
	fmt.Println()
}
