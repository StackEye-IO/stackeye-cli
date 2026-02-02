// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// PaginationDefaults contains default values for pagination flags.
const (
	// DefaultPage is the default page number (1-indexed).
	DefaultPage = 1
	// DefaultLimit is the default number of results per page.
	DefaultLimit = 20
	// MaxLimit is the maximum allowed results per page.
	MaxLimit = 100
	// MinLimit is the minimum allowed results per page.
	MinLimit = 1
)

// PaginationFlags holds common pagination flag values.
// Use this struct to add consistent --page and --limit flags to list commands.
type PaginationFlags struct {
	// Page is the 1-indexed page number to retrieve.
	Page int
	// Limit is the number of results per page (1-100).
	Limit int
}

// DefaultPaginationFlags returns a PaginationFlags with sensible defaults.
// Page defaults to 1, Limit defaults to 20.
func DefaultPaginationFlags() *PaginationFlags {
	return &PaginationFlags{
		Page:  DefaultPage,
		Limit: DefaultLimit,
	}
}

// AddPaginationFlags registers --page and --limit flags on a cobra command.
// The flags are bound to the provided PaginationFlags struct.
//
// Example:
//
//	flags := DefaultPaginationFlags()
//	AddPaginationFlags(cmd, flags)
func AddPaginationFlags(cmd *cobra.Command, flags *PaginationFlags) {
	cmd.Flags().IntVar(&flags.Page, "page", flags.Page, "page number for pagination")
	cmd.Flags().IntVar(&flags.Limit, "limit", flags.Limit, "results per page (max: 100)")
}

// ValidatePaginationFlags validates the page and limit values.
// Returns an error if values are outside acceptable ranges.
//
// Validation rules:
//   - Page must be >= 1
//   - Limit must be between 1 and 100 (inclusive)
func ValidatePaginationFlags(flags *PaginationFlags) error {
	if flags.Page < 1 {
		return fmt.Errorf("invalid page %d: must be at least 1", flags.Page)
	}

	if flags.Limit < MinLimit || flags.Limit > MaxLimit {
		return fmt.Errorf("invalid limit %d: must be between %d and %d", flags.Limit, MinLimit, MaxLimit)
	}

	return nil
}

// PageToOffset converts a 1-indexed page number to a 0-indexed offset.
// This is useful when the API uses offset-based pagination.
//
// Example:
//
//	offset := PageToOffset(2, 20) // Returns 20 (skip first page of 20 items)
func PageToOffset(page, limit int) int {
	if page < 1 {
		return 0
	}
	return (page - 1) * limit
}

// FormatPaginationInfo formats pagination metadata for display.
// Returns a human-readable string showing current position in results.
//
// Examples:
//   - "Page 1 of 5 (100 total items)"
//   - "Page 3 of 3 (45 total items)"
//   - "1 item" (when total is 1)
//   - "45 items" (when showing all on one page)
func FormatPaginationInfo(page, limit, total int) string {
	if total == 0 {
		return ""
	}

	// Calculate total pages
	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	// Handle single item case
	itemWord := "items"
	if total == 1 {
		itemWord = "item"
	}

	// If everything fits on one page, just show count
	if totalPages <= 1 {
		return fmt.Sprintf("%d %s", total, itemWord)
	}

	// Show full pagination info
	return fmt.Sprintf("Page %d of %d (%d total %s)", page, totalPages, total, itemWord)
}

// TotalPages calculates the total number of pages given limit and total count.
func TotalPages(limit, total int) int {
	if limit <= 0 {
		return 0
	}
	pages := total / limit
	if total%limit > 0 {
		pages++
	}
	return pages
}

// HasNextPage returns true if there are more pages after the current page.
func HasNextPage(page, limit, total int) bool {
	return page < TotalPages(limit, total)
}

// HasPreviousPage returns true if there are pages before the current page.
func HasPreviousPage(page int) bool {
	return page > 1
}
