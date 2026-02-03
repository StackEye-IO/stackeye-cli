package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-cli/internal/cmd"
	"github.com/spf13/cobra/doc"
)

func TestGenMarkdownTree(t *testing.T) {
	outputDir := t.TempDir()

	root := cmd.RootCmd()
	root.PersistentPreRunE = nil
	root.PersistentPreRun = nil

	if err := doc.GenMarkdownTree(root, outputDir); err != nil {
		t.Fatalf("GenMarkdownTree failed: %v", err)
	}

	// Verify markdown docs were generated.
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("no markdown docs generated")
	}

	// Verify the root markdown doc exists.
	rootDoc := filepath.Join(outputDir, "stackeye.md")
	if _, err := os.Stat(rootDoc); os.IsNotExist(err) {
		t.Fatal("root markdown doc stackeye.md not generated")
	}

	// Verify content of root markdown doc has expected sections.
	content, err := os.ReadFile(rootDoc)
	if err != nil {
		t.Fatalf("failed to read root markdown doc: %v", err)
	}

	text := string(content)
	for _, section := range []string{"## stackeye", "### Synopsis", "### Options"} {
		if !strings.Contains(text, section) {
			t.Errorf("root markdown doc missing expected section %q", section)
		}
	}

	// Verify subcommand markdown docs exist (spot-check a few).
	expectedSubcommands := []string{
		"stackeye_version.md",
		"stackeye_login.md",
		"stackeye_probe.md",
		"stackeye_alert.md",
	}
	for _, name := range expectedSubcommands {
		path := filepath.Join(outputDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected markdown doc %s not generated", name)
		}
	}

	t.Logf("Generated %d markdown docs", len(entries))
}

func TestMarkdownContentMatchesHelp(t *testing.T) {
	outputDir := t.TempDir()

	root := cmd.RootCmd()
	root.PersistentPreRunE = nil
	root.PersistentPreRun = nil

	if err := doc.GenMarkdownTree(root, outputDir); err != nil {
		t.Fatalf("GenMarkdownTree failed: %v", err)
	}

	// Read root markdown doc and verify it contains the short description from --help.
	content, err := os.ReadFile(filepath.Join(outputDir, "stackeye.md"))
	if err != nil {
		t.Fatalf("failed to read root markdown doc: %v", err)
	}

	text := string(content)

	// The short description should appear in the markdown doc.
	if !strings.Contains(text, "Eye on your stack") {
		t.Error("markdown doc does not contain the short description from --help")
	}

	// The long description content should appear.
	if !strings.Contains(text, "uptime monitoring platform") {
		t.Error("markdown doc does not contain long description content")
	}
}
