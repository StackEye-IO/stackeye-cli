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

// dashboardTimeout is the maximum time to wait for the API response.
const dashboardTimeout = 30 * time.Second

// dashboardFlags holds the flag values for the dashboard command.
type dashboardFlags struct {
	period string
}

// NewDashboardCmd creates and returns the dashboard command.
func NewDashboardCmd() *cobra.Command {
	flags := &dashboardFlags{}

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Display organization monitoring overview",
		Long: `Display a summary of your organization's monitoring status.

Shows probe status counts, active alerts, overall uptime, and top probes.
This provides a quick snapshot of your infrastructure health from the
command line.

Period Options:
  24h     Last 24 hours (default)
  7d      Last 7 days
  30d     Last 30 days

Examples:
  # Show dashboard summary
  stackeye dashboard

  # Show dashboard for last 7 days
  stackeye dashboard --period 7d

  # Output as JSON for scripting
  stackeye dashboard -o json`,
		Aliases: []string{"dash", "status", "overview"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDashboard(cmd.Context(), flags)
		},
	}

	cmd.Flags().StringVar(&flags.period, "period", "24h", "time period: 24h, 7d, 30d")

	return cmd
}

// runDashboard executes the dashboard command logic.
func runDashboard(ctx context.Context, flags *dashboardFlags) error {
	// Validate period flag
	switch flags.period {
	case "24h", "7d", "30d":
		// Valid periods
	default:
		return fmt.Errorf("invalid period %q: must be 24h, 7d, or 30d", flags.period)
	}

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get dashboard stats with timeout
	reqCtx, cancel := context.WithTimeout(ctx, dashboardTimeout)
	defer cancel()

	result, err := client.GetDashboardStats(reqCtx, apiClient, flags.period)
	if err != nil {
		return fmt.Errorf("failed to get dashboard stats: %w", err)
	}

	// Check output format - use JSON/YAML if requested, otherwise pretty print
	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(result)
		}
	}

	// Pretty print for table format (default)
	printDashboard(result, flags.period)
	return nil
}

// printDashboard formats and prints the dashboard in a human-friendly format.
func printDashboard(resp *client.DashboardResponse, period string) {
	stats := resp.Stats

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                   STACKEYE DASHBOARD                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Overall health
	fmt.Printf("  Overall Uptime (%s):  %.2f%%\n", period, stats.OverallUptime)
	fmt.Printf("  Avg Response Time:     %.0fms\n", stats.AvgResponseTimeMs)
	fmt.Println()

	// Probe status summary
	fmt.Println("  ┌─────────────────────────────────┐")
	fmt.Println("  │         PROBE STATUS            │")
	fmt.Println("  ├─────────────────────────────────┤")
	fmt.Printf("  │  Total Probes:     %4d         │\n", stats.TotalProbes)
	fmt.Printf("  │  ● Up:             %4d         │\n", stats.ProbesUp)
	fmt.Printf("  │  ○ Down:           %4d         │\n", stats.ProbesDown)
	fmt.Printf("  │  ◐ Degraded:       %4d         │\n", stats.ProbesDegraded)
	fmt.Printf("  │  ⏸ Paused:         %4d         │\n", stats.ProbesPaused)
	fmt.Printf("  │  ◌ Pending:        %4d         │\n", stats.ProbesPending)
	fmt.Println("  └─────────────────────────────────┘")
	fmt.Println()

	// Active alerts
	fmt.Println("  ┌─────────────────────────────────┐")
	fmt.Println("  │         ACTIVE ALERTS           │")
	fmt.Println("  ├─────────────────────────────────┤")
	if stats.ActiveAlertsCount == 0 {
		fmt.Println("  │  ✓ No active alerts             │")
	} else {
		fmt.Printf("  │  ⚠ %d alert(s) need attention   │\n", stats.ActiveAlertsCount)
	}
	fmt.Println("  └─────────────────────────────────┘")
	fmt.Println()

	// Top probes (if any)
	if len(stats.TopProbes) > 0 {
		fmt.Println("  ┌────────────────────────────────────────────────────────────┐")
		fmt.Println("  │                       TOP PROBES                           │")
		fmt.Println("  ├────────────────────────────────────────────────────────────┤")
		fmt.Println("  │  NAME                       STATUS   UPTIME   RESPONSE     │")
		fmt.Println("  ├────────────────────────────────────────────────────────────┤")
		for _, probe := range stats.TopProbes {
			name := truncate(probe.Name, 25)
			statusIcon := getStatusIcon(probe.Status)
			fmt.Printf("  │  %-25s  %s %-6s  %5.1f%%   %5dms     │\n",
				name, statusIcon, probe.Status, probe.Uptime, probe.ResponseTime)
		}
		fmt.Println("  └────────────────────────────────────────────────────────────┘")
		fmt.Println()
	}

	// Quick actions hint
	fmt.Println("  Quick Actions:")
	fmt.Println("    stackeye probe list     - View all probes")
	fmt.Println("    stackeye alert list     - View all alerts")
	fmt.Println("    stackeye probe test ID  - Test a specific probe")
	fmt.Println()
}

// getStatusIcon returns a Unicode icon for the given status.
func getStatusIcon(status string) string {
	switch status {
	case "up":
		return "●"
	case "down":
		return "○"
	case "degraded":
		return "◐"
	case "paused":
		return "⏸"
	case "pending":
		return "◌"
	default:
		return "?"
	}
}

// truncate truncates a string to the specified length, adding "..." if needed.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
