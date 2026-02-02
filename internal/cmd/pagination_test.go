// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDefaultPaginationFlags(t *testing.T) {
	flags := DefaultPaginationFlags()

	if flags.Page != DefaultPage {
		t.Errorf("expected Page=%d, got %d", DefaultPage, flags.Page)
	}

	if flags.Limit != DefaultLimit {
		t.Errorf("expected Limit=%d, got %d", DefaultLimit, flags.Limit)
	}
}

func TestAddPaginationFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	flags := DefaultPaginationFlags()

	AddPaginationFlags(cmd, flags)

	// Verify --page flag exists
	pageFlag := cmd.Flags().Lookup("page")
	if pageFlag == nil {
		t.Error("expected --page flag to be registered")
	}
	if pageFlag != nil && pageFlag.DefValue != "1" {
		t.Errorf("expected --page default=1, got %s", pageFlag.DefValue)
	}

	// Verify --limit flag exists
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Error("expected --limit flag to be registered")
	}
	if limitFlag != nil && limitFlag.DefValue != "20" {
		t.Errorf("expected --limit default=20, got %s", limitFlag.DefValue)
	}
}

func TestAddPaginationFlags_CustomDefaults(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	flags := &PaginationFlags{Page: 5, Limit: 50}

	AddPaginationFlags(cmd, flags)

	pageFlag := cmd.Flags().Lookup("page")
	if pageFlag != nil && pageFlag.DefValue != "5" {
		t.Errorf("expected --page default=5, got %s", pageFlag.DefValue)
	}

	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag != nil && limitFlag.DefValue != "50" {
		t.Errorf("expected --limit default=50, got %s", limitFlag.DefValue)
	}
}

func TestValidatePaginationFlags_Valid(t *testing.T) {
	tests := []struct {
		name  string
		page  int
		limit int
	}{
		{"default values", 1, 20},
		{"minimum values", 1, 1},
		{"maximum limit", 1, 100},
		{"high page number", 999, 50},
		{"mid-range values", 5, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &PaginationFlags{Page: tt.page, Limit: tt.limit}
			err := ValidatePaginationFlags(flags)
			if err != nil {
				t.Errorf("expected no error for page=%d, limit=%d, got: %v", tt.page, tt.limit, err)
			}
		})
	}
}

func TestValidatePaginationFlags_InvalidPage(t *testing.T) {
	tests := []struct {
		name        string
		page        int
		expectedMsg string
	}{
		{"zero page", 0, "invalid page 0"},
		{"negative page", -1, "invalid page -1"},
		{"very negative page", -100, "invalid page -100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &PaginationFlags{Page: tt.page, Limit: 20}
			err := ValidatePaginationFlags(flags)
			if err == nil {
				t.Error("expected error for invalid page, got nil")
			}
			if err != nil && !strings.Contains(err.Error(), tt.expectedMsg) {
				t.Errorf("expected error containing %q, got: %v", tt.expectedMsg, err)
			}
		})
	}
}

func TestValidatePaginationFlags_InvalidLimit(t *testing.T) {
	tests := []struct {
		name        string
		limit       int
		expectedMsg string
	}{
		{"zero limit", 0, "invalid limit 0"},
		{"negative limit", -1, "invalid limit -1"},
		{"over max limit", 101, "invalid limit 101"},
		{"way over max", 1000, "invalid limit 1000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &PaginationFlags{Page: 1, Limit: tt.limit}
			err := ValidatePaginationFlags(flags)
			if err == nil {
				t.Error("expected error for invalid limit, got nil")
			}
			if err != nil && !strings.Contains(err.Error(), tt.expectedMsg) {
				t.Errorf("expected error containing %q, got: %v", tt.expectedMsg, err)
			}
		})
	}
}

func TestPageToOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		expected int
	}{
		{"first page", 1, 20, 0},
		{"second page", 2, 20, 20},
		{"third page", 3, 20, 40},
		{"custom limit", 2, 50, 50},
		{"page 10", 10, 10, 90},
		{"zero page returns 0", 0, 20, 0},
		{"negative page returns 0", -1, 20, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PageToOffset(tt.page, tt.limit)
			if result != tt.expected {
				t.Errorf("PageToOffset(%d, %d) = %d, want %d", tt.page, tt.limit, result, tt.expected)
			}
		})
	}
}

func TestFormatPaginationInfo(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		total    int
		expected string
	}{
		{"empty results", 1, 20, 0, ""},
		{"single item", 1, 20, 1, "1 item"},
		{"few items on one page", 1, 20, 15, "15 items"},
		{"exactly one page", 1, 20, 20, "20 items"},
		{"two pages first", 1, 20, 25, "Page 1 of 2 (25 total items)"},
		{"two pages second", 2, 20, 25, "Page 2 of 2 (25 total items)"},
		{"many pages", 3, 10, 100, "Page 3 of 10 (100 total items)"},
		{"exact multiple", 2, 25, 100, "Page 2 of 4 (100 total items)"},
		{"last page partial", 5, 10, 45, "Page 5 of 5 (45 total items)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPaginationInfo(tt.page, tt.limit, tt.total)
			if result != tt.expected {
				t.Errorf("FormatPaginationInfo(%d, %d, %d) = %q, want %q",
					tt.page, tt.limit, tt.total, result, tt.expected)
			}
		})
	}
}

func TestTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		total    int
		expected int
	}{
		{"zero total", 20, 0, 0},
		{"less than limit", 20, 15, 1},
		{"exactly limit", 20, 20, 1},
		{"one over limit", 20, 21, 2},
		{"multiple pages", 20, 100, 5},
		{"partial last page", 20, 95, 5},
		{"zero limit", 0, 100, 0},
		{"negative limit", -1, 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TotalPages(tt.limit, tt.total)
			if result != tt.expected {
				t.Errorf("TotalPages(%d, %d) = %d, want %d", tt.limit, tt.total, result, tt.expected)
			}
		})
	}
}

func TestHasNextPage(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		total    int
		expected bool
	}{
		{"first of many", 1, 20, 100, true},
		{"middle page", 3, 20, 100, true},
		{"last page", 5, 20, 100, false},
		{"only one page", 1, 20, 15, false},
		{"exactly one page", 1, 20, 20, false},
		{"second to last", 4, 20, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasNextPage(tt.page, tt.limit, tt.total)
			if result != tt.expected {
				t.Errorf("HasNextPage(%d, %d, %d) = %v, want %v",
					tt.page, tt.limit, tt.total, result, tt.expected)
			}
		})
	}
}

func TestHasPreviousPage(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		expected bool
	}{
		{"first page", 1, false},
		{"second page", 2, true},
		{"tenth page", 10, true},
		{"zero page", 0, false},
		{"negative page", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasPreviousPage(tt.page)
			if result != tt.expected {
				t.Errorf("HasPreviousPage(%d) = %v, want %v", tt.page, result, tt.expected)
			}
		})
	}
}

func TestPaginationConstants(t *testing.T) {
	// Verify constants are set correctly
	if DefaultPage != 1 {
		t.Errorf("DefaultPage = %d, want 1", DefaultPage)
	}
	if DefaultLimit != 20 {
		t.Errorf("DefaultLimit = %d, want 20", DefaultLimit)
	}
	if MaxLimit != 100 {
		t.Errorf("MaxLimit = %d, want 100", MaxLimit)
	}
	if MinLimit != 1 {
		t.Errorf("MinLimit = %d, want 1", MinLimit)
	}
}
