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

// privateRegionGetTimeout is the maximum time to wait for the API response.
const privateRegionGetTimeout = 30 * time.Second

// NewPrivateRegionGetCmd creates and returns the private-region get command.
func NewPrivateRegionGetCmd() *cobra.Command {
	var regionID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of a private monitoring region",
		Long: `Get details of a specific private monitoring region by its ID.

Displays the full region configuration including location, status,
and timestamps.

Examples:
  # Get a private region by ID
  stackeye private-region get --id prv-nyc-office

  # Get in JSON format
  stackeye private-region get --id prv-nyc-office -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrivateRegionGet(cmd.Context(), regionID)
		},
	}

	cmd.Flags().StringVarP(&regionID, "id", "i", "", "Private region ID (e.g., prv-nyc-office) (required)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}

// runPrivateRegionGet executes the private-region get command logic.
func runPrivateRegionGet(ctx context.Context, regionID string) error {
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, privateRegionGetTimeout)
	defer cancel()

	response, err := client.GetPrivateRegion(reqCtx, apiClient, regionID)
	if err != nil {
		return fmt.Errorf("failed to get private region: %w", err)
	}

	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(response)
		}
	}

	printPrivateRegionDetail(&response.Data)
	return nil
}

// printPrivateRegionDetail formats and prints a single private region in detail.
func printPrivateRegionDetail(r *client.PrivateRegion) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                   PRIVATE REGION                           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  ID:           %s\n", r.ID)
	fmt.Printf("  Name:         %s\n", r.Name)
	fmt.Printf("  Display Name: %s\n", r.DisplayName)
	fmt.Printf("  Status:       %s\n", r.Status)
	fmt.Println()
	fmt.Println("  Location:")
	fmt.Printf("    Continent:    %s\n", r.Continent)
	fmt.Printf("    Country Code: %s\n", r.CountryCode)
	if r.City != nil {
		fmt.Printf("    City:         %s\n", *r.City)
	}
	fmt.Println()
	fmt.Printf("  Created: %s\n", formatPrivateRegionDate(r.CreatedAt))
	fmt.Printf("  Updated: %s\n", formatPrivateRegionDate(r.UpdatedAt))
	fmt.Println()
}
