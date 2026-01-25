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

// adminWorkerKeyHealthTimeout is the maximum time to wait for the API response.
const adminWorkerKeyHealthTimeout = 30 * time.Second

// workerKeyHealthThreshold is the threshold flag value (in minutes).
var workerKeyHealthThreshold int

// NewAdminWorkerKeyHealthCmd creates and returns the worker-key health command.
func NewAdminWorkerKeyHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check worker health status",
		Long: `Check the health status of all active workers.

This command displays the health status of regional probe workers, categorizing
them as either "healthy" (checked in recently) or "stale" (not checked in
within the threshold period).

Workers are expected to check in regularly via their heartbeat mechanism.
The default threshold is 5 minutes - workers that haven't checked in within
this period are considered stale and may need investigation.

Examples:
  # Check worker health with default threshold (5 minutes)
  stackeye admin worker-key health

  # Check with a custom threshold (10 minutes)
  stackeye admin worker-key health --threshold 10

  # Output as JSON for scripting
  stackeye admin worker-key health -o json`,
		Aliases: []string{"status"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdminWorkerKeyHealth(cmd.Context())
		},
	}

	// Add optional flags
	cmd.Flags().IntVarP(&workerKeyHealthThreshold, "threshold", "t", 0, "Stale threshold in minutes (default: server default of 5)")

	return cmd
}

// runAdminWorkerKeyHealth executes the worker-key health command logic.
func runAdminWorkerKeyHealth(ctx context.Context) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build options
	var opts *admin.GetWorkerHealthOptions
	if workerKeyHealthThreshold > 0 {
		opts = &admin.GetWorkerHealthOptions{
			ThresholdMinutes: workerKeyHealthThreshold,
		}
	}

	// Call SDK to get worker health with timeout
	reqCtx, cancel := context.WithTimeout(ctx, adminWorkerKeyHealthTimeout)
	defer cancel()

	response, err := admin.GetWorkerHealth(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to get worker health: %w", err)
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
	printWorkerHealth(response)
	return nil
}

// printWorkerHealth formats and prints the worker health status in a human-friendly format.
func printWorkerHealth(status *admin.WorkerHealthResponse) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    WORKER HEALTH STATUS                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Summary statistics
	fmt.Println("  Summary:")
	fmt.Println("  ┌─────────────────────────────────────────────────────────┐")
	fmt.Printf("  │  Total Active:    %-40d │\n", status.TotalActive)
	fmt.Printf("  │  Healthy:         %-40d │\n", status.TotalHealthy)
	fmt.Printf("  │  Stale:           %-40d │\n", status.TotalStale)
	fmt.Printf("  │  Checked At:      %-40s │\n", formatHealthCheckTime(status.CheckedAt))
	fmt.Println("  └─────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Healthy workers
	if len(status.HealthyWorkers) > 0 {
		fmt.Println("  Healthy Workers:")
		fmt.Println("  ┌─────────────────────────────────────────────────────────┐")
		for _, w := range status.HealthyWorkers {
			lastSeen := formatWorkerLastSeen(w.LastSeenAt)
			fmt.Printf("  │  %-12s  %-20s  Last seen: %-12s │\n",
				truncateHealthField(w.Region, 12),
				truncateHealthField(w.KeyPrefix, 20),
				lastSeen)
		}
		fmt.Println("  └─────────────────────────────────────────────────────────┘")
		fmt.Println()
	}

	// Stale workers (highlighted)
	if len(status.StaleWorkers) > 0 {
		fmt.Println("  Stale Workers (need attention):")
		fmt.Println("  ┌─────────────────────────────────────────────────────────┐")
		for _, w := range status.StaleWorkers {
			lastSeen := formatWorkerLastSeen(w.LastSeenAt)
			fmt.Printf("  │  %-12s  %-20s  Last seen: %-12s │\n",
				truncateHealthField(w.Region, 12),
				truncateHealthField(w.KeyPrefix, 20),
				lastSeen)
		}
		fmt.Println("  └─────────────────────────────────────────────────────────┘")
		fmt.Println()
	}

	// Status message
	if status.TotalStale == 0 && status.TotalActive > 0 {
		fmt.Println("  All workers are healthy.")
	} else if status.TotalStale > 0 {
		fmt.Printf("  Warning: %d worker(s) have not checked in recently.\n", status.TotalStale)
	} else if status.TotalActive == 0 {
		fmt.Println("  No active workers found.")
	}
	fmt.Println()
}

// formatHealthCheckTime formats the health check timestamp.
func formatHealthCheckTime(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}
	return t.Format("15:04:05 MST")
}

// formatWorkerLastSeen formats the last seen time of a worker.
func formatWorkerLastSeen(t *time.Time) string {
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

// truncateHealthField truncates a string to fit in the display.
func truncateHealthField(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
