package debug

import (
	"bytes"
	"strings"
	"testing"
)

// resetState resets package-level state for test isolation.
func resetState() {
	mu.Lock()
	defer mu.Unlock()
	currentVerbosityGetter = nil
	writer = &bytes.Buffer{}
	rebuildLogger()
}

func TestLog_SuppressedBelowVerbosity(t *testing.T) {
	resetState()
	buf := &bytes.Buffer{}
	SetWriter(buf)
	SetVerbosityGetter(func() int { return 2 })

	Log(3, "should not appear")

	if buf.Len() != 0 {
		t.Errorf("expected no output at verbosity 2 for level 3 message, got: %s", buf.String())
	}
}

func TestLog_EmittedAtVerbosity(t *testing.T) {
	resetState()
	buf := &bytes.Buffer{}
	SetWriter(buf)
	SetVerbosityGetter(func() int { return 4 })

	Log(3, "config loaded", "path", "/etc/stackeye")

	output := buf.String()
	if !strings.Contains(output, "config loaded") {
		t.Errorf("expected 'config loaded' in output, got: %s", output)
	}
	if !strings.Contains(output, "path=/etc/stackeye") {
		t.Errorf("expected 'path=/etc/stackeye' in output, got: %s", output)
	}
	if !strings.Contains(output, "v=3") {
		t.Errorf("expected 'v=3' in output, got: %s", output)
	}
}

func TestLogf_FormatsMessage(t *testing.T) {
	resetState()
	buf := &bytes.Buffer{}
	SetWriter(buf)
	SetVerbosityGetter(func() int { return 5 })

	Logf(4, "resolved context: %s", "production")

	output := buf.String()
	if !strings.Contains(output, "resolved context: production") {
		t.Errorf("expected formatted message, got: %s", output)
	}
}

func TestLogf_SuppressedBelowVerbosity(t *testing.T) {
	resetState()
	buf := &bytes.Buffer{}
	SetWriter(buf)
	SetVerbosityGetter(func() int { return 0 })

	Logf(4, "should not appear: %s", "test")

	if buf.Len() != 0 {
		t.Errorf("expected no output, got: %s", buf.String())
	}
}

func TestEnabled(t *testing.T) {
	resetState()
	SetVerbosityGetter(func() int { return 5 })

	tests := []struct {
		level    int
		expected bool
	}{
		{0, true},
		{3, true},
		{5, true},
		{6, false},
		{10, false},
	}

	for _, tt := range tests {
		got := Enabled(tt.level)
		if got != tt.expected {
			t.Errorf("Enabled(%d) = %v, want %v (verbosity=5)", tt.level, got, tt.expected)
		}
	}
}

func TestEnabled_NoGetter(t *testing.T) {
	resetState()
	// No verbosity getter set

	if Enabled(1) {
		t.Error("expected Enabled(1) = false with no getter")
	}
	if !Enabled(0) {
		t.Error("expected Enabled(0) = true with no getter (verbosity defaults to 0)")
	}
}

func TestConfigLoaded(t *testing.T) {
	resetState()
	buf := &bytes.Buffer{}
	SetWriter(buf)
	SetVerbosityGetter(func() int { return 3 })

	ConfigLoaded("/home/user/.config/stackeye/config.yaml", "default")

	output := buf.String()
	if !strings.Contains(output, "config loaded") {
		t.Errorf("expected 'config loaded', got: %s", output)
	}
	if !strings.Contains(output, "context=default") {
		t.Errorf("expected 'context=default', got: %s", output)
	}
}

func TestConfigLoaded_Suppressed(t *testing.T) {
	resetState()
	buf := &bytes.Buffer{}
	SetWriter(buf)
	SetVerbosityGetter(func() int { return 2 })

	ConfigLoaded("/path", "ctx")

	if buf.Len() != 0 {
		t.Errorf("expected no output at verbosity 2, got: %s", buf.String())
	}
}

func TestConfigError(t *testing.T) {
	resetState()
	buf := &bytes.Buffer{}
	SetWriter(buf)
	SetVerbosityGetter(func() int { return 1 })

	ConfigError("/bad/path", &testError{"file not found"})

	output := buf.String()
	if !strings.Contains(output, "config error") {
		t.Errorf("expected 'config error', got: %s", output)
	}
	if !strings.Contains(output, "file not found") {
		t.Errorf("expected error message, got: %s", output)
	}
}

func TestEnvOverride(t *testing.T) {
	resetState()
	buf := &bytes.Buffer{}
	SetWriter(buf)
	SetVerbosityGetter(func() int { return 3 })

	EnvOverride("STACKEYE_API_URL", "https://custom.api.example.com")

	output := buf.String()
	if !strings.Contains(output, "env override") {
		t.Errorf("expected 'env override', got: %s", output)
	}
	if !strings.Contains(output, "STACKEYE_API_URL") {
		t.Errorf("expected env var name, got: %s", output)
	}
}

func TestEnvOverride_RedactsSensitiveValues(t *testing.T) {
	resetState()
	buf := &bytes.Buffer{}
	SetWriter(buf)
	SetVerbosityGetter(func() int { return 3 })

	EnvOverride("STACKEYE_API_KEY", "se_abcdef1234567890")

	output := buf.String()
	if strings.Contains(output, "abcdef1234567890") {
		t.Errorf("expected redacted value, but found full key in output: %s", output)
	}
	if !strings.Contains(output, "se_a****") {
		t.Errorf("expected redacted value 'se_a****', got: %s", output)
	}
}

func TestEnvOverride_DoesNotRedactNonSensitive(t *testing.T) {
	resetState()
	buf := &bytes.Buffer{}
	SetWriter(buf)
	SetVerbosityGetter(func() int { return 3 })

	EnvOverride("STACKEYE_API_URL", "https://api.stackeye.io")

	output := buf.String()
	if !strings.Contains(output, "https://api.stackeye.io") {
		t.Errorf("expected full URL in output, got: %s", output)
	}
}

// testError is a simple error type for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
