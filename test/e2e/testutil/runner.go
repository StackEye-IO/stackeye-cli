package testutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// CLIRunner executes CLI commands for E2E testing.
type CLIRunner struct {
	// BinaryPath is the path to the CLI binary.
	BinaryPath string
	// ConfigPath is the path to the config file.
	ConfigPath string
	// XDGConfigHome is the XDG_CONFIG_HOME value to use.
	// This ensures the CLI's Save() function writes to the correct location.
	XDGConfigHome string
	// Env contains additional environment variables.
	Env map[string]string
}

// RunResult contains the result of running a CLI command.
type RunResult struct {
	// Stdout contains the standard output.
	Stdout string
	// Stderr contains the standard error.
	Stderr string
	// ExitCode is the exit code of the command.
	ExitCode int
	// Err is any error that occurred during execution.
	Err error
}

// NewCLIRunner creates a new CLI runner with the given config path and XDG config home.
// It expects the binary to be built and available in the project root or path.
// The xdgConfigHome parameter is set as XDG_CONFIG_HOME environment variable so that
// the CLI's Save() function writes to the correct location (not the user's actual config).
func NewCLIRunner(configPath, xdgConfigHome string) *CLIRunner {
	binaryPath := findBinary()
	fmt.Fprintf(os.Stderr, "[DEBUG RUNNER] BinaryPath: %s\n", binaryPath)
	fmt.Fprintf(os.Stderr, "[DEBUG RUNNER] ConfigPath: %s\n", configPath)
	fmt.Fprintf(os.Stderr, "[DEBUG RUNNER] XDGConfigHome: %s\n", xdgConfigHome)
	return &CLIRunner{
		BinaryPath:    binaryPath,
		ConfigPath:    configPath,
		XDGConfigHome: xdgConfigHome,
		Env:           make(map[string]string),
	}
}

// findBinary locates the stackeye binary.
// It checks environment variable, then computed paths from the source file location.
func findBinary() string {
	// Check environment variable first
	if path := os.Getenv("STACKEYE_CLI_BINARY"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Get the directory of this source file to compute the module root
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		// This file is at test/e2e/testutil/runner.go
		// Module root is 3 directories up
		moduleRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename))))
		binPath := filepath.Join(moduleRoot, "bin", "stackeye")
		if _, err := os.Stat(binPath); err == nil {
			return binPath
		}

		// Also check module root directly
		rootPath := filepath.Join(moduleRoot, "stackeye")
		if _, err := os.Stat(rootPath); err == nil {
			return rootPath
		}
	}

	// Check common locations relative to current working directory
	locations := []string{
		"./stackeye",
		"./bin/stackeye",
		"../../../stackeye",
		"../../../bin/stackeye",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	// Fall back to PATH lookup
	if path, err := exec.LookPath("stackeye"); err == nil {
		return path
	}

	// Return default - tests will fail with clear error if not found
	return "./stackeye"
}

// Run executes the CLI with the given arguments.
func (r *CLIRunner) Run(args ...string) *RunResult {
	// Prepend config flag
	fullArgs := []string{"--config", r.ConfigPath}
	fullArgs = append(fullArgs, args...)

	return r.RunRaw(fullArgs...)
}

// RunRaw executes the CLI with exact arguments (no automatic config flag).
func (r *CLIRunner) RunRaw(args ...string) *RunResult {
	// Debug: verify config file exists before running
	if info, err := os.Stat(r.ConfigPath); err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG RUNNER] Config file check FAILED: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG RUNNER] Config file exists, size=%d\n", info.Size())
	}

	cmd := exec.Command(r.BinaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range r.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Disable color output for consistent parsing
	cmd.Env = append(cmd.Env, "NO_COLOR=1")
	// Enable debug output to see actual errors
	cmd.Env = append(cmd.Env, "STACKEYE_DEBUG=1")
	// Set XDG_CONFIG_HOME so Save() writes to test directory, not user's actual config
	if r.XDGConfigHome != "" {
		cmd.Env = append(cmd.Env, "XDG_CONFIG_HOME="+r.XDGConfigHome)
	}

	fmt.Fprintf(os.Stderr, "[DEBUG RUNNER] Executing: %s %v\n", r.BinaryPath, args)

	err := cmd.Run()

	result := &RunResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
		Err:      err,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}

// RunWithOutput executes and returns stdout, failing on non-zero exit.
func (r *CLIRunner) RunWithOutput(args ...string) (string, error) {
	result := r.Run(args...)
	if result.ExitCode != 0 {
		return "", fmt.Errorf("command failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}
	return result.Stdout, nil
}

// RunExpectError executes and expects a non-zero exit code.
func (r *CLIRunner) RunExpectError(args ...string) *RunResult {
	return r.Run(args...)
}

// Success returns true if the command succeeded.
func (r *RunResult) Success() bool {
	return r.ExitCode == 0
}

// Failed returns true if the command failed.
func (r *RunResult) Failed() bool {
	return r.ExitCode != 0
}

// Contains returns true if stdout contains the given substring.
func (r *RunResult) Contains(substr string) bool {
	return strings.Contains(r.Stdout, substr)
}

// StderrContains returns true if stderr contains the given substring.
func (r *RunResult) StderrContains(substr string) bool {
	return strings.Contains(r.Stderr, substr)
}

// Lines returns stdout split into lines.
func (r *RunResult) Lines() []string {
	lines := strings.Split(strings.TrimSpace(r.Stdout), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}
	}
	return lines
}

// StderrLines returns stderr split into lines.
func (r *RunResult) StderrLines() []string {
	lines := strings.Split(strings.TrimSpace(r.Stderr), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}
	}
	return lines
}

// CombinedOutput returns stdout and stderr combined.
func (r *RunResult) CombinedOutput() string {
	return r.Stdout + r.Stderr
}

// String returns a debug representation of the result.
func (r *RunResult) String() string {
	return fmt.Sprintf("RunResult{ExitCode: %d, Stdout: %q, Stderr: %q}", r.ExitCode, r.Stdout, r.Stderr)
}
