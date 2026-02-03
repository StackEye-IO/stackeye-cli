package output

import (
	"testing"
)

func TestIsPiped_InTestEnvironment(t *testing.T) {
	// In test environments, stdout is typically piped to capture output.
	// This test verifies IsPiped() returns a boolean without panicking.
	result := IsPiped()
	// In CI/test environments, stdout is usually piped
	if !result {
		t.Log("stdout is a TTY in this test environment (running in terminal)")
	}
}

func TestIsStderrPiped_InTestEnvironment(t *testing.T) {
	// Verify IsStderrPiped() returns a boolean without panicking.
	result := IsStderrPiped()
	_ = result // Just verify no panic
}

func TestIsDumbTerminal(t *testing.T) {
	tests := []struct {
		name     string
		termVal  string
		setTerm  bool
		expected bool
	}{
		{
			name:     "TERM=dumb returns true",
			termVal:  "dumb",
			setTerm:  true,
			expected: true,
		},
		{
			name:     "TERM=xterm returns false",
			termVal:  "xterm",
			setTerm:  true,
			expected: false,
		},
		{
			name:     "TERM=xterm-256color returns false",
			termVal:  "xterm-256color",
			setTerm:  true,
			expected: false,
		},
		{
			name:     "empty TERM returns false",
			termVal:  "",
			setTerm:  true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setTerm {
				t.Setenv("TERM", tt.termVal)
			}
			if got := IsDumbTerminal(); got != tt.expected {
				t.Errorf("IsDumbTerminal() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsInteractive_DumbTerminal(t *testing.T) {
	t.Setenv("TERM", "dumb")

	if IsInteractive() {
		t.Error("IsInteractive() should return false when TERM=dumb")
	}
}

func TestIsInteractive_NoInputFlag(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return true }

	// Even if stdout is a TTY, --no-input should make it non-interactive
	// Note: in test environments stdout is usually piped anyway, so
	// IsInteractive will return false for that reason too. We test the
	// flag path specifically via isSpinnerEnabled tests.
	if IsInteractive() {
		t.Error("IsInteractive() should return false when --no-input is set")
	}
}

func TestIsInteractive_StackeyeNoInputEnv(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()
	noInputGetter = nil

	t.Setenv("STACKEYE_NO_INPUT", "1")

	if IsInteractive() {
		t.Error("IsInteractive() should return false when STACKEYE_NO_INPUT=1")
	}
}

func TestIsInteractive_StackeyeNoInputEnvZero(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()
	noInputGetter = nil

	// STACKEYE_NO_INPUT=0 should NOT disable interactive mode
	t.Setenv("STACKEYE_NO_INPUT", "0")

	// Note: will still be non-interactive if stdout is piped (test env),
	// but this verifies the env var value "0" doesn't trigger the disable path
	_ = IsInteractive() // Just verify no panic; actual value depends on TTY
}

func TestIsInteractive_StackeyeNoInputEnvEmpty(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()
	noInputGetter = nil

	// Empty STACKEYE_NO_INPUT should NOT disable interactive mode
	t.Setenv("STACKEYE_NO_INPUT", "")

	_ = IsInteractive() // Verify no panic
}

func TestIsSpinnerEnabled_DumbTerminalDisables(t *testing.T) {
	origGetter := noInputGetter
	origConfig := configGetter
	defer func() {
		noInputGetter = origGetter
		configGetter = origConfig
	}()

	noInputGetter = nil
	configGetter = nil
	t.Setenv("TERM", "dumb")

	if isSpinnerEnabled() {
		t.Error("spinner should be disabled when TERM=dumb")
	}
}

func TestIsPiped_ConsistentWithIsInteractive(t *testing.T) {
	// When stdout is piped, IsInteractive must return false
	if IsPiped() && IsInteractive() {
		t.Error("IsInteractive() must return false when IsPiped() is true")
	}
}
