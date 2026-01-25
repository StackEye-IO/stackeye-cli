// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client/admin"
	"github.com/spf13/cobra"
)

// adminWorkerKeyCreateTimeout is the maximum time to wait for the API response.
const adminWorkerKeyCreateTimeout = 30 * time.Second

// workerKeyCreateRegion is the region flag value.
var workerKeyCreateRegion string

// workerKeyCreateName is the optional name flag value.
var workerKeyCreateName string

// NewAdminWorkerKeyCreateCmd creates and returns the worker-key create command.
func NewAdminWorkerKeyCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new worker key",
		Long: `Create a new worker authentication key for a regional probe.

This command creates a new worker key for authenticating regional probe
workers with the central API server. The full key is only displayed once
at creation time - store it securely as it cannot be retrieved later.

Worker keys are assigned to a specific region and used by the probe
controller in that region to authenticate API requests.

IMPORTANT: The key value shown in the output is sensitive and will only
be displayed this one time. Copy it immediately and store it securely.

Examples:
  # Create a worker key for NYC region
  stackeye admin worker-key create --region nyc3

  # Create a worker key with a custom name
  stackeye admin worker-key create --region lon1 --name "London Probe Primary"

  # Output as JSON for scripting
  stackeye admin worker-key create --region fra1 -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdminWorkerKeyCreate(cmd.Context())
		},
	}

	// Add required flags
	cmd.Flags().StringVarP(&workerKeyCreateRegion, "region", "r", "", "Region identifier for the worker key (required)")
	_ = cmd.MarkFlagRequired("region")

	// Add optional flags
	cmd.Flags().StringVarP(&workerKeyCreateName, "name", "n", "", "Optional display name for the worker key")

	return cmd
}

// runAdminWorkerKeyCreate executes the worker-key create command logic.
func runAdminWorkerKeyCreate(ctx context.Context) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build request
	req := admin.CreateWorkerKeyRequest{
		Region: workerKeyCreateRegion,
		Name:   workerKeyCreateName,
	}

	// Call SDK to create worker key with timeout
	reqCtx, cancel := context.WithTimeout(ctx, adminWorkerKeyCreateTimeout)
	defer cancel()

	response, err := admin.CreateWorkerKey(reqCtx, apiClient, req)
	if err != nil {
		return fmt.Errorf("failed to create worker key: %w", err)
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
	printWorkerKeyCreated(response)
	return nil
}

// printWorkerKeyCreated formats and prints the created worker key in a human-friendly format.
func printWorkerKeyCreated(response *admin.CreateWorkerKeyResponse) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              WORKER KEY CREATED SUCCESSFULLY               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Key details
	fmt.Println("  Key Details:")
	fmt.Println("  ┌─────────────────────────────────────────────────────────┐")
	fmt.Printf("  │  ID:        %-44s │\n", truncateWorkerKeyField(response.Data.ID, 44))
	fmt.Printf("  │  Region:    %-44s │\n", response.Data.Region)
	fmt.Printf("  │  Prefix:    %-44s │\n", response.Data.KeyPrefix)
	fmt.Printf("  │  Type:      %-44s │\n", response.Data.KeyType)
	fmt.Printf("  │  Active:    %-44v │\n", response.Data.IsActive)
	fmt.Printf("  │  Created:   %-44s │\n", formatWorkerKeyTime(response.Data.CreatedAt))
	fmt.Println("  └─────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Important warning about the key
	fmt.Println("  ┌─────────────────────────────────────────────────────────┐")
	fmt.Println("  │  ⚠️  IMPORTANT: Save this key immediately!               │")
	fmt.Println("  │  It will NOT be shown again.                            │")
	fmt.Println("  └─────────────────────────────────────────────────────────┘")
	fmt.Println()

	// The actual key (highlighted)
	fmt.Println("  Worker Key:")
	fmt.Println("  ┌─────────────────────────────────────────────────────────┐")
	fmt.Printf("    %s\n", response.Data.Key)
	fmt.Println("  └─────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Usage hint
	fmt.Println("  Usage:")
	fmt.Println("    Add to your regional probe controller configuration:")
	fmt.Println()
	fmt.Println("    WORKER_KEY=" + response.Data.Key)
	fmt.Println()
	fmt.Println("    Or in the Authorization header:")
	fmt.Println("    Authorization: WorkerKey " + response.Data.Key)
	fmt.Println()
}

// truncateWorkerKeyField truncates a string to fit in the display.
func truncateWorkerKeyField(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatWorkerKeyTime formats a time.Time for display.
func formatWorkerKeyTime(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}
	return t.Format("2006-01-02 15:04:05 MST")
}
