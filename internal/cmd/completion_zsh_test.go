package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestZshCompletionCmd(t *testing.T) {
	t.Run("generates completion script", func(t *testing.T) {
		cmd := newZshCompletionCmd()

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
			// Verify script contains expected zsh completion patterns
			expectedPatterns := []string{
				"#compdef stackeye",
				"_stackeye",
				"compdef",
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
		cmd := newZshCompletionCmd()
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

		// Script should still be valid zsh and contain core patterns
		if !strings.Contains(output, "#compdef stackeye") {
			t.Error("completion script missing #compdef directive")
		}
		if !strings.Contains(output, "_stackeye") {
			t.Error("completion script missing _stackeye function")
		}
	})
}

func TestZshCompletionScriptSyntax(t *testing.T) {
	// Skip if zsh is not available
	zshPath, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not available, skipping syntax validation")
	}

	cmd := newZshCompletionCmd()

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

	// Validate zsh syntax using zsh -n
	zshCmd := exec.Command(zshPath, "-n")
	zshCmd.Stdin = strings.NewReader(script)
	var stderr bytes.Buffer
	zshCmd.Stderr = &stderr

	if err := zshCmd.Run(); err != nil {
		t.Errorf("zsh syntax validation failed: %v\nstderr: %s", err, stderr.String())
	}
}

func TestCompletionCmdHasZshSubcommand(t *testing.T) {
	cmd := NewCompletionCmd()

	t.Run("has zsh subcommand", func(t *testing.T) {
		zshCmd, _, err := cmd.Find([]string{"zsh"})
		if err != nil {
			t.Fatalf("failed to find zsh subcommand: %v", err)
		}
		if zshCmd.Use != "zsh" {
			t.Errorf("expected Use='zsh', got %q", zshCmd.Use)
		}
	})

	t.Run("completion long help mentions zsh", func(t *testing.T) {
		if !strings.Contains(cmd.Long, "zsh") {
			t.Error("completion command Long help should mention zsh")
		}
	})
}
