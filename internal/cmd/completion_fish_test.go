package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestFishCompletionCmd(t *testing.T) {
	t.Run("generates completion script", func(t *testing.T) {
		cmd := newFishCompletionCmd()

		// Use pipe with goroutine to avoid deadlock
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("failed to create pipe: %v", err)
		}

		oldStdout := os.Stdout
		os.Stdout = w

		// Read output in background to prevent pipe buffer deadlock
		outputCh := make(chan string)
		errCh := make(chan error)
		go func() {
			var buf bytes.Buffer
			_, readErr := io.Copy(&buf, r)
			if readErr != nil {
				errCh <- readErr
				return
			}
			outputCh <- buf.String()
		}()

		runErr := cmd.RunE(cmd, []string{})

		// Close write end and restore stdout
		w.Close()
		os.Stdout = oldStdout

		if runErr != nil {
			t.Fatalf("RunE returned error: %v", runErr)
		}

		// Wait for reader goroutine
		select {
		case readErr := <-errCh:
			t.Fatalf("failed to read output: %v", readErr)
		case output := <-outputCh:
			// Verify script contains expected fish completion patterns
			expectedPatterns := []string{
				"complete -c stackeye",
				"function",
			}

			for _, pattern := range expectedPatterns {
				if !strings.Contains(output, pattern) {
					t.Errorf("completion script missing expected pattern: %q", pattern)
				}
			}

			// Verify script is non-trivial (should be at least a few KB)
			if len(output) < 500 {
				t.Errorf("completion script seems too small: %d bytes", len(output))
			}
		}
	})

	t.Run("no-descriptions flag works", func(t *testing.T) {
		cmd := newFishCompletionCmd()
		_ = cmd.Flags().Set("no-descriptions", "true")

		// Use pipe with goroutine to avoid deadlock
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("failed to create pipe: %v", err)
		}

		oldStdout := os.Stdout
		os.Stdout = w

		// Read output in background
		outputCh := make(chan string)
		go func() {
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			outputCh <- buf.String()
		}()

		runErr := cmd.RunE(cmd, []string{})

		w.Close()
		os.Stdout = oldStdout

		if runErr != nil {
			t.Fatalf("RunE with --no-descriptions returned error: %v", runErr)
		}

		output := <-outputCh

		// Script should still contain core fish completion patterns
		if !strings.Contains(output, "complete -c stackeye") {
			t.Error("completion script missing complete -c directive")
		}
	})
}

func TestFishCompletionScriptSyntax(t *testing.T) {
	// Skip if fish is not available
	fishPath, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish not available, skipping syntax validation")
	}

	cmd := newFishCompletionCmd()

	// Use pipe with goroutine to avoid deadlock
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("failed to create pipe: %v", pipeErr)
	}

	oldStdout := os.Stdout
	os.Stdout = w

	// Read output in background
	scriptCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		scriptCh <- buf.String()
	}()

	err = cmd.RunE(cmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}

	script := <-scriptCh

	// Validate fish syntax using fish -n (no-execute mode)
	fishCmd := exec.Command(fishPath, "-n")
	fishCmd.Stdin = strings.NewReader(script)
	var stderr bytes.Buffer
	fishCmd.Stderr = &stderr

	if err := fishCmd.Run(); err != nil {
		t.Errorf("fish syntax validation failed: %v\nstderr: %s", err, stderr.String())
	}
}

func TestCompletionCmdHasFishSubcommand(t *testing.T) {
	cmd := NewCompletionCmd()

	t.Run("has fish subcommand", func(t *testing.T) {
		fishCmd, _, err := cmd.Find([]string{"fish"})
		if err != nil {
			t.Fatalf("failed to find fish subcommand: %v", err)
		}
		if fishCmd.Use != "fish" {
			t.Errorf("expected Use='fish', got %q", fishCmd.Use)
		}
	})

	t.Run("completion long help mentions fish", func(t *testing.T) {
		if !strings.Contains(cmd.Long, "fish") {
			t.Error("completion command Long help should mention fish")
		}
	})
}
