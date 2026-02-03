package output

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// testProbe is a sample struct for testing output formatting.
type testProbe struct {
	ID     string `json:"id" yaml:"id" table:"ID"`
	Name   string `json:"name" yaml:"name" table:"Name"`
	URL    string `json:"url" yaml:"url" table:"URL"`
	Status string `json:"status" yaml:"status" table:"Status"`
}

func TestNewPrinter_NilConfig(t *testing.T) {
	p := NewPrinter(nil)
	if p == nil {
		t.Fatal("NewPrinter(nil) returned nil")
	}
	if p.formatter == nil {
		t.Fatal("Printer has nil formatter")
	}
	// Default format should be table
	if p.Format() != sdkoutput.FormatTable {
		t.Errorf("expected default format Table, got %v", p.Format())
	}
}

func TestNewPrinter_WithConfig(t *testing.T) {
	tests := []struct {
		name           string
		outputFormat   config.OutputFormat
		expectedFormat sdkoutput.Format
	}{
		{"table format", config.OutputFormatTable, sdkoutput.FormatTable},
		{"json format", config.OutputFormatJSON, sdkoutput.FormatJSON},
		{"yaml format", config.OutputFormatYAML, sdkoutput.FormatYAML},
		{"wide format", config.OutputFormatWide, sdkoutput.FormatWide},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Preferences: &config.Preferences{
					OutputFormat: tt.outputFormat,
					Color:        config.ColorModeNever,
				},
			}
			p := NewPrinter(cfg)
			if p.Format() != tt.expectedFormat {
				t.Errorf("expected format %v, got %v", tt.expectedFormat, p.Format())
			}
		})
	}
}

func TestPrinter_Print_TableFormat(t *testing.T) {
	var buf bytes.Buffer
	opts := sdkoutput.DefaultOptions().
		WithFormat(sdkoutput.FormatTable).
		WithWriter(&buf)
	p := NewPrinterWithOptions(opts)

	probes := []testProbe{
		{ID: "p1", Name: "API Check", URL: "https://api.example.com", Status: "up"},
		{ID: "p2", Name: "Web Check", URL: "https://www.example.com", Status: "down"},
	}

	if err := p.Print(probes); err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	output := buf.String()
	// Table format should contain headers and data
	if !strings.Contains(output, "ID") {
		t.Error("expected output to contain 'ID' header")
	}
	if !strings.Contains(output, "API Check") {
		t.Error("expected output to contain 'API Check'")
	}
	if !strings.Contains(output, "p1") {
		t.Error("expected output to contain 'p1'")
	}
}

func TestPrinter_Print_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	opts := sdkoutput.DefaultOptions().
		WithFormat(sdkoutput.FormatJSON).
		WithWriter(&buf)
	p := NewPrinterWithOptions(opts)

	probes := []testProbe{
		{ID: "p1", Name: "API Check", URL: "https://api.example.com", Status: "up"},
	}

	if err := p.Print(probes); err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	output := buf.String()
	// JSON format should contain JSON structure
	if !strings.Contains(output, `"id":`) {
		t.Error("expected JSON output to contain '\"id\":'")
	}
	if !strings.Contains(output, `"p1"`) {
		t.Error("expected JSON output to contain '\"p1\"'")
	}
}

func TestPrinter_Print_YAMLFormat(t *testing.T) {
	var buf bytes.Buffer
	opts := sdkoutput.DefaultOptions().
		WithFormat(sdkoutput.FormatYAML).
		WithWriter(&buf)
	p := NewPrinterWithOptions(opts)

	probes := []testProbe{
		{ID: "p1", Name: "API Check", URL: "https://api.example.com", Status: "up"},
	}

	if err := p.Print(probes); err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	output := buf.String()
	// YAML format should contain YAML structure
	if !strings.Contains(output, "id:") {
		t.Error("expected YAML output to contain 'id:'")
	}
	if !strings.Contains(output, "p1") {
		t.Error("expected YAML output to contain 'p1'")
	}
}

func TestPrinter_Print_EmptySlice(t *testing.T) {
	var buf bytes.Buffer
	opts := sdkoutput.DefaultOptions().
		WithFormat(sdkoutput.FormatTable).
		WithWriter(&buf)
	p := NewPrinterWithOptions(opts)

	var probes []testProbe // Empty slice

	if err := p.Print(probes); err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	// Empty slice should produce no output
	if buf.Len() != 0 {
		t.Errorf("expected empty output for empty slice, got: %q", buf.String())
	}
}

func TestPrinter_Print_NilData(t *testing.T) {
	var buf bytes.Buffer
	opts := sdkoutput.DefaultOptions().
		WithFormat(sdkoutput.FormatTable).
		WithWriter(&buf)
	p := NewPrinterWithOptions(opts)

	if err := p.Print(nil); err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	// Nil data should produce no output
	if buf.Len() != 0 {
		t.Errorf("expected empty output for nil data, got: %q", buf.String())
	}
}

func TestPrinter_PrintError_TableFormat(t *testing.T) {
	var buf bytes.Buffer
	opts := sdkoutput.DefaultOptions().
		WithFormat(sdkoutput.FormatTable)
	opts.ErrWriter = &buf
	p := NewPrinterWithOptions(opts)

	testErr := errors.New("connection refused")
	if err := p.PrintError(testErr); err != nil {
		t.Fatalf("PrintError failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Error:") {
		t.Error("expected error output to contain 'Error:'")
	}
	if !strings.Contains(output, "connection refused") {
		t.Error("expected error output to contain 'connection refused'")
	}
}

func TestPrinter_PrintError_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	opts := sdkoutput.DefaultOptions().
		WithFormat(sdkoutput.FormatJSON)
	opts.ErrWriter = &buf
	p := NewPrinterWithOptions(opts)

	testErr := errors.New("api error")
	if err := p.PrintError(testErr); err != nil {
		t.Fatalf("PrintError failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"error":`) {
		t.Error("expected JSON error output to contain '\"error\":'")
	}
	if !strings.Contains(output, "api error") {
		t.Error("expected JSON error output to contain 'api error'")
	}
}

func TestPrinter_PrintError_NilError(t *testing.T) {
	var buf bytes.Buffer
	opts := sdkoutput.DefaultOptions()
	opts.ErrWriter = &buf
	p := NewPrinterWithOptions(opts)

	if err := p.PrintError(nil); err != nil {
		t.Fatalf("PrintError(nil) failed: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected no output for nil error, got: %q", buf.String())
	}
}

func TestPrinter_PrintEmpty_TableFormat(t *testing.T) {
	var buf bytes.Buffer
	opts := sdkoutput.DefaultOptions().
		WithFormat(sdkoutput.FormatTable).
		WithWriter(&buf)
	p := NewPrinterWithOptions(opts)

	if err := p.PrintEmpty("No probes found"); err != nil {
		t.Fatalf("PrintEmpty failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No probes found") {
		t.Errorf("expected output to contain message, got: %q", output)
	}
}

func TestPrinter_PrintEmpty_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	opts := sdkoutput.DefaultOptions().
		WithFormat(sdkoutput.FormatJSON).
		WithWriter(&buf)
	p := NewPrinterWithOptions(opts)

	if err := p.PrintEmpty("No probes found"); err != nil {
		t.Fatalf("PrintEmpty failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[]") {
		t.Errorf("expected JSON empty output to contain '[]', got: %q", output)
	}
}

func TestPrintIfNotEmpty_WithData(t *testing.T) {
	// Save and restore global config getter
	oldGetter := loadConfigGetter()
	defer func() { storeConfigGetter(oldGetter) }()

	// Set up a config getter that returns JSON format
	storeConfigGetter(func() *config.Config {
		return &config.Config{
			Preferences: &config.Preferences{
				OutputFormat: config.OutputFormatJSON,
				Color:        config.ColorModeNever,
			},
		}
	})

	probes := []testProbe{
		{ID: "p1", Name: "Test", URL: "https://test.com", Status: "up"},
	}

	// Should not error with non-empty data
	if err := PrintIfNotEmpty(probes, "No probes"); err != nil {
		t.Fatalf("PrintIfNotEmpty failed: %v", err)
	}
}

func TestPrintIfNotEmpty_EmptySlice(t *testing.T) {
	// Save and restore global config getter
	oldGetter := loadConfigGetter()
	defer func() { storeConfigGetter(oldGetter) }()

	// Set up a config getter
	storeConfigGetter(func() *config.Config {
		return &config.Config{
			Preferences: &config.Preferences{
				OutputFormat: config.OutputFormatTable,
				Color:        config.ColorModeNever,
			},
		}
	})

	var probes []testProbe

	// Should not error with empty data
	if err := PrintIfNotEmpty(probes, "No probes found"); err != nil {
		t.Fatalf("PrintIfNotEmpty failed: %v", err)
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		expected bool
	}{
		{"nil", nil, true},
		{"empty slice", []string{}, true},
		{"non-empty slice", []string{"a"}, false},
		{"empty array", [0]int{}, true},
		{"non-empty array", [1]int{1}, false},
		{"nil pointer to slice", (*[]string)(nil), true},
		{"struct", testProbe{}, false},
		{"empty map", map[string]any{}, true},
		{"non-empty map", map[string]any{"key": "value"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmpty(tt.data)
			if result != tt.expected {
				t.Errorf("isEmpty(%v) = %v, expected %v", tt.data, result, tt.expected)
			}
		})
	}
}

func TestGlobalFunctions_NoConfigGetter(t *testing.T) {
	// Save and restore global config getter
	oldGetter := loadConfigGetter()
	defer func() { storeConfigGetter(oldGetter) }()

	// Clear config getter to simulate uninitialized state
	storeConfigGetter(nil)

	// Global functions should still work with defaults
	if err := Print(nil); err != nil {
		t.Fatalf("Print failed with no config getter: %v", err)
	}

	if err := PrintError(nil); err != nil {
		t.Fatalf("PrintError failed with no config getter: %v", err)
	}

	if err := PrintEmpty("test"); err != nil {
		t.Fatalf("PrintEmpty failed with no config getter: %v", err)
	}
}

func TestSetConfigGetter(t *testing.T) {
	// Save and restore global config getter
	oldGetter := loadConfigGetter()
	defer func() { storeConfigGetter(oldGetter) }()

	called := false
	testGetter := func() *config.Config {
		called = true
		return nil
	}

	SetConfigGetter(testGetter)

	// Trigger a function that uses the getter
	_ = getPrinter()

	if !called {
		t.Error("config getter was not called after SetConfigGetter")
	}
}

func TestPrinter_SetWriter(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	opts := sdkoutput.DefaultOptions().
		WithFormat(sdkoutput.FormatJSON).
		WithWriter(&buf1)
	p := NewPrinterWithOptions(opts)

	// Print to initial writer
	probes := []testProbe{{ID: "p1", Name: "Test", URL: "https://test.com", Status: "up"}}
	if err := p.Print(probes); err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	if buf1.Len() == 0 {
		t.Error("expected output in buf1")
	}

	// Change writer and print again
	p.SetWriter(&buf2)
	if err := p.Print(probes); err != nil {
		t.Fatalf("Print after SetWriter failed: %v", err)
	}

	if buf2.Len() == 0 {
		t.Error("expected output in buf2 after SetWriter")
	}

	// PrintEmpty should also use the new writer
	buf2.Reset()
	if err := p.PrintEmpty("No data"); err != nil {
		t.Fatalf("PrintEmpty after SetWriter failed: %v", err)
	}

	if !strings.Contains(buf2.String(), "[]") {
		t.Errorf("expected PrintEmpty to write to buf2, got: %q", buf2.String())
	}
}

func TestPrinter_SetErrWriter(t *testing.T) {
	var errBuf bytes.Buffer
	opts := sdkoutput.DefaultOptions().
		WithFormat(sdkoutput.FormatTable)
	opts.ErrWriter = &errBuf
	p := NewPrinterWithOptions(opts)

	p.SetErrWriter(&errBuf)

	// Verify the field was set (internal state change)
	// The actual error output goes through the SDK formatter
	if p.errWriter != &errBuf {
		t.Error("SetErrWriter did not update the errWriter field")
	}
}

func TestPrinter_PrintEmpty_YAMLFormat(t *testing.T) {
	var buf bytes.Buffer
	opts := sdkoutput.DefaultOptions().
		WithFormat(sdkoutput.FormatYAML).
		WithWriter(&buf)
	p := NewPrinterWithOptions(opts)

	if err := p.PrintEmpty("No probes found"); err != nil {
		t.Fatalf("PrintEmpty YAML failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[]") {
		t.Errorf("expected YAML empty output to contain '[]', got: %q", output)
	}
}
