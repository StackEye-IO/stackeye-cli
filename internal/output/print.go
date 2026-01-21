// Package output provides CLI output helpers that bridge the CLI's global flags
// with the SDK's output formatters.
//
// This package simplifies command output by providing convenience functions
// that automatically respect the user's --output and --no-color preferences.
//
// Usage:
//
//	// In a command's RunE function:
//	data, err := api.ListProbes(ctx)
//	if err != nil {
//	    return output.PrintError(err)
//	}
//	if len(data) == 0 {
//	    output.PrintEmpty("No probes found")
//	    return nil
//	}
//	return output.Print(data)
package output

import (
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// configGetter is a function that returns the current CLI configuration.
// This is set by the cmd package to avoid circular imports.
var configGetter func() *config.Config

// SetConfigGetter sets the function used to retrieve the current configuration.
// This should be called once during CLI initialization from the cmd package.
func SetConfigGetter(getter func() *config.Config) {
	configGetter = getter
}

// Printer wraps an SDK formatter and provides CLI-specific output helpers.
type Printer struct {
	formatter sdkoutput.Formatter
	writer    io.Writer
	errWriter io.Writer
}

// NewPrinter creates a new Printer using the given configuration.
// If cfg is nil, default options are used (table format, auto color).
func NewPrinter(cfg *config.Config) *Printer {
	opts := sdkoutput.DefaultOptions()

	if cfg != nil && cfg.Preferences != nil {
		// Map config preferences to formatter options
		opts.Format = sdkoutput.Format(cfg.Preferences.OutputFormat)
		opts.Color = sdkoutput.ColorMode(cfg.Preferences.Color)
	}

	return &Printer{
		formatter: sdkoutput.New(opts),
		writer:    opts.Writer,
		errWriter: opts.ErrWriter,
	}
}

// NewPrinterWithOptions creates a Printer with explicit options.
// This is useful for testing or when fine-grained control is needed.
func NewPrinterWithOptions(opts *sdkoutput.Options) *Printer {
	if opts == nil {
		opts = sdkoutput.DefaultOptions()
	}
	writer := opts.Writer
	if writer == nil {
		writer = os.Stdout
	}
	errWriter := opts.ErrWriter
	if errWriter == nil {
		errWriter = os.Stderr
	}
	return &Printer{
		formatter: sdkoutput.New(opts),
		writer:    writer,
		errWriter: errWriter,
	}
}

// Print formats and outputs data using the configured format.
// Data can be a struct, pointer to struct, or slice of structs.
// Empty slices produce no output (for JSON/YAML) or a blank line (for tables).
func (p *Printer) Print(data any) error {
	return p.formatter.Print(data)
}

// PrintError formats and outputs an error message.
// For table format, outputs "Error: <message>" to stderr.
// For JSON/YAML, outputs {"error": "<message>"} to stderr.
func (p *Printer) PrintError(err error) error {
	if err == nil {
		return nil
	}
	return p.formatter.PrintError(err)
}

// PrintEmpty outputs a user-friendly message when no results are found.
// For table format, prints the message to the configured writer.
// For JSON/YAML, outputs an empty array [].
func (p *Printer) PrintEmpty(message string) error {
	format := p.formatter.Format()

	switch format {
	case sdkoutput.FormatJSON:
		_, err := fmt.Fprintln(p.writer, "[]")
		return err
	case sdkoutput.FormatYAML:
		_, err := fmt.Fprintln(p.writer, "[]")
		return err
	default:
		// Table format: print human-friendly message
		_, err := fmt.Fprintln(p.writer, message)
		return err
	}
}

// Format returns the current output format.
func (p *Printer) Format() sdkoutput.Format {
	return p.formatter.Format()
}

// SetWriter changes the output destination for data.
func (p *Printer) SetWriter(w io.Writer) {
	p.writer = w
	p.formatter.SetWriter(w)
}

// SetErrWriter changes the output destination for errors.
func (p *Printer) SetErrWriter(w io.Writer) {
	p.errWriter = w
}

// --- Global convenience functions ---

// getPrinter returns a Printer configured from the current CLI configuration.
// If no config getter is set or config is nil, returns a default printer.
func getPrinter() *Printer {
	var cfg *config.Config
	if configGetter != nil {
		cfg = configGetter()
	}
	return NewPrinter(cfg)
}

// Print formats and outputs data using the CLI's configured format.
// This is a convenience function that uses the global configuration.
//
// Example:
//
//	probes, err := client.ListProbes(ctx)
//	if err != nil {
//	    return err
//	}
//	return output.Print(probes)
func Print(data any) error {
	return getPrinter().Print(data)
}

// PrintError formats and outputs an error message.
// This is a convenience function that uses the global configuration.
//
// Example:
//
//	if err := client.DeleteProbe(ctx, id); err != nil {
//	    return output.PrintError(err)
//	}
func PrintError(err error) error {
	if err == nil {
		return nil
	}
	return getPrinter().PrintError(err)
}

// PrintEmpty outputs a user-friendly message for empty results.
// This is a convenience function that uses the global configuration.
//
// Example:
//
//	probes, err := client.ListProbes(ctx)
//	if err != nil {
//	    return err
//	}
//	if len(probes) == 0 {
//	    output.PrintEmpty("No probes found. Create one with 'stackeye probe create'")
//	    return nil
//	}
//	return output.Print(probes)
func PrintEmpty(message string) error {
	return getPrinter().PrintEmpty(message)
}

// PrintIfNotEmpty is a convenience function that handles the common pattern
// of printing data if it exists, or an empty message if not.
//
// Example:
//
//	probes, err := client.ListProbes(ctx)
//	if err != nil {
//	    return err
//	}
//	return output.PrintIfNotEmpty(probes, "No probes found")
func PrintIfNotEmpty(data any, emptyMessage string) error {
	if isEmpty(data) {
		return PrintEmpty(emptyMessage)
	}
	return Print(data)
}

// isEmpty checks if data is empty (nil, empty slice, or empty array).
func isEmpty(data any) bool {
	if data == nil {
		return true
	}

	v := reflect.ValueOf(data)

	// Handle pointer
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}

	// Check for empty slice/array/map
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	}

	return false
}
