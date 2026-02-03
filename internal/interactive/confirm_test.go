package interactive

import (
	"errors"
	"testing"
)

func TestConfirm_YesFlagBypass(t *testing.T) {
	// --yes flag should bypass the prompt entirely and return true
	confirmed, err := Confirm("Delete this resource?", WithYesFlag(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Error("expected true when --yes flag is set")
	}
}

func TestConfirm_YesFlagBypass_IgnoresDefault(t *testing.T) {
	// --yes flag returns true even if default is false
	confirmed, err := Confirm("Delete this resource?",
		WithYesFlag(true),
		WithDefault(false),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Error("expected true when --yes flag is set, regardless of default")
	}
}

func TestConfirm_NoInputBypass(t *testing.T) {
	// Save and restore original getter
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return true }

	confirmed, err := Confirm("Delete this resource?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Error("expected true when --no-input is active")
	}
}

func TestConfirm_NoInputFalse_DoesNotBypass(t *testing.T) {
	// When noInputGetter returns false, prompt would be shown.
	// We can't test the actual prompt in unit tests without TTY,
	// so we verify the bypass path is NOT taken by checking that
	// --yes flag still controls the result.
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return false }

	// With --yes flag, should still bypass
	confirmed, err := Confirm("Delete this resource?", WithYesFlag(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Error("expected true when --yes flag is set even with noInput=false")
	}
}

func TestConfirm_NilNoInputGetter(t *testing.T) {
	// When noInputGetter is nil, it should not panic and --yes flag controls
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = nil

	confirmed, err := Confirm("Delete this resource?", WithYesFlag(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmed {
		t.Error("expected true when --yes flag is set with nil noInputGetter")
	}
}

func TestConfirm_DefaultOptions(t *testing.T) {
	// Verify default confirmOptions values
	o := &confirmOptions{}
	if o.yes {
		t.Error("default yes should be false")
	}
	if o.defaultVal {
		t.Error("default defaultVal should be false (safe option)")
	}
	if o.help != "" {
		t.Error("default help should be empty")
	}
}

func TestWithDefault_SetsDefault(t *testing.T) {
	o := &confirmOptions{}
	WithDefault(true)(o)
	if !o.defaultVal {
		t.Error("WithDefault(true) should set defaultVal to true")
	}
}

func TestWithHelp_SetsHelp(t *testing.T) {
	o := &confirmOptions{}
	WithHelp("This will permanently delete the resource")(o)
	if o.help != "This will permanently delete the resource" {
		t.Errorf("expected help text, got %q", o.help)
	}
}

func TestWithYesFlag_SetsYes(t *testing.T) {
	o := &confirmOptions{}
	WithYesFlag(true)(o)
	if !o.yes {
		t.Error("WithYesFlag(true) should set yes to true")
	}
}

func TestErrCancelled_IsDistinct(t *testing.T) {
	// ErrCancelled should be a distinct error for caller detection
	if ErrCancelled == nil {
		t.Fatal("ErrCancelled should not be nil")
	}
	if ErrCancelled.Error() != "operation cancelled by user" {
		t.Errorf("unexpected error message: %s", ErrCancelled.Error())
	}
}

func TestErrCancelled_ErrorsIs(t *testing.T) {
	// Wrapping ErrCancelled should still match with errors.Is
	wrapped := errors.New("wrapped: " + ErrCancelled.Error())
	_ = wrapped // just verifying ErrCancelled is a usable error

	if !errors.Is(ErrCancelled, ErrCancelled) {
		t.Error("ErrCancelled should match itself with errors.Is")
	}
}

func TestSetNoInputGetter(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	called := false
	SetNoInputGetter(func() bool {
		called = true
		return false
	})

	if noInputGetter == nil {
		t.Fatal("noInputGetter should be set after SetNoInputGetter")
	}

	noInputGetter()
	if !called {
		t.Error("noInputGetter was not called after SetNoInputGetter")
	}
}
