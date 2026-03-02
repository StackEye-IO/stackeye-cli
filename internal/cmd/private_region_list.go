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

// privateRegionListTimeout is the maximum time to wait for the API response.
const privateRegionListTimeout = 30 * time.Second

// NewPrivateRegionListCmd creates and returns the private-region list command.
func NewPrivateRegionListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all private monitoring regions",
		Long: `List all private monitoring regions for the current organization.

Displays each region's ID, display name, location, and status.
Regions are ordered by creation time (oldest first).

Examples:
  # List all private regions in table format (default)
  stackeye private-region list

  # List in JSON format for scripting
  stackeye private-region list -o json

  # List in YAML format
  stackeye private-region list -o yaml`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrivateRegionList(cmd.Context())
		},
	}

	return cmd
}

// runPrivateRegionList executes the private-region list command logic.
func runPrivateRegionList(ctx context.Context) error {
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, privateRegionListTimeout)
	defer cancel()

	response, err := client.ListPrivateRegions(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to list private regions: %w", err)
	}

	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(response)
		}
	}

	printPrivateRegionList(response)
	return nil
}

// printPrivateRegionList formats and prints the private region list in a human-friendly format.
func printPrivateRegionList(response *client.PrivateRegionListResponse) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                   PRIVATE REGIONS                          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(response.Data) == 0 {
		fmt.Println("  No private regions found.")
		fmt.Println()
		fmt.Println("  Create one with: stackeye private-region create --slug <slug> \\")
		fmt.Println("      --display-name <name> --continent <continent> --country-code <code>")
		fmt.Println()
		return
	}

	fmt.Printf("  Total: %d region(s)\n\n", response.Meta.Total)

	fmt.Println("  ┌────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("  │ %-20s │ %-20s │ %-15s │ %-6s │ %-8s │\n",
		"ID", "DISPLAY NAME", "LOCATION", "CODE", "STATUS")
	fmt.Println("  ├────────────────────────────────────────────────────────────────────────────────┤")

	for _, r := range response.Data {
		location := r.Continent
		if r.City != nil && *r.City != "" {
			location = *r.City + ", " + r.CountryCode
		}

		fmt.Printf("  │ %-20s │ %-20s │ %-15s │ %-6s │ %-8s │\n",
			truncatePrivateRegionField(r.ID, 20),
			truncatePrivateRegionField(r.DisplayName, 20),
			truncatePrivateRegionField(location, 15),
			truncatePrivateRegionField(r.CountryCode, 6),
			r.Status)
	}

	fmt.Println("  └────────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
}

// truncatePrivateRegionField truncates a string to fit within maxLen characters.
func truncatePrivateRegionField(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatPrivateRegionDate formats an ISO 8601 timestamp for compact display.
func formatPrivateRegionDate(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	return t.Format("Jan 02, 2006")
}
