// Package interactive provides CLI-level helpers for interactive user prompts.
//
// These helpers wrap the SDK's interactive package with CLI-specific concerns
// such as --yes flag bypass, --no-input global flag bypass, and consistent
// error handling across all commands.
package interactive

import (
	"errors"
	"fmt"

	sdkinteractive "github.com/StackEye-IO/stackeye-go-sdk/interactive"
)

// ErrCancelled is returned when the user cancels a confirmation prompt (Ctrl+C).
var ErrCancelled = errors.New("operation cancelled by user")

// noInputGetter returns true if interactive prompts should be disabled globally.
// Set via SetNoInputGetter from root command initialization.
var noInputGetter func() bool

// SetNoInputGetter sets the function used to check if --no-input is active.
// Called from root.go init() to wire up the global flag.
func SetNoInputGetter(getter func() bool) {
	noInputGetter = getter
}

// confirmOptions holds configuration for a confirmation prompt.
type confirmOptions struct {
	yes        bool // --yes flag bypasses prompt
	defaultVal bool // default answer when ambiguous (false = safe)
	help       string
}

// ConfirmOption configures the behavior of Confirm.
type ConfirmOption func(*confirmOptions)

// WithYesFlag sets whether the --yes flag was provided, bypassing the prompt.
func WithYesFlag(yes bool) ConfirmOption {
	return func(o *confirmOptions) {
		o.yes = yes
	}
}

// WithDefault sets the default answer when the user presses Enter without input.
// Defaults to false (No) for safety on destructive operations.
func WithDefault(d bool) ConfirmOption {
	return func(o *confirmOptions) {
		o.defaultVal = d
	}
}

// WithHelp sets optional help text shown when the user presses '?'.
func WithHelp(help string) ConfirmOption {
	return func(o *confirmOptions) {
		o.help = help
	}
}

// Confirm prompts the user for a yes/no confirmation.
//
// The prompt is bypassed (returns true) when:
//   - The --yes flag is set via WithYesFlag(true)
//   - The --no-input global flag is active
//
// Returns true if confirmed, false if denied, or an error if the prompt fails.
// Returns ErrCancelled if the user presses Ctrl+C.
func Confirm(message string, opts ...ConfirmOption) (bool, error) {
	o := &confirmOptions{
		defaultVal: false, // Default to No (safe option)
	}
	for _, opt := range opts {
		opt(o)
	}

	// Bypass prompt if --yes flag or --no-input is active
	if o.yes {
		return true, nil
	}
	if noInputGetter != nil && noInputGetter() {
		return true, nil
	}

	confirmed, err := sdkinteractive.AskConfirm(&sdkinteractive.ConfirmPromptOptions{
		Message: message,
		Default: o.defaultVal,
		Help:    o.help,
	})
	if err != nil {
		if errors.Is(err, sdkinteractive.ErrPromptCancelled) {
			return false, ErrCancelled
		}
		return false, fmt.Errorf("failed to prompt for confirmation: %w", err)
	}

	return confirmed, nil
}
