package interactive

import (
	"errors"
	"fmt"

	sdkinteractive "github.com/StackEye-IO/stackeye-go-sdk/interactive"
)

// inputOptions holds configuration for a text input prompt.
type inputOptions struct {
	defaultVal string
	help       string
	validate   func(string) error
}

// InputOption configures the behavior of Input.
type InputOption func(*inputOptions)

// WithInputDefault sets the default value when the user presses Enter without input.
func WithInputDefault(d string) InputOption {
	return func(o *inputOptions) {
		o.defaultVal = d
	}
}

// WithInputHelp sets optional help text shown when the user presses '?'.
func WithInputHelp(help string) InputOption {
	return func(o *inputOptions) {
		o.help = help
	}
}

// WithInputValidate sets a validation function for the input.
// Return nil if valid, or an error describing the issue.
func WithInputValidate(fn func(string) error) InputOption {
	return func(o *inputOptions) {
		o.validate = fn
	}
}

// Input prompts the user for text input.
//
// The prompt is bypassed when the --no-input global flag is active,
// returning the default value if set, or an empty string otherwise.
//
// Returns the user's input string, or an error if the prompt fails.
// Returns ErrCancelled if the user presses Ctrl+C.
func Input(message string, opts ...InputOption) (string, error) {
	o := &inputOptions{}
	for _, opt := range opts {
		opt(o)
	}

	// Bypass prompt if --no-input is active
	if noInputGetter != nil && noInputGetter() {
		return o.defaultVal, nil
	}

	result, err := sdkinteractive.AskString(&sdkinteractive.StringPromptOptions{
		Message:  message,
		Default:  o.defaultVal,
		Help:     o.help,
		Validate: o.validate,
	})
	if err != nil {
		if errors.Is(err, sdkinteractive.ErrPromptCancelled) {
			return "", ErrCancelled
		}
		return "", fmt.Errorf("failed to prompt for input: %w", err)
	}

	return result, nil
}

// passwordOptions holds configuration for a password input prompt.
type passwordOptions struct {
	help     string
	validate func(string) error
}

// PasswordOption configures the behavior of Password.
type PasswordOption func(*passwordOptions)

// WithPasswordHelp sets optional help text shown when the user presses '?'.
func WithPasswordHelp(help string) PasswordOption {
	return func(o *passwordOptions) {
		o.help = help
	}
}

// WithPasswordValidate sets a validation function for the password.
// Return nil if valid, or an error describing the issue.
func WithPasswordValidate(fn func(string) error) PasswordOption {
	return func(o *passwordOptions) {
		o.validate = fn
	}
}

// Password prompts the user for a password (input is masked).
//
// The prompt is bypassed when the --no-input global flag is active,
// returning an empty string. Passwords have no default value for security.
//
// Returns the user's password string, or an error if the prompt fails.
// Returns ErrCancelled if the user presses Ctrl+C.
func Password(message string, opts ...PasswordOption) (string, error) {
	o := &passwordOptions{}
	for _, opt := range opts {
		opt(o)
	}

	// Bypass prompt if --no-input is active (no default for passwords)
	if noInputGetter != nil && noInputGetter() {
		return "", nil
	}

	result, err := sdkinteractive.AskPassword(&sdkinteractive.PasswordPromptOptions{
		Message:  message,
		Help:     o.help,
		Validate: o.validate,
	})
	if err != nil {
		if errors.Is(err, sdkinteractive.ErrPromptCancelled) {
			return "", ErrCancelled
		}
		return "", fmt.Errorf("failed to prompt for password: %w", err)
	}

	return result, nil
}
