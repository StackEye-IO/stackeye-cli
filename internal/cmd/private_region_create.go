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

// privateRegionCreateTimeout is the maximum time to wait for the API response.
const privateRegionCreateTimeout = 30 * time.Second

// NewPrivateRegionCreateCmd creates and returns the private-region create command.
func NewPrivateRegionCreateCmd() *cobra.Command {
	var slug string
	var displayName string
	var continent string
	var countryCode string
	var city string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new private monitoring region",
		Long: `Create a new private monitoring region for the current organization.

The region ID is derived as prv-{slug} and must be globally unique.
After creation, generate a bootstrap key via the web UI or CLI before
the appliance can connect.

Slug rules:
  - 1 to 15 characters
  - Lowercase letters, digits, and hyphens only
  - The resulting region ID will be prv-{slug}

Examples:
  # Create a private region (minimal)
  stackeye private-region create --slug nyc-office \
      --display-name "NYC Office" \
      --continent "North America" \
      --country-code US

  # Create with optional city
  stackeye private-region create --slug lon-dc \
      --display-name "London DC" \
      --continent "Europe" \
      --country-code GB \
      --city "London"

  # Preview without creating
  stackeye private-region create --slug nyc-office \
      --display-name "NYC Office" \
      --continent "North America" \
      --country-code US \
      --dry-run`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var cityPtr *string
			if city != "" {
				cityPtr = &city
			}
			return runPrivateRegionCreate(cmd.Context(), slug, displayName, continent, countryCode, cityPtr)
		},
	}

	cmd.Flags().StringVarP(&slug, "slug", "s", "", "Unique slug (1-15 chars, e.g., nyc-office) — becomes prv-{slug} (required)")
	cmd.Flags().StringVarP(&displayName, "display-name", "n", "", "Human-readable display name (1-50 chars) (required)")
	cmd.Flags().StringVarP(&continent, "continent", "C", "", "Geographic continent (e.g., \"North America\") (required)")
	cmd.Flags().StringVarP(&countryCode, "country-code", "c", "", "ISO 3166-1 alpha-2 country code (e.g., US) (required)")
	cmd.Flags().StringVar(&city, "city", "", "Optional city name")

	_ = cmd.MarkFlagRequired("slug")
	_ = cmd.MarkFlagRequired("display-name")
	_ = cmd.MarkFlagRequired("continent")
	_ = cmd.MarkFlagRequired("country-code")

	return cmd
}

// runPrivateRegionCreate executes the private-region create command logic.
func runPrivateRegionCreate(ctx context.Context, slug, displayName, continent, countryCode string, city *string) error {
	if GetDryRun() {
		cityStr := "(none)"
		if city != nil {
			cityStr = *city
		}
		dryrun.PrintAction("create", "private region",
			"Slug", slug,
			"Display Name", displayName,
			"Continent", continent,
			"Country Code", countryCode,
			"City", cityStr,
		)
		return nil
	}

	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	req := client.CreatePrivateRegionRequest{
		Slug:        slug,
		DisplayName: displayName,
		Continent:   continent,
		CountryCode: countryCode,
		City:        city,
	}

	reqCtx, cancel := context.WithTimeout(ctx, privateRegionCreateTimeout)
	defer cancel()

	response, err := client.CreatePrivateRegion(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create private region: %w", err)
	}

	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(response)
		}
	}

	printPrivateRegionCreated(&response.Data)
	return nil
}

// printPrivateRegionCreated formats and prints a newly created private region.
func printPrivateRegionCreated(r *client.PrivateRegion) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║               PRIVATE REGION CREATED                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  ID:           %s\n", r.ID)
	fmt.Printf("  Display Name: %s\n", r.DisplayName)
	fmt.Printf("  Status:       %s\n", r.Status)
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Printf("    Generate a bootstrap key: stackeye private-region rotate --id %s\n", r.ID)
	fmt.Println("    Install the key on your appliance before it can connect.")
	fmt.Println()
}
