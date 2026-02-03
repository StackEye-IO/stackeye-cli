// Program generate creates markdown documentation for all StackEye CLI commands
// using cobra/doc. Run via 'make markdown' or 'go run ./docs/markdown/generate.go'.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	// Import the internal cmd package to access the root command directly.
	// We use the internal package rather than pkg/cli because the generator
	// is part of this module and needs to disable PersistentPreRunE
	// (config loading) which would fail without a real config file.
	"github.com/StackEye-IO/stackeye-cli/internal/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	outputDir := filepath.Join("docs", "markdown", "pages")
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	root := cmd.RootCmd()

	// Disable PersistentPreRunE so generation doesn't require a config file.
	root.PersistentPreRunE = nil
	root.PersistentPreRun = nil

	if err := doc.GenMarkdownTree(root, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating markdown docs: %v\n", err)
		os.Exit(1)
	}

	// Count generated files for output.
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading output directory: %v\n", err)
		os.Exit(1)
	}

	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}

	fmt.Printf("Generated %d markdown docs in %s\n", count, outputDir)
}
