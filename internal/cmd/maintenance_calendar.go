// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/spf13/cobra"
)

// maintenanceCalendarTimeout is the maximum time to wait for the API response.
const maintenanceCalendarTimeout = 30 * time.Second

// maintenanceCalendarFlags holds the flag values for the maintenance calendar command.
type maintenanceCalendarFlags struct {
	viewMonth      bool
	from           string
	includeExpired bool
}

// NewMaintenanceCalendarCmd creates and returns the maintenance calendar subcommand.
func NewMaintenanceCalendarCmd() *cobra.Command {
	flags := &maintenanceCalendarFlags{}

	cmd := &cobra.Command{
		Use:   "calendar",
		Short: "Show maintenance windows in calendar view",
		Long: `Display scheduled maintenance windows in a visual calendar format.

Shows maintenance windows organized by day, making it easy to see when
planned downtime is scheduled. By default, shows the current week.

View Modes:
  Week (default)  Shows 7 days starting from Monday of the current week
  Month           Shows the full current month with all weeks

The calendar displays each day with its scheduled maintenance windows,
including the window name, duration, and time range.

Examples:
  # Show current week (default)
  stackeye maintenance calendar

  # Show current month
  stackeye maintenance calendar --month

  # Show week starting from a specific date
  stackeye maintenance calendar --from 2024-01-15

  # Include expired/past maintenance windows
  stackeye maintenance calendar --include-expired`,
		Aliases: []string{"cal"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMaintenanceCalendar(cmd.Context(), flags)
		},
	}

	// Define command-specific flags
	cmd.Flags().BoolVar(&flags.viewMonth, "month", false, "show full month view")
	cmd.Flags().StringVar(&flags.from, "from", "", "start date for the view (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&flags.includeExpired, "include-expired", false, "include expired maintenance windows")

	return cmd
}

// runMaintenanceCalendar executes the maintenance calendar command logic.
func runMaintenanceCalendar(ctx context.Context, flags *maintenanceCalendarFlags) error {
	// Parse the start date if provided
	var startDate time.Time
	var err error

	if flags.from != "" {
		startDate, err = time.ParseInLocation("2006-01-02", flags.from, time.Local)
		if err != nil {
			return fmt.Errorf("invalid date format %q: use YYYY-MM-DD", flags.from)
		}
	} else {
		startDate = time.Now()
	}

	// Calculate the date range based on view mode
	rangeStart, rangeEnd := calculateDateRange(startDate, flags.viewMonth)

	// Get authenticated API client
	apiClient, err := api.GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize API client: %w", err)
	}

	// Build list options - fetch all maintenance windows in the range
	opts := &client.ListMutesOptions{
		Limit:           100, // Fetch up to 100 windows
		Offset:          0,
		IncludeExpired:  flags.includeExpired,
		MaintenanceOnly: true,
	}

	// Call SDK to list maintenance windows with timeout
	reqCtx, cancel := context.WithTimeout(ctx, maintenanceCalendarTimeout)
	defer cancel()

	result, err := client.ListMutes(reqCtx, apiClient, opts)
	if err != nil {
		return fmt.Errorf("failed to list maintenance windows: %w", err)
	}

	// Warn if results may be truncated
	if result.Total > int64(len(result.Data)) {
		fmt.Printf("Warning: Showing %d of %d maintenance windows. Results may be incomplete.\n\n",
			len(result.Data), result.Total)
	}

	// Filter windows to the date range
	windows := filterWindowsByDateRange(result.Data, rangeStart, rangeEnd)

	// Build and print the calendar
	printCalendar(rangeStart, rangeEnd, windows, flags.viewMonth)

	return nil
}

// calculateDateRange determines the start and end dates based on view mode.
func calculateDateRange(baseDate time.Time, isMonthView bool) (time.Time, time.Time) {
	if isMonthView {
		// Start of the month
		firstOfMonth := time.Date(baseDate.Year(), baseDate.Month(), 1, 0, 0, 0, 0, baseDate.Location())
		// End of the month
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
		return firstOfMonth, lastOfMonth
	}

	// Week view: find Monday of the current week
	weekday := baseDate.Weekday()
	daysToMonday := int(weekday) - int(time.Monday)
	if daysToMonday < 0 {
		daysToMonday += 7
	}
	monday := baseDate.AddDate(0, 0, -daysToMonday)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
	sunday := monday.AddDate(0, 0, 6)

	return monday, sunday
}

// filterWindowsByDateRange filters maintenance windows to only those that overlap with the date range.
func filterWindowsByDateRange(windows []client.AlertMute, rangeStart, rangeEnd time.Time) []client.AlertMute {
	var filtered []client.AlertMute
	rangeEndOfDay := time.Date(rangeEnd.Year(), rangeEnd.Month(), rangeEnd.Day(), 23, 59, 59, 0, rangeEnd.Location())

	for _, w := range windows {
		windowEnd := w.StartsAt.Add(time.Duration(w.DurationMinutes) * time.Minute)

		// Check if window overlaps with the date range
		// Window overlaps if: windowStart <= rangeEnd AND windowEnd >= rangeStart
		if !w.StartsAt.After(rangeEndOfDay) && !windowEnd.Before(rangeStart) {
			filtered = append(filtered, w)
		}
	}

	return filtered
}

// printCalendar renders the calendar view to stdout.
func printCalendar(rangeStart, rangeEnd time.Time, windows []client.AlertMute, isMonthView bool) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Print header
	if isMonthView {
		fmt.Printf("Maintenance Calendar - %s %d\n", rangeStart.Month().String(), rangeStart.Year())
	} else {
		fmt.Printf("Maintenance Calendar - Week of %s\n", rangeStart.Format("Jan 2, 2006"))
	}
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println()

	// Group windows by day
	windowsByDay := groupWindowsByDay(windows)

	// Iterate through each day in the range
	current := rangeStart
	for !current.After(rangeEnd) {
		dayKey := current.Format("2006-01-02")
		dayWindows := windowsByDay[dayKey]

		isToday := current.Equal(today)
		printDay(current, dayWindows, isToday)

		current = current.AddDate(0, 0, 1)
	}

	// Print summary
	fmt.Println()
	fmt.Println(strings.Repeat("─", 70))
	totalWindows := len(windows)
	if totalWindows == 0 {
		fmt.Println("No maintenance windows scheduled in this period.")
	} else {
		fmt.Printf("Total: %d maintenance window(s) scheduled\n", totalWindows)
	}
}

// groupWindowsByDay organizes maintenance windows by their start date.
func groupWindowsByDay(windows []client.AlertMute) map[string][]client.AlertMute {
	byDay := make(map[string][]client.AlertMute)

	for _, w := range windows {
		// A window might span multiple days - add it to each day it covers
		windowEnd := w.StartsAt.Add(time.Duration(w.DurationMinutes) * time.Minute)
		current := time.Date(w.StartsAt.Year(), w.StartsAt.Month(), w.StartsAt.Day(), 0, 0, 0, 0, w.StartsAt.Location())

		for !current.After(windowEnd) {
			dayKey := current.Format("2006-01-02")
			byDay[dayKey] = append(byDay[dayKey], w)
			current = current.AddDate(0, 0, 1)
		}
	}

	return byDay
}

// printDay renders a single day with its maintenance windows.
func printDay(date time.Time, windows []client.AlertMute, isToday bool) {
	// Format the day header
	dayName := date.Weekday().String()[:3]
	dateStr := date.Format("Jan 2")

	todayMarker := ""
	if isToday {
		todayMarker = " (TODAY)"
	}

	// Print day header
	fmt.Printf("%s %s%s\n", dayName, dateStr, todayMarker)

	if len(windows) == 0 {
		fmt.Println("  No scheduled maintenance")
	} else {
		// Sort windows by start time
		sort.Slice(windows, func(i, j int) bool {
			return windows[i].StartsAt.Before(windows[j].StartsAt)
		})

		for _, w := range windows {
			printWindow(w, date)
		}
	}
	fmt.Println()
}

// printWindow renders a single maintenance window entry.
func printWindow(w client.AlertMute, dayDate time.Time) {
	windowEnd := w.StartsAt.Add(time.Duration(w.DurationMinutes) * time.Minute)

	// Format the time range
	startTime := w.StartsAt.Format("15:04")
	endTime := windowEnd.Format("15:04")

	// Check if window spans this entire day
	dayStart := time.Date(dayDate.Year(), dayDate.Month(), dayDate.Day(), 0, 0, 0, 0, dayDate.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)

	timeRange := fmt.Sprintf("%s - %s", startTime, endTime)
	if w.StartsAt.Before(dayStart) && windowEnd.After(dayEnd) {
		timeRange = "All day"
	} else if w.StartsAt.Before(dayStart) {
		timeRange = fmt.Sprintf("... - %s", endTime)
	} else if windowEnd.After(dayEnd) {
		timeRange = fmt.Sprintf("%s - ...", startTime)
	}

	// Get the window name or generate a default
	name := "Maintenance"
	if w.MaintenanceName != nil && *w.MaintenanceName != "" {
		name = *w.MaintenanceName
	}

	// Truncate name if too long
	if len(name) > 35 {
		name = name[:32] + "..."
	}

	// Format scope
	scope := formatCalendarScope(w)

	// Format duration
	duration := formatCalendarDuration(w.DurationMinutes)

	// Print the window
	fmt.Printf("  %s %-35s [%s] %s\n", timeRange, name, scope, duration)
}

// formatCalendarScope returns a short scope description.
func formatCalendarScope(w client.AlertMute) string {
	switch w.ScopeType {
	case client.MuteScopeOrganization:
		return "Org"
	case client.MuteScopeProbe:
		return "Probe"
	case client.MuteScopeChannel:
		return "Channel"
	case client.MuteScopeAlertType:
		return "Type"
	default:
		return string(w.ScopeType)
	}
}

// formatCalendarDuration formats duration for calendar display.
func formatCalendarDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	remainingMins := minutes % 60
	if remainingMins == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, remainingMins)
}
