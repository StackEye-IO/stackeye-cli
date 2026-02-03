package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-cli/internal/cmd"
	"github.com/spf13/cobra/doc"
)

func TestGenManTree(t *testing.T) {
	outputDir := t.TempDir()

	root := cmd.RootCmd()
	root.PersistentPreRunE = nil
	root.PersistentPreRun = nil

	header := &doc.GenManHeader{
		Title:   "STACKEYE",
		Section: "1",
		Source:  "StackEye CLI",
		Manual:  "StackEye Manual",
	}

	if err := doc.GenManTree(root, header, outputDir); err != nil {
		t.Fatalf("GenManTree failed: %v", err)
	}

	// Verify man pages were generated.
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("no man pages generated")
	}

	// Verify the root man page exists.
	rootManPage := filepath.Join(outputDir, "stackeye.1")
	if _, err := os.Stat(rootManPage); os.IsNotExist(err) {
		t.Fatal("root man page stackeye.1 not generated")
	}

	// Verify content of root man page has expected sections.
	content, err := os.ReadFile(rootManPage)
	if err != nil {
		t.Fatalf("failed to read root man page: %v", err)
	}

	text := string(content)
	for _, section := range []string{".TH", ".SH NAME", ".SH SYNOPSIS"} {
		if !strings.Contains(text, section) {
			t.Errorf("root man page missing expected section %q", section)
		}
	}

	// Verify subcommand man pages exist (spot-check a few).
	expectedSubcommands := []string{
		"stackeye-version.1",
		"stackeye-login.1",
		"stackeye-probe.1",
		"stackeye-alert.1",
	}
	for _, name := range expectedSubcommands {
		path := filepath.Join(outputDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected man page %s not generated", name)
		}
	}

	t.Logf("Generated %d man pages", len(entries))
}

func TestManPageContentMatchesHelp(t *testing.T) {
	outputDir := t.TempDir()

	root := cmd.RootCmd()
	root.PersistentPreRunE = nil
	root.PersistentPreRun = nil

	header := &doc.GenManHeader{
		Title:   "STACKEYE",
		Section: "1",
		Source:  "StackEye CLI",
		Manual:  "StackEye Manual",
	}

	if err := doc.GenManTree(root, header, outputDir); err != nil {
		t.Fatalf("GenManTree failed: %v", err)
	}

	// Read root man page and verify it contains the short description from --help.
	content, err := os.ReadFile(filepath.Join(outputDir, "stackeye.1"))
	if err != nil {
		t.Fatalf("failed to read root man page: %v", err)
	}

	text := string(content)

	// The short description should appear in the man page.
	if !strings.Contains(text, "Eye on your stack") {
		t.Error("man page does not contain the short description from --help")
	}

	// The long description content should appear.
	if !strings.Contains(text, "uptime monitoring platform") {
		t.Error("man page does not contain long description content")
	}
}
