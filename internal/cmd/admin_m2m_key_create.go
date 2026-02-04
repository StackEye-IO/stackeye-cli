// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/dryrun"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client/admin"
	"github.com/spf13/cobra"
)

// adminM2MKeyCreateTimeout is the maximum time to wait for the API response.
const adminM2MKeyCreateTimeout = 30 * time.Second

// NewAdminM2MKeyCreateCmd creates and returns the m2m-key create command.
func NewAdminM2MKeyCreateCmd() *cobra.Command {
	var region string
	var global bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new M2M key",
		Long: `Create a new machine-to-machine (M2M) authentication key.

Creates either a regional or global M2M key. You must specify either
--region or --global (but not both).

Regional keys are scoped to a specific monitoring region (e.g., nyc3, ams3)
and can only be used by controllers in that region.

Global keys have no region restriction and can be used across all regions.

IMPORTANT: The plaintext key is only displayed once at creation time.
Store it securely as it cannot be retrieved later.

Key formats:
  - Regional: se_m2m_<region>_<64_hex_chars>
  - Global:   se_m2m_<64_hex_chars>

Examples:
  # Create a regional M2M key for NYC3
  stackeye admin m2m-key create --region nyc3

  # Create a global M2M key
  stackeye admin m2m-key create --global

  # Output in JSON format
  stackeye admin m2m-key create --global -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdminM2MKeyCreate(cmd.Context(), region, global)
		},
	}

	cmd.Flags().StringVarP(&region, "region", "r", "", "Region for the key (e.g., nyc3)")
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Create a global key (no region)")

	return cmd
}

// runAdminM2MKeyCreate executes the m2m-key create command logic.
func runAdminM2MKeyCreate(ctx context.Context, region string, global bool) error {
	// Validate flags
	if region == "" && !global {
		return fmt.Errorf("must specify either --region or --global")
	}
	if region != "" && global {
		return fmt.Errorf("cannot specify both --region and --global")
	}

	// Dry-run check: after validation, before API calls
	if GetDryRun() {
		keyType := "global"
		if region != "" {
			keyType = fmt.Sprintf("regional (%s)", region)
		}
		dryrun.PrintAction("create", "M2M key",
			"Type", keyType,
		)
		return nil
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build request
	req := admin.CreateM2MKeyRequest{
		Region: region, // Empty string for global keys
	}

	// Call SDK to create M2M key with timeout
	reqCtx, cancel := context.WithTimeout(ctx, adminM2MKeyCreateTimeout)
	defer cancel()

	response, err := admin.CreateM2MKey(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create M2M key: %w", err)
	}

	// Check output format - use JSON/YAML if requested, otherwise pretty print
	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(response)
		}
	}

	// Pretty print for table format (default)
	printM2MKeyCreated(response, region)
	return nil
}

// printM2MKeyCreated formats and prints the newly created M2M key.
func printM2MKeyCreated(response *admin.CreateM2MKeyResponse, region string) {
	keyType := "global"
	if region != "" {
		keyType = fmt.Sprintf("regional (%s)", region)
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                   M2M KEY CREATED                          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Type:   %s\n", keyType)
	fmt.Printf("  ID:     %s\n", response.Data.ID)
	fmt.Printf("  Prefix: %s\n", response.Data.KeyPrefix)
	fmt.Println()
	fmt.Println("  Plaintext Key:")
	fmt.Println("  ┌────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("    %s\n", response.Data.Key)
	fmt.Println("  └────────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("  ⚠ WARNING: Save this key now - it cannot be retrieved later!")
	fmt.Println()
}
