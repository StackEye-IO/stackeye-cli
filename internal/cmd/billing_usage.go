// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-cli/internal/output"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// billingUsageTimeout is the maximum time to wait for the API response.
const billingUsageTimeout = 30 * time.Second

// NewBillingUsageCmd creates and returns the billing usage command.
func NewBillingUsageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage",
		Short: "Show current resource usage against plan limits",
		Long: `Display current resource usage for your organization against plan limits.

Shows your consumption of monitors, team members, and probe checks
for the current billing period. Use this to track your usage and
plan for upgrades when approaching limits.

Output includes:
  - Monitor usage (current count vs. plan limit)
  - Team member usage (current count vs. plan limit)
  - Probe checks performed this billing period
  - Current billing period dates

Examples:
  # Show usage statistics
  stackeye billing usage

  # Output as JSON for scripting
  stackeye billing usage -o json

  # Output as YAML
  stackeye billing usage -o yaml`,
		Aliases: []string{"metrics", "stats"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBillingUsage(cmd.Context())
		},
	}

	return cmd
}

// runBillingUsage executes the billing usage command logic.
func runBillingUsage(ctx context.Context) error {
	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Call SDK to get usage info with timeout
	reqCtx, cancel := context.WithTimeout(ctx, billingUsageTimeout)
	defer cancel()

	usage, err := client.GetUsage(reqCtx, apiClient)
	if err != nil {
		return fmt.Errorf("failed to get usage info: %w", err)
	}

	// Check output format - use JSON/YAML if requested, otherwise pretty print
	cfg := GetConfig()
	if cfg != nil && cfg.Preferences != nil {
		switch cfg.Preferences.OutputFormat {
		case "json", "yaml":
			return output.Print(usage)
		}
	}

	// Pretty print for table format (default)
	printUsageInfo(usage)
	return nil
}

// printUsageInfo formats and prints the usage info in a human-friendly format.
func printUsageInfo(usage *client.UsageInfo) {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    RESOURCE USAGE                          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Billing period section
	fmt.Println("  ┌─────────────────────────────────────────┐")
	fmt.Println("  │            BILLING PERIOD               │")
	fmt.Println("  ├─────────────────────────────────────────┤")
	periodStart := parseAndFormatDate(usage.PeriodStart)
	periodEnd := parseAndFormatDate(usage.PeriodEnd)
	fmt.Printf("  │  Start:    %-28s │\n", periodStart)
	fmt.Printf("  │  End:      %-28s │\n", periodEnd)
	fmt.Println("  └─────────────────────────────────────────┘")
	fmt.Println()

	// Monitors section
	monitorsPercent := calculateUsagePercent(usage.MonitorsCount, usage.MonitorsLimit)
	fmt.Println("  ┌─────────────────────────────────────────┐")
	fmt.Println("  │              MONITORS                   │")
	fmt.Println("  ├─────────────────────────────────────────┤")
	fmt.Printf("  │  Usage:    %d / %d", usage.MonitorsCount, usage.MonitorsLimit)
	fmt.Printf("%s │\n", padRight("", 25-len(fmt.Sprintf("%d / %d", usage.MonitorsCount, usage.MonitorsLimit))))
	fmt.Printf("  │  %s  %5.1f%% │\n", formatUsageBar(monitorsPercent, 24), monitorsPercent)
	if monitorsPercent >= 90 {
		fmt.Println("  │  ⚠ Approaching limit!                  │")
	}
	fmt.Println("  └─────────────────────────────────────────┘")
	fmt.Println()

	// Team Members section
	teamPercent := calculateUsagePercent(usage.TeamMembersCount, usage.TeamMembersLimit)
	fmt.Println("  ┌─────────────────────────────────────────┐")
	fmt.Println("  │            TEAM MEMBERS                 │")
	fmt.Println("  ├─────────────────────────────────────────┤")
	fmt.Printf("  │  Usage:    %d / %d", usage.TeamMembersCount, usage.TeamMembersLimit)
	fmt.Printf("%s │\n", padRight("", 25-len(fmt.Sprintf("%d / %d", usage.TeamMembersCount, usage.TeamMembersLimit))))
	fmt.Printf("  │  %s  %5.1f%% │\n", formatUsageBar(teamPercent, 24), teamPercent)
	if teamPercent >= 90 {
		fmt.Println("  │  ⚠ Approaching limit!                  │")
	}
	fmt.Println("  └─────────────────────────────────────────┘")
	fmt.Println()

	// Checks section
	fmt.Println("  ┌─────────────────────────────────────────┐")
	fmt.Println("  │          PROBE CHECKS (Period)         │")
	fmt.Println("  ├─────────────────────────────────────────┤")
	checksFormatted := formatLargeNumber(usage.ChecksCount)
	fmt.Printf("  │  Total:    %-28s │\n", checksFormatted)
	fmt.Println("  └─────────────────────────────────────────┘")
	fmt.Println()

	// Quick actions hint
	fmt.Println("  Quick Actions:")
	fmt.Println("    stackeye billing info      - View subscription details")
	fmt.Println("    stackeye billing invoices  - View invoice history")
	fmt.Println()
}

// calculateUsagePercent calculates usage percentage safely.
func calculateUsagePercent(used, limit int) float64 {
	if limit <= 0 {
		return 0
	}
	return float64(used) / float64(limit) * 100
}

// formatUsageBar creates a visual progress bar for usage percentage.
// width specifies the total character width of the bar including brackets.
func formatUsageBar(percent float64, width int) string {
	if width < 4 {
		width = 4
	}
	innerWidth := width - 2 // Account for [ and ]

	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}

	filledCount := min(int(percent/100*float64(innerWidth)), innerWidth)

	// Use different characters based on usage level
	fillChar := "█"
	if percent >= 90 {
		fillChar = "▓" // Warning indicator for high usage
	}

	filled := strings.Repeat(fillChar, filledCount)
	empty := strings.Repeat("░", innerWidth-filledCount)

	return "[" + filled + empty + "]"
}

// formatLargeNumber formats a large number with thousand separators.
func formatLargeNumber(n int64) string {
	if n < 0 {
		return "-" + formatLargeNumber(-n)
	}
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	str := fmt.Sprintf("%d", n)
	var result strings.Builder

	// Process from right to left
	length := len(str)
	for i, ch := range str {
		if i > 0 && (length-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(ch)
	}

	return result.String()
}

// padRight pads a string with spaces on the right to reach the specified width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
