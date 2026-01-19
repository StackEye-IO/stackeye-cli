package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestBashCompletionCmd(t *testing.T) {
	t.Run("generates completion script", func(t *testing.T) {
		cmd := newBashCompletionCmd()

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
			// Verify script contains expected bash completion patterns
			expectedPatterns := []string{
				"# bash completion V2 for stackeye",
				"__start_stackeye",
				"complete -o",
				"stackeye",
			}

			for _, pattern := range expectedPatterns {
				if !strings.Contains(output, pattern) {
					t.Errorf("completion script missing expected pattern: %q", pattern)
				}
			}

			// Verify script is non-trivial (should be at least a few KB)
			if len(output) < 1000 {
				t.Errorf("completion script seems too small: %d bytes", len(output))
			}
		}
	})

	t.Run("no-descriptions flag works", func(t *testing.T) {
		cmd := newBashCompletionCmd()
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

		// Script should still be valid bash
		if !strings.Contains(output, "__start_stackeye") {
			t.Error("completion script missing __start_stackeye function")
		}
	})
}

func TestBashCompletionScriptSyntax(t *testing.T) {
	// Skip if bash is not available
	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not available, skipping syntax validation")
	}

	cmd := newBashCompletionCmd()

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

	// Validate bash syntax using bash -n
	bashCmd := exec.Command(bashPath, "-n")
	bashCmd.Stdin = strings.NewReader(script)
	var stderr bytes.Buffer
	bashCmd.Stderr = &stderr

	if err := bashCmd.Run(); err != nil {
		t.Errorf("bash syntax validation failed: %v\nstderr: %s", err, stderr.String())
	}
}

func TestCompletionCmdStructure(t *testing.T) {
	cmd := NewCompletionCmd()

	t.Run("has correct use", func(t *testing.T) {
		if cmd.Use != "completion" {
			t.Errorf("expected Use='completion', got %q", cmd.Use)
		}
	})

	t.Run("has short description", func(t *testing.T) {
		if cmd.Short == "" {
			t.Error("completion command missing Short description")
		}
	})

	t.Run("has long description with installation instructions", func(t *testing.T) {
		if cmd.Long == "" {
			t.Error("completion command missing Long description")
		}
		// Verify Long description includes installation instructions for all shells
		expectedPatterns := []string{
			"bash",
			"zsh",
			"fish",
			"powershell",
			"bash_completion",
			"compinit",
		}
		for _, pattern := range expectedPatterns {
			if !strings.Contains(strings.ToLower(cmd.Long), strings.ToLower(pattern)) {
				t.Errorf("Long description missing expected pattern: %q", pattern)
			}
		}
	})

	t.Run("has all shell subcommands", func(t *testing.T) {
		shells := []string{"bash", "zsh", "fish", "powershell"}
		for _, shell := range shells {
			subCmd, _, err := cmd.Find([]string{shell})
			if err != nil {
				t.Errorf("failed to find %s subcommand: %v", shell, err)
				continue
			}
			if subCmd.Use != shell {
				t.Errorf("expected Use=%q, got %q", shell, subCmd.Use)
			}
		}
	})

	t.Run("skips config loading", func(t *testing.T) {
		// PersistentPreRunE should be set and return nil
		if cmd.PersistentPreRunE == nil {
			t.Error("PersistentPreRunE not set - completion command may try to load config")
		}

		err := cmd.PersistentPreRunE(cmd, []string{})
		if err != nil {
			t.Errorf("PersistentPreRunE returned error: %v", err)
		}
	})
}
