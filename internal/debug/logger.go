// Package debug provides structured debug logging for the StackEye CLI.
//
// Debug logging is enabled via the --debug flag (shorthand for --v=6) or
// STACKEYE_DEBUG environment variable. Output is written to stderr to keep
// stdout clean for machine-parseable command output.
//
// The package uses Go's log/slog for structured logging, matching the backend's
// logging approach. Verbosity levels follow kubectl conventions (0-10):
//
//   - 0: Errors only (default, no debug output)
//   - 1: Warnings
//   - 2: Info messages
//   - 3: Extended info (config/context details)
//   - 4: Debug messages (internal flow)
//   - 5: HTTP requests (method, URL, duration)
//   - 6: HTTP responses (+ status code, body size) [--debug shorthand]
//   - 7: Request headers (redacted)
//   - 8: Response headers
//   - 9: Full bodies (truncated at 10KB)
//   - 10: Trace (curl equivalent, timing breakdown)
//
// Usage:
//
//	// In root.go init():
//	debug.SetVerbosityGetter(GetVerbosity)
//
//	// In any package:
//	debug.Log(3, "loading config", "path", configPath)
//	debug.Logf(4, "resolved context: %s", contextName)
package debug

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// verbosityGetter returns the current verbosity level.
type verbosityGetter func() int

var (
	mu                     sync.RWMutex
	currentVerbosityGetter verbosityGetter
	logger                 *slog.Logger
	writer                 io.Writer = os.Stderr
)

func init() {
	// Initialize with a no-op logger; replaced when SetVerbosityGetter is called.
	logger = slog.New(slog.NewTextHandler(io.Discard, nil))
}

// SetVerbosityGetter sets the function used to retrieve the verbosity level.
// This should be called during CLI initialization (root.go init).
func SetVerbosityGetter(getter func() int) {
	mu.Lock()
	defer mu.Unlock()
	currentVerbosityGetter = getter
	rebuildLogger()
}

// SetWriter overrides the output writer (default: os.Stderr).
// Intended for testing.
func SetWriter(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	writer = w
	rebuildLogger()
}

// rebuildLogger creates a new slog.Logger with the current writer.
// Must be called with mu held.
func rebuildLogger() {
	logger = slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// getVerbosity returns the current verbosity level, or 0 if not configured.
func getVerbosity() int {
	mu.RLock()
	getter := currentVerbosityGetter
	mu.RUnlock()
	if getter == nil {
		return 0
	}
	return getter()
}

// Enabled returns true if debug output is active at the given verbosity level.
func Enabled(level int) bool {
	return getVerbosity() >= level
}

// Log emits a structured debug message if the current verbosity is >= level.
// Key-value pairs are passed as slog attributes.
//
// Example:
//
//	debug.Log(3, "config loaded", "path", "/home/user/.config/stackeye/config.yaml")
func Log(level int, msg string, args ...any) {
	if getVerbosity() < level {
		return
	}

	mu.RLock()
	l := logger
	mu.RUnlock()

	l.Debug(msg, append([]any{slog.Int("v", level)}, args...)...)
}

// Logf emits a formatted debug message if the current verbosity is >= level.
//
// Example:
//
//	debug.Logf(4, "resolved context: %s", contextName)
func Logf(level int, format string, args ...any) {
	if getVerbosity() < level {
		return
	}

	mu.RLock()
	l := logger
	mu.RUnlock()

	l.Debug(fmt.Sprintf(format, args...), slog.Int("v", level))
}

// ConfigLoaded logs that configuration was successfully loaded.
// Emits at verbosity level 3 (extended info).
func ConfigLoaded(path string, context string) {
	Log(3, "config loaded", "path", path, "context", context)
}

// ConfigError logs a configuration loading error.
// Emits at verbosity level 1 (warnings).
func ConfigError(path string, err error) {
	Log(1, "config error", "path", path, "error", err.Error())
}

// sensitiveEnvSuffixes lists env var name suffixes that indicate secret values.
var sensitiveEnvSuffixes = []string{"_KEY", "_SECRET", "_TOKEN", "_PASSWORD"}

// EnvOverride logs when an environment variable overrides a config value.
// Values for env vars ending in _KEY, _SECRET, _TOKEN, or _PASSWORD are redacted.
// Emits at verbosity level 3 (extended info).
func EnvOverride(envVar string, value string) {
	display := value
	upper := strings.ToUpper(envVar)
	for _, suffix := range sensitiveEnvSuffixes {
		if strings.HasSuffix(upper, suffix) {
			if len(value) > 4 {
				display = value[:4] + "****"
			} else {
				display = "****"
			}
			break
		}
	}
	Log(3, "env override", "var", envVar, "value", display)
}
