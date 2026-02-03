package interactive

import (
	"errors"
	"fmt"

	sdkinteractive "github.com/StackEye-IO/stackeye-go-sdk/interactive"
)

// ErrNoOptions is returned when Select or MultiSelect is called with an empty options list.
var ErrNoOptions = errors.New("no options provided")

// selectOptions holds configuration for a single-selection prompt.
type selectOptions struct {
	defaultVal string
	help       string
	pageSize   int
}

// SelectOption configures the behavior of Select.
type SelectOption func(*selectOptions)

// WithSelectDefault sets the default selection (must be in the options list).
func WithSelectDefault(d string) SelectOption {
	return func(o *selectOptions) {
		o.defaultVal = d
	}
}

// WithSelectHelp sets optional help text shown when the user presses '?'.
func WithSelectHelp(help string) SelectOption {
	return func(o *selectOptions) {
		o.help = help
	}
}

// WithSelectPageSize sets how many options are visible at once.
func WithSelectPageSize(size int) SelectOption {
	return func(o *selectOptions) {
		o.pageSize = size
	}
}

// Select prompts the user to choose a single option from a list.
//
// The prompt is bypassed when the --no-input global flag is active,
// returning the default value if set, or the first option otherwise.
//
// Returns the selected option string, or an error if the prompt fails.
// Returns ErrCancelled if the user presses Ctrl+C.
// Returns ErrNoOptions if the options slice is empty.
func Select(message string, options []string, opts ...SelectOption) (string, error) {
	if len(options) == 0 {
		return "", ErrNoOptions
	}

	o := &selectOptions{}
	for _, opt := range opts {
		opt(o)
	}

	// Bypass prompt if --no-input is active
	if noInputGetter != nil && noInputGetter() {
		if o.defaultVal != "" {
			return o.defaultVal, nil
		}
		return options[0], nil //nolint:gosec // len(options) == 0 is checked above
	}

	sdkOpts := &sdkinteractive.SelectPromptOptions{
		Message: message,
		Options: options,
		Default: o.defaultVal,
		Help:    o.help,
	}
	if o.pageSize > 0 {
		sdkOpts.PageSize = o.pageSize
	}

	selected, err := sdkinteractive.AskSelect(sdkOpts)
	if err != nil {
		if errors.Is(err, sdkinteractive.ErrPromptCancelled) {
			return "", ErrCancelled
		}
		return "", fmt.Errorf("failed to prompt for selection: %w", err)
	}

	return selected, nil
}

// multiSelectOptions holds configuration for a multi-selection prompt.
type multiSelectOptions struct {
	defaults []string
	help     string
	pageSize int
	validate func([]string) error
}

// MultiSelectOption configures the behavior of MultiSelect.
type MultiSelectOption func(*multiSelectOptions)

// WithMultiSelectDefaults sets the options pre-selected by default.
func WithMultiSelectDefaults(defaults []string) MultiSelectOption {
	return func(o *multiSelectOptions) {
		o.defaults = defaults
	}
}

// WithMultiSelectHelp sets optional help text shown when the user presses '?'.
func WithMultiSelectHelp(help string) MultiSelectOption {
	return func(o *multiSelectOptions) {
		o.help = help
	}
}

// WithMultiSelectPageSize sets how many options are visible at once.
func WithMultiSelectPageSize(size int) MultiSelectOption {
	return func(o *multiSelectOptions) {
		o.pageSize = size
	}
}

// WithMultiSelectValidate sets a validation function for the selections.
func WithMultiSelectValidate(fn func([]string) error) MultiSelectOption {
	return func(o *multiSelectOptions) {
		o.validate = fn
	}
}

// MultiSelect prompts the user to choose multiple options from a list.
//
// The prompt is bypassed when the --no-input global flag is active,
// returning the defaults if set, or all options otherwise.
//
// Returns the selected options, or an error if the prompt fails.
// Returns ErrCancelled if the user presses Ctrl+C.
// Returns ErrNoOptions if the options slice is empty.
func MultiSelect(message string, options []string, opts ...MultiSelectOption) ([]string, error) {
	if len(options) == 0 {
		return nil, ErrNoOptions
	}

	o := &multiSelectOptions{}
	for _, opt := range opts {
		opt(o)
	}

	// Bypass prompt if --no-input is active
	if noInputGetter != nil && noInputGetter() {
		if len(o.defaults) > 0 {
			return o.defaults, nil
		}
		return options, nil
	}

	sdkOpts := &sdkinteractive.MultiSelectPromptOptions{
		Message:  message,
		Options:  options,
		Defaults: o.defaults,
		Help:     o.help,
		Validate: o.validate,
	}
	if o.pageSize > 0 {
		sdkOpts.PageSize = o.pageSize
	}

	selected, err := sdkinteractive.AskMultiSelect(sdkOpts)
	if err != nil {
		if errors.Is(err, sdkinteractive.ErrPromptCancelled) {
			return nil, ErrCancelled
		}
		return nil, fmt.Errorf("failed to prompt for multi-selection: %w", err)
	}

	return selected, nil
}
