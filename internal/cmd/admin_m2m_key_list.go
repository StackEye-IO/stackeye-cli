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

// adminM2MKeyListTimeout is the maximum time to wait for the API response.
const adminM2MKeyListTimeout = 30 * time.Second

// NewAdminM2MKeyListCmd creates and returns the m2m-key list command.
func NewAdminM2MKeyListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all M2M keys",
		Long: `List all machine-to-machine (M2M) authentication keys.

This command displays all M2M keys with their status, region assignment,
and last activity timestamp. M2M keys are used for service-to-service
authentication.

The output includes:
  - Key prefix (for identification without exposing full key)
  - Region assignment (or "global" if not region-specific)
  - Active/inactive status
  - Last seen timestamp
  - Creation date

Examples:
  # List all M2M keys in table format (default)
  stackeye admin m2m-key list

  # List in JSON format for scripting
  stackeye admin m2m-key list -o json

  # List in YAML format
  stackeye admin m2m-key list -o yaml`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdminM2MKeyList(cmd.Context())
		},
	}

	return cmd
}

// runAdminM2MKeyList executes the m2m-key list command logic.
func runAdminM2MKeyList(ctx context.Context) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to list M2M keys with timeout
	reqCtx, cancel := context.WithTimeout(ctx, adminM2MKeyListTimeout)
	defer cancel()

	response, err := admin.ListM2MKeys(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to list M2M keys: %w", err)
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
	printM2MKeyList(response)
	return nil
}

// printM2MKeyList formats and prints the M2M key list in a human-friendly format.
func printM2MKeyList(response *admin.M2MKeyListResponse) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                      M2M KEYS                              ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(response.Data) == 0 {
		fmt.Println("  No M2M keys found.")
		fmt.Println()
		return
	}

	fmt.Printf("  Total: %d key(s)\n\n", response.Total)

	// Print header
	fmt.Println("  ┌──────────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("  │ %-20s │ %-10s │ %-8s │ %-14s │ %-12s │\n",
		"KEY PREFIX", "REGION", "ACTIVE", "LAST SEEN", "CREATED")
	fmt.Println("  ├──────────────────────────────────────────────────────────────────────────────┤")

	// Print each key
	for _, key := range response.Data {
		region := key.Region
		if region == "" {
			region = "global"
		}

		activeStr := "No"
		if key.IsActive {
			activeStr = "Yes"
		}

		lastSeen := formatM2MKeyLastSeen(key.LastSeenAt)
		created := formatM2MKeyCreated(key.CreatedAt)

		fmt.Printf("  │ %-20s │ %-10s │ %-8s │ %-14s │ %-12s │\n",
			truncateM2MKeyField(key.KeyPrefix, 20),
			truncateM2MKeyField(region, 10),
			activeStr,
			lastSeen,
			created)
	}

	fmt.Println("  └──────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
}

// formatM2MKeyLastSeen formats the last seen time of an M2M key.
func formatM2MKeyLastSeen(t *time.Time) string {
	if t == nil {
		return "Never"
	}
	if t.IsZero() {
		return "Never"
	}

	// Calculate relative time
	since := time.Since(*t)
	switch {
	case since < time.Minute:
		return "Just now"
	case since < time.Hour:
		mins := int(since.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	case since < 24*time.Hour:
		hours := int(since.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(since.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// formatM2MKeyCreated formats the creation timestamp for display.
func formatM2MKeyCreated(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}
	return t.Format("Jan 02, 2006")
}

// truncateM2MKeyField truncates a string to fit in the display.
func truncateM2MKeyField(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
