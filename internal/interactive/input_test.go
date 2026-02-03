package interactive

import (
	"errors"
	"testing"
)

// --- Input tests ---

func TestInput_NoInputBypass_ReturnsDefault(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return true }

	result, err := Input("Enter name:", WithInputDefault("default-value"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "default-value" {
		t.Errorf("expected 'default-value', got %q", result)
	}
}

func TestInput_NoInputBypass_ReturnsEmptyWithoutDefault(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return true }

	result, err := Input("Enter name:")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestInput_NoInputFalse_DoesNotBypass(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return false }

	// Without a TTY, AskString will fail. We verify it doesn't take the bypass path.
	_, err := Input("Enter name:")
	if err == nil {
		t.Error("expected error when noInput=false and no TTY available")
	}
}

func TestInput_NilNoInputGetter(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = nil

	// Without a TTY and no bypass, the SDK call will fail
	_, err := Input("Enter name:")
	if err == nil {
		t.Error("expected error when noInputGetter is nil and no TTY")
	}
}

func TestInputOptions_Defaults(t *testing.T) {
	o := &inputOptions{}
	if o.defaultVal != "" {
		t.Error("default defaultVal should be empty")
	}
	if o.help != "" {
		t.Error("default help should be empty")
	}
	if o.validate != nil {
		t.Error("default validate should be nil")
	}
}

func TestWithInputDefault_SetsDefault(t *testing.T) {
	o := &inputOptions{}
	WithInputDefault("my-default")(o)
	if o.defaultVal != "my-default" {
		t.Errorf("expected 'my-default', got %q", o.defaultVal)
	}
}

func TestWithInputHelp_SetsHelp(t *testing.T) {
	o := &inputOptions{}
	WithInputHelp("Enter your full name")(o)
	if o.help != "Enter your full name" {
		t.Errorf("expected help text, got %q", o.help)
	}
}

func TestWithInputValidate_SetsValidate(t *testing.T) {
	o := &inputOptions{}
	fn := func(s string) error {
		if s == "" {
			return errors.New("cannot be empty")
		}
		return nil
	}
	WithInputValidate(fn)(o)
	if o.validate == nil {
		t.Fatal("validate should be set")
	}
	if err := o.validate(""); err == nil {
		t.Error("expected error for empty input")
	}
	if err := o.validate("valid"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Password tests ---

func TestPassword_NoInputBypass_ReturnsEmpty(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return true }

	result, err := Password("Enter password:")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string for password bypass, got %q", result)
	}
}

func TestPassword_NoInputFalse_DoesNotBypass(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return false }

	// Without a TTY, AskPassword will fail
	_, err := Password("Enter password:")
	if err == nil {
		t.Error("expected error when noInput=false and no TTY available")
	}
}

func TestPassword_NilNoInputGetter(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = nil

	// Without a TTY and no bypass, the SDK call will fail
	_, err := Password("Enter password:")
	if err == nil {
		t.Error("expected error when noInputGetter is nil and no TTY")
	}
}

func TestPasswordOptions_Defaults(t *testing.T) {
	o := &passwordOptions{}
	if o.help != "" {
		t.Error("default help should be empty")
	}
	if o.validate != nil {
		t.Error("default validate should be nil")
	}
}

func TestWithPasswordHelp_SetsHelp(t *testing.T) {
	o := &passwordOptions{}
	WithPasswordHelp("Enter your API key")(o)
	if o.help != "Enter your API key" {
		t.Errorf("expected help text, got %q", o.help)
	}
}

func TestWithPasswordValidate_SetsValidate(t *testing.T) {
	o := &passwordOptions{}
	fn := func(s string) error {
		if len(s) < 8 {
			return errors.New("password must be at least 8 characters")
		}
		return nil
	}
	WithPasswordValidate(fn)(o)
	if o.validate == nil {
		t.Fatal("validate should be set")
	}
	if err := o.validate("short"); err == nil {
		t.Error("expected error for short password")
	}
	if err := o.validate("long-enough-password"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
