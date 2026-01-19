package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestPowerShellCompletionCmd(t *testing.T) {
	t.Run("generates completion script", func(t *testing.T) {
		cmd := newPowerShellCompletionCmd()

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
			// Verify script contains expected PowerShell completion patterns
			expectedPatterns := []string{
				"Register-ArgumentCompleter",
				"stackeye",
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
		cmd := newPowerShellCompletionCmd()
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

		// Script should still contain core PowerShell completion patterns
		if !strings.Contains(output, "Register-ArgumentCompleter") {
			t.Error("completion script missing Register-ArgumentCompleter")
		}
	})
}

func TestPowerShellCompletionScriptSyntax(t *testing.T) {
	// Skip if pwsh (PowerShell Core) is not available
	pwshPath, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("pwsh not available, skipping syntax validation")
	}

	cmd := newPowerShellCompletionCmd()

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

	// Validate PowerShell syntax using pwsh -Command with the script
	// We use -NoProfile to skip profile loading and -NonInteractive to avoid prompts
	pwshCmd := exec.Command(pwshPath, "-NoProfile", "-NonInteractive", "-Command", "-")
	// Parse script to check syntax without executing
	checkScript := `
$script = @'
` + script + `
'@
$tokens = $null
$errors = $null
[System.Management.Automation.Language.Parser]::ParseInput($script, [ref]$tokens, [ref]$errors)
if ($errors.Count -gt 0) {
    $errors | ForEach-Object { Write-Error $_.Message }
    exit 1
}
exit 0
`
	pwshCmd.Stdin = strings.NewReader(checkScript)
	var stderr bytes.Buffer
	pwshCmd.Stderr = &stderr

	if err := pwshCmd.Run(); err != nil {
		t.Errorf("PowerShell syntax validation failed: %v\nstderr: %s", err, stderr.String())
	}
}

func TestCompletionCmdHasPowerShellSubcommand(t *testing.T) {
	cmd := NewCompletionCmd()

	t.Run("has powershell subcommand", func(t *testing.T) {
		psCmd, _, err := cmd.Find([]string{"powershell"})
		if err != nil {
			t.Fatalf("failed to find powershell subcommand: %v", err)
		}
		if psCmd.Use != "powershell" {
			t.Errorf("expected Use='powershell', got %q", psCmd.Use)
		}
	})

	t.Run("completion long help mentions PowerShell", func(t *testing.T) {
		if !strings.Contains(cmd.Long, "PowerShell") {
			t.Error("completion command Long help should mention PowerShell")
		}
	})
}
