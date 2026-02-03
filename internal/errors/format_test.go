package errors

import (
	"bytes"
	"strings"
	"testing"

	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

func TestNewErrorFormatter(t *testing.T) {
	f := NewErrorFormatter()
	if f == nil {
		t.Fatal("NewErrorFormatter returned nil")
	}
	if f.cm == nil {
		t.Error("ErrorFormatter.cm is nil")
	}
	if f.writer == nil {
		t.Error("ErrorFormatter.writer is nil")
	}
}

func TestNewErrorFormatterWithWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	f := NewErrorFormatterWithWriter(buf)

	if f == nil {
		t.Fatal("NewErrorFormatterWithWriter returned nil")
	}
	if f.writer != buf {
		t.Error("ErrorFormatter.writer does not match provided buffer")
	}
}

func TestNewErrorFormatterWithColorManager(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	if f == nil {
		t.Fatal("NewErrorFormatterWithColorManager returned nil")
	}
	if f.cm != cm {
		t.Error("ErrorFormatter.cm does not match provided ColorManager")
	}
	if f.writer != buf {
		t.Error("ErrorFormatter.writer does not match provided buffer")
	}
}

func TestPrintError(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	f.PrintError("something went wrong")

	output := buf.String()
	if !strings.Contains(output, "Error:") {
		t.Errorf("PrintError output missing 'Error:' prefix, got: %q", output)
	}
	if !strings.Contains(output, "something went wrong") {
		t.Errorf("PrintError output missing message, got: %q", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("PrintError output missing trailing newline, got: %q", output)
	}
}

func TestPrintWarning(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	f.PrintWarning("this might be a problem")

	output := buf.String()
	if !strings.Contains(output, "Warning:") {
		t.Errorf("PrintWarning output missing 'Warning:' prefix, got: %q", output)
	}
	if !strings.Contains(output, "this might be a problem") {
		t.Errorf("PrintWarning output missing message, got: %q", output)
	}
}

func TestPrintHint(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	f.PrintHint("try running 'stackeye login'")

	output := buf.String()
	if !strings.HasPrefix(output, "  ") {
		t.Errorf("PrintHint output not indented, got: %q", output)
	}
	if !strings.Contains(output, "try running 'stackeye login'") {
		t.Errorf("PrintHint output missing message, got: %q", output)
	}
}

func TestPrintContext(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	f.PrintContext("Probe ID", "abc123")

	output := buf.String()
	if !strings.HasPrefix(output, "  ") {
		t.Errorf("PrintContext output not indented, got: %q", output)
	}
	if !strings.Contains(output, "Probe ID:") {
		t.Errorf("PrintContext output missing key, got: %q", output)
	}
	if !strings.Contains(output, "abc123") {
		t.Errorf("PrintContext output missing value, got: %q", output)
	}
}

func TestPrintSuggestion(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	f.PrintSuggestion("Run 'stackeye probe list' to see available probes")

	output := buf.String()
	if !strings.HasPrefix(output, "  ") {
		t.Errorf("PrintSuggestion output not indented, got: %q", output)
	}
	if !strings.Contains(output, "Suggestion:") {
		t.Errorf("PrintSuggestion output missing 'Suggestion:' prefix, got: %q", output)
	}
	if !strings.Contains(output, "stackeye probe list") {
		t.Errorf("PrintSuggestion output missing suggestion text, got: %q", output)
	}
}

func TestPrintRequestID(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	f.PrintRequestID("req_xyz789")

	output := buf.String()
	if !strings.HasPrefix(output, "  ") {
		t.Errorf("PrintRequestID output not indented, got: %q", output)
	}
	if !strings.Contains(output, "Request ID:") {
		t.Errorf("PrintRequestID output missing 'Request ID:' label, got: %q", output)
	}
	if !strings.Contains(output, "req_xyz789") {
		t.Errorf("PrintRequestID output missing request ID, got: %q", output)
	}
	if !strings.Contains(output, "support") {
		t.Errorf("PrintRequestID output missing support hint, got: %q", output)
	}
}

func TestPrintRequestID_Empty(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	f.PrintRequestID("")

	if buf.Len() != 0 {
		t.Errorf("PrintRequestID with empty ID should produce no output, got: %q", buf.String())
	}
}

func TestPrintValidationErrors(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	f.PrintValidationErrors(map[string]string{
		"name": "name is required",
		"url":  "invalid URL format",
	})

	output := buf.String()
	if !strings.Contains(output, "name:") {
		t.Errorf("PrintValidationErrors output missing 'name' field, got: %q", output)
	}
	if !strings.Contains(output, "name is required") {
		t.Errorf("PrintValidationErrors output missing name message, got: %q", output)
	}
	if !strings.Contains(output, "url:") {
		t.Errorf("PrintValidationErrors output missing 'url' field, got: %q", output)
	}
	if !strings.Contains(output, "invalid URL format") {
		t.Errorf("PrintValidationErrors output missing url message, got: %q", output)
	}
}

func TestFormatError(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	result := f.FormatError("test error")

	if !strings.Contains(result, "Error:") {
		t.Errorf("FormatError missing 'Error:' prefix, got: %q", result)
	}
	if !strings.Contains(result, "test error") {
		t.Errorf("FormatError missing message, got: %q", result)
	}
}

func TestFormatWarning(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	result := f.FormatWarning("test warning")

	if !strings.Contains(result, "Warning:") {
		t.Errorf("FormatWarning missing 'Warning:' prefix, got: %q", result)
	}
	if !strings.Contains(result, "test warning") {
		t.Errorf("FormatWarning missing message, got: %q", result)
	}
}

func TestFormatSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	result := f.FormatSuccess("operation completed")

	if !strings.Contains(result, "Success:") {
		t.Errorf("FormatSuccess missing 'Success:' prefix, got: %q", result)
	}
	if !strings.Contains(result, "operation completed") {
		t.Errorf("FormatSuccess missing message, got: %q", result)
	}
}

func TestFormatInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	result := f.FormatInfo("some information")

	if !strings.Contains(result, "Info:") {
		t.Errorf("FormatInfo missing 'Info:' prefix, got: %q", result)
	}
	if !strings.Contains(result, "some information") {
		t.Errorf("FormatInfo missing message, got: %q", result)
	}
}

func TestGetSuggestion(t *testing.T) {
	tests := []struct {
		errorType string
		wantEmpty bool
		contains  string
	}{
		{"auth_required", false, "login"},
		{"AUTH_REQUIRED", false, "login"}, // Case insensitive
		{"probe_not_found", false, "probe list"},
		{"rate_limited", false, "wait"},
		{"plan_limit", false, "billing"},
		{"server_error", false, "status.stackeye.io"},
		{"unknown_error_type", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.errorType, func(t *testing.T) {
			result := GetSuggestion(tt.errorType)
			if tt.wantEmpty {
				if result != "" {
					t.Errorf("GetSuggestion(%q) = %q, want empty", tt.errorType, result)
				}
			} else {
				if result == "" {
					t.Errorf("GetSuggestion(%q) returned empty string", tt.errorType)
				}
				if tt.contains != "" && !strings.Contains(strings.ToLower(result), strings.ToLower(tt.contains)) {
					t.Errorf("GetSuggestion(%q) = %q, want to contain %q", tt.errorType, result, tt.contains)
				}
			}
		})
	}
}

func TestGetUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		errorCode  string
		defaultMsg string
		want       string
	}{
		{"unauthorized", "", "Authentication required."},
		{"UNAUTHORIZED", "", "Authentication required."}, // Case insensitive
		{"not_found", "", "Resource not found."},
		{"rate_limited", "", "Rate limit exceeded."},
		{"unknown_code", "fallback message", "fallback message"},
		{"unknown_code", "", "An unexpected error occurred."},
	}

	for _, tt := range tests {
		t.Run(tt.errorCode, func(t *testing.T) {
			result := GetUserFriendlyMessage(tt.errorCode, tt.defaultMsg)
			if result != tt.want {
				t.Errorf("GetUserFriendlyMessage(%q, %q) = %q, want %q",
					tt.errorCode, tt.defaultMsg, result, tt.want)
			}
		})
	}
}

func TestErrorWithContext_Print(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	e := &ErrorWithContext{
		Message: "probe not found",
		Context: map[string]string{
			"Probe ID": "abc123",
		},
		Hint:      "Run 'stackeye probe list' to see available probes",
		RequestID: "req_xyz789",
	}

	e.Print(f)

	output := buf.String()
	if !strings.Contains(output, "probe not found") {
		t.Errorf("ErrorWithContext.Print missing message, got: %q", output)
	}
	if !strings.Contains(output, "Probe ID:") {
		t.Errorf("ErrorWithContext.Print missing context key, got: %q", output)
	}
	if !strings.Contains(output, "abc123") {
		t.Errorf("ErrorWithContext.Print missing context value, got: %q", output)
	}
	if !strings.Contains(output, "Suggestion:") {
		t.Errorf("ErrorWithContext.Print missing suggestion, got: %q", output)
	}
	if !strings.Contains(output, "Request ID:") {
		t.Errorf("ErrorWithContext.Print missing request ID, got: %q", output)
	}
}

func TestErrorWithContext_Print_MinimalFields(t *testing.T) {
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	e := &ErrorWithContext{
		Message: "something went wrong",
	}

	e.Print(f)

	output := buf.String()
	if !strings.Contains(output, "something went wrong") {
		t.Errorf("ErrorWithContext.Print missing message, got: %q", output)
	}
	// Should not contain optional fields
	if strings.Contains(output, "Suggestion:") {
		t.Errorf("ErrorWithContext.Print should not have suggestion without Hint, got: %q", output)
	}
	if strings.Contains(output, "Request ID:") {
		t.Errorf("ErrorWithContext.Print should not have request ID without RequestID, got: %q", output)
	}
}

func TestColoredOutput(t *testing.T) {
	// Test that colors are applied when ColorAlways is used
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorAlways)
	f := NewErrorFormatterWithColorManager(cm, buf)

	f.PrintError("test error")

	output := buf.String()
	// ANSI escape codes start with \x1b[
	if !strings.Contains(output, "\x1b[") {
		t.Errorf("ColorAlways mode should produce ANSI escape codes, got: %q", output)
	}
}

func TestNoColorOutput(t *testing.T) {
	// Test that no colors are applied when ColorNever is used
	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	f.PrintError("test error")

	output := buf.String()
	// Should not contain ANSI escape codes
	if strings.Contains(output, "\x1b[") {
		t.Errorf("ColorNever mode should not produce ANSI escape codes, got: %q", output)
	}
}

func TestCommonSuggestions_Coverage(t *testing.T) {
	// Verify that key error types have suggestions
	requiredKeys := []string{
		"auth_required",
		"forbidden",
		"probe_not_found",
		"rate_limited",
		"plan_limit",
		"server_error",
		"connection_refused",
		"network_error",
		"invalid_url",
		"config_not_found",
	}

	for _, key := range requiredKeys {
		if _, ok := CommonSuggestions[key]; !ok {
			t.Errorf("CommonSuggestions missing required key: %q", key)
		}
	}
}

func TestAPIErrorMessages_Coverage(t *testing.T) {
	// Verify that key error codes have mappings
	requiredCodes := []string{
		"unauthorized",
		"forbidden",
		"not_found",
		"rate_limited",
		"plan_limit_exceeded",
		"validation",
		"internal_server",
		"service_unavailable",
	}

	for _, code := range requiredCodes {
		if _, ok := APIErrorMessages[code]; !ok {
			t.Errorf("APIErrorMessages missing required code: %q", code)
		}
	}
}

func TestFullErrorFormat(t *testing.T) {
	// Test the complete error format as specified in the task
	// Expected format:
	// Error: probe not found
	//   Probe ID: abc123
	//   Suggestion: Run 'stackeye probe list' to see available probes
	//   Request ID: req_xyz789

	buf := &bytes.Buffer{}
	cm := sdkoutput.NewColorManager(sdkoutput.ColorNever)
	f := NewErrorFormatterWithColorManager(cm, buf)

	f.PrintError("probe not found")
	f.PrintContext("Probe ID", "abc123")
	f.PrintSuggestion("Run 'stackeye probe list' to see available probes")
	f.PrintRequestID("req_xyz789")

	output := buf.String()
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 lines of output, got %d: %q", len(lines), output)
	}

	// First line: Error message
	if !strings.HasPrefix(lines[0], "Error:") {
		t.Errorf("First line should start with 'Error:', got: %q", lines[0])
	}

	// Remaining lines: indented context
	for i := 1; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], "  ") {
			t.Errorf("Line %d should be indented, got: %q", i+1, lines[i])
		}
	}
}
