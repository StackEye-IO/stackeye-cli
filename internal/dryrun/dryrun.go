package dryrun

import (
	"fmt"
	"os"
)

// PrintAction outputs a dry-run message for a single operation with key-value details.
// Example: PrintAction("create", "probe", "Name", "My Probe", "URL", "https://example.com")
func PrintAction(action, resource string, details ...string) {
	fmt.Fprintf(os.Stderr, "Dry run: Would %s %s with:\n", action, resource)

	// Print key-value pairs
	for i := 0; i < len(details); i += 2 {
		if i+1 < len(details) {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", details[i], details[i+1])
		}
	}

	fmt.Fprintf(os.Stderr, "\nNo %s was %sd (dry run).\n", resource, action)
}

// PrintBatchAction outputs a dry-run message for batch operations.
// Example: PrintBatchAction("delete", "probe", []string{"abc-123", "def-456"})
func PrintBatchAction(action, resource string, items []string) {
	count := len(items)
	fmt.Fprintf(os.Stderr, "Dry run: Would %s %d %s(s):\n", action, count, resource)

	for _, item := range items {
		fmt.Fprintf(os.Stderr, "  - %s\n", item)
	}

	fmt.Fprintf(os.Stderr, "\nNo %ss were %sd (dry run).\n", resource, action)
}
