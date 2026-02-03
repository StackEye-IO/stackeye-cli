// Package output provides CLI output helpers.
package output

import (
	"os"

	"github.com/mattn/go-isatty"
)

// IsPiped returns true if stdout is not a terminal (i.e., output is being
// piped or redirected to a file). When piped, commands should prefer
// machine-readable output and disable colors.
func IsPiped() bool {
	fd := os.Stdout.Fd()
	return !isatty.IsTerminal(fd) && !isatty.IsCygwinTerminal(fd)
}

// IsStderrPiped returns true if stderr is not a terminal. When stderr is
// piped, interactive elements like spinners and progress bars should be
// disabled since they use ANSI escape sequences that corrupt piped output.
func IsStderrPiped() bool {
	fd := os.Stderr.Fd()
	return !isatty.IsTerminal(fd) && !isatty.IsCygwinTerminal(fd)
}

// IsDumbTerminal returns true if the TERM environment variable is set to
// "dumb", indicating a terminal that does not support ANSI escape sequences.
// Colors and animations should be disabled for dumb terminals.
func IsDumbTerminal() bool {
	return os.Getenv("TERM") == "dumb"
}

// IsInteractive returns true if the CLI is running in an interactive context
// where colors, spinners, prompts, and other interactive features are
// appropriate. It checks (in order):
//
//  1. stdout is a TTY (not piped/redirected)
//  2. TERM is not "dumb"
//  3. --no-input flag is not set
//  4. STACKEYE_NO_INPUT environment variable is not set
//
// When IsInteractive returns false, commands should:
//   - Disable colors (unless --color=always)
//   - Disable spinners and progress bars
//   - Skip interactive prompts (use defaults or error)
//   - Use machine-readable output when appropriate
func IsInteractive() bool {
	if IsPiped() {
		return false
	}

	if IsDumbTerminal() {
		return false
	}

	if noInputGetter != nil && noInputGetter() {
		return false
	}

	if v, ok := os.LookupEnv("STACKEYE_NO_INPUT"); ok && v != "0" && v != "" {
		return false
	}

	return true
}
