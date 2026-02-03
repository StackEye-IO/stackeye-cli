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

func TestIsPiped_OverrideTrue(t *testing.T) {
	orig := loadIsPipedOverride()
	defer func() { storeIsPipedOverride(orig) }()

	storeIsPipedOverride(func() bool { return true })

	if !IsPiped() {
		t.Error("IsPiped() should return true when override returns true")
	}
}

func TestIsPiped_OverrideFalse(t *testing.T) {
	orig := loadIsPipedOverride()
	defer func() { storeIsPipedOverride(orig) }()

	storeIsPipedOverride(func() bool { return false })

	if IsPiped() {
		t.Error("IsPiped() should return false when override returns false")
	}
}

func TestIsPiped_OverrideNilFallsThrough(t *testing.T) {
	orig := loadIsPipedOverride()
	defer func() { storeIsPipedOverride(orig) }()

	storeIsPipedOverride(nil)

	// With nil override, IsPiped should fall through to OS detection.
	// Just verify it runs without panic; actual value depends on environment.
	_ = IsPiped()
}

func TestSetIsPipedOverride(t *testing.T) {
	orig := loadIsPipedOverride()
	defer func() { storeIsPipedOverride(orig) }()

	called := false
	SetIsPipedOverride(func() bool {
		called = true
		return false
	})

	IsPiped()
	if !called {
		t.Error("isPipedOverride was not called after SetIsPipedOverride")
	}

	SetIsPipedOverride(nil)
	if loadIsPipedOverride() != nil {
		t.Error("isPipedOverride should be nil after SetIsPipedOverride(nil)")
	}
}

func TestIsInteractive_WithPipedOverrideFalse(t *testing.T) {
	origPiped := loadIsPipedOverride()
	origNoInput := loadNoInputGetter()
	defer func() {
		storeIsPipedOverride(origPiped)
		storeNoInputGetter(origNoInput)
	}()

	// Override IsPiped to return false (simulate real TTY)
	storeIsPipedOverride(func() bool { return false })
	storeNoInputGetter(nil)
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("STACKEYE_NO_INPUT", "")

	if !IsInteractive() {
		t.Error("IsInteractive() should return true when IsPiped override returns false and no other disabling conditions")
	}
}

func TestIsInteractive_WithPipedOverrideTrue(t *testing.T) {
	origPiped := loadIsPipedOverride()
	defer func() { storeIsPipedOverride(origPiped) }()

	// Override IsPiped to return true (simulate piped output)
	storeIsPipedOverride(func() bool { return true })

	if IsInteractive() {
		t.Error("IsInteractive() should return false when IsPiped override returns true")
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
	origGetter := loadNoInputGetter()
	defer func() { storeNoInputGetter(origGetter) }()

	storeNoInputGetter(func() bool { return true })

	// Even if stdout is a TTY, --no-input should make it non-interactive
	// Note: in test environments stdout is usually piped anyway, so
	// IsInteractive will return false for that reason too. We test the
	// flag path specifically via isAnimationEnabled tests.
	if IsInteractive() {
		t.Error("IsInteractive() should return false when --no-input is set")
	}
}

func TestIsInteractive_StackeyeNoInputEnv(t *testing.T) {
	origGetter := loadNoInputGetter()
	defer func() { storeNoInputGetter(origGetter) }()
	storeNoInputGetter(nil)

	t.Setenv("STACKEYE_NO_INPUT", "1")

	if IsInteractive() {
		t.Error("IsInteractive() should return false when STACKEYE_NO_INPUT=1")
	}
}

func TestIsInteractive_StackeyeNoInputEnvZero(t *testing.T) {
	origGetter := loadNoInputGetter()
	defer func() { storeNoInputGetter(origGetter) }()
	storeNoInputGetter(nil)

	// STACKEYE_NO_INPUT=0 should NOT disable interactive mode
	t.Setenv("STACKEYE_NO_INPUT", "0")

	// Note: will still be non-interactive if stdout is piped (test env),
	// but this verifies the env var value "0" doesn't trigger the disable path
	_ = IsInteractive() // Just verify no panic; actual value depends on TTY
}

func TestIsInteractive_StackeyeNoInputEnvEmpty(t *testing.T) {
	origGetter := loadNoInputGetter()
	defer func() { storeNoInputGetter(origGetter) }()
	storeNoInputGetter(nil)

	// Empty STACKEYE_NO_INPUT should NOT disable interactive mode
	t.Setenv("STACKEYE_NO_INPUT", "")

	_ = IsInteractive() // Verify no panic
}

func TestIsAnimationEnabled_DumbTerminalDisables(t *testing.T) {
	origGetter := loadNoInputGetter()
	origConfig := loadConfigGetter()
	defer func() {
		storeNoInputGetter(origGetter)
		storeConfigGetter(origConfig)
	}()

	storeNoInputGetter(nil)
	storeConfigGetter(nil)
	t.Setenv("TERM", "dumb")

	if isAnimationEnabled() {
		t.Error("animations should be disabled when TERM=dumb")
	}
}

func TestIsPiped_ConsistentWithIsInteractive(t *testing.T) {
	// When stdout is piped, IsInteractive must return false
	if IsPiped() && IsInteractive() {
		t.Error("IsInteractive() must return false when IsPiped() is true")
	}
}
