package interactive

import (
	"errors"
	"testing"
)

// --- Select tests ---

func TestSelect_NoInputBypass_ReturnsDefault(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return true }

	selected, err := Select("Pick one:", []string{"a", "b", "c"}, WithSelectDefault("b"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != "b" {
		t.Errorf("expected 'b', got %q", selected)
	}
}

func TestSelect_NoInputBypass_ReturnsFirstOption(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return true }

	selected, err := Select("Pick one:", []string{"x", "y", "z"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != "x" {
		t.Errorf("expected 'x', got %q", selected)
	}
}

func TestSelect_NoInputFalse_DoesNotBypass(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return false }

	// Without a TTY, AskSelect will fail. We verify it doesn't take the bypass path.
	_, err := Select("Pick one:", []string{"a", "b"})
	if err == nil {
		t.Error("expected error when noInput=false and no TTY available")
	}
	// The error should NOT be ErrNoOptions (bypass was not taken)
	if errors.Is(err, ErrNoOptions) {
		t.Error("should not get ErrNoOptions with valid options")
	}
}

func TestSelect_NilNoInputGetter(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = nil

	// Without a TTY and no bypass, the SDK call will fail
	_, err := Select("Pick one:", []string{"a", "b"})
	if err == nil {
		t.Error("expected error when noInputGetter is nil and no TTY")
	}
}

func TestSelect_EmptyOptions(t *testing.T) {
	_, err := Select("Pick one:", []string{})
	if !errors.Is(err, ErrNoOptions) {
		t.Errorf("expected ErrNoOptions, got %v", err)
	}
}

func TestSelect_NilOptions(t *testing.T) {
	_, err := Select("Pick one:", nil)
	if !errors.Is(err, ErrNoOptions) {
		t.Errorf("expected ErrNoOptions, got %v", err)
	}
}

func TestSelectOptions_Defaults(t *testing.T) {
	o := &selectOptions{}
	if o.defaultVal != "" {
		t.Error("default defaultVal should be empty")
	}
	if o.help != "" {
		t.Error("default help should be empty")
	}
	if o.pageSize != 0 {
		t.Error("default pageSize should be 0")
	}
}

func TestWithSelectDefault_SetsDefault(t *testing.T) {
	o := &selectOptions{}
	WithSelectDefault("option-b")(o)
	if o.defaultVal != "option-b" {
		t.Errorf("expected 'option-b', got %q", o.defaultVal)
	}
}

func TestWithSelectHelp_SetsHelp(t *testing.T) {
	o := &selectOptions{}
	WithSelectHelp("Choose the deployment target")(o)
	if o.help != "Choose the deployment target" {
		t.Errorf("expected help text, got %q", o.help)
	}
}

func TestWithSelectPageSize_SetsPageSize(t *testing.T) {
	o := &selectOptions{}
	WithSelectPageSize(15)(o)
	if o.pageSize != 15 {
		t.Errorf("expected 15, got %d", o.pageSize)
	}
}

// --- MultiSelect tests ---

func TestMultiSelect_NoInputBypass_ReturnsDefaults(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return true }

	defaults := []string{"b", "c"}
	selected, err := MultiSelect("Pick some:", []string{"a", "b", "c"}, WithMultiSelectDefaults(defaults))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 2 || selected[0] != "b" || selected[1] != "c" {
		t.Errorf("expected [b c], got %v", selected)
	}
}

func TestMultiSelect_NoInputBypass_ReturnsAllOptions(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return true }

	selected, err := MultiSelect("Pick some:", []string{"x", "y", "z"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(selected) != 3 || selected[0] != "x" || selected[1] != "y" || selected[2] != "z" {
		t.Errorf("expected [x y z], got %v", selected)
	}
}

func TestMultiSelect_NoInputFalse_DoesNotBypass(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = func() bool { return false }

	_, err := MultiSelect("Pick some:", []string{"a", "b"})
	if err == nil {
		t.Error("expected error when noInput=false and no TTY available")
	}
	if errors.Is(err, ErrNoOptions) {
		t.Error("should not get ErrNoOptions with valid options")
	}
}

func TestMultiSelect_NilNoInputGetter(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	noInputGetter = nil

	_, err := MultiSelect("Pick some:", []string{"a", "b"})
	if err == nil {
		t.Error("expected error when noInputGetter is nil and no TTY")
	}
}

func TestMultiSelect_EmptyOptions(t *testing.T) {
	_, err := MultiSelect("Pick some:", []string{})
	if !errors.Is(err, ErrNoOptions) {
		t.Errorf("expected ErrNoOptions, got %v", err)
	}
}

func TestMultiSelect_NilOptions(t *testing.T) {
	_, err := MultiSelect("Pick some:", nil)
	if !errors.Is(err, ErrNoOptions) {
		t.Errorf("expected ErrNoOptions, got %v", err)
	}
}

func TestMultiSelectOptions_Defaults(t *testing.T) {
	o := &multiSelectOptions{}
	if o.defaults != nil {
		t.Error("default defaults should be nil")
	}
	if o.help != "" {
		t.Error("default help should be empty")
	}
	if o.pageSize != 0 {
		t.Error("default pageSize should be 0")
	}
	if o.validate != nil {
		t.Error("default validate should be nil")
	}
}

func TestWithMultiSelectDefaults_SetsDefaults(t *testing.T) {
	o := &multiSelectOptions{}
	WithMultiSelectDefaults([]string{"a", "c"})(o)
	if len(o.defaults) != 2 || o.defaults[0] != "a" || o.defaults[1] != "c" {
		t.Errorf("expected [a c], got %v", o.defaults)
	}
}

func TestWithMultiSelectHelp_SetsHelp(t *testing.T) {
	o := &multiSelectOptions{}
	WithMultiSelectHelp("Select monitoring regions")(o)
	if o.help != "Select monitoring regions" {
		t.Errorf("expected help text, got %q", o.help)
	}
}

func TestWithMultiSelectPageSize_SetsPageSize(t *testing.T) {
	o := &multiSelectOptions{}
	WithMultiSelectPageSize(20)(o)
	if o.pageSize != 20 {
		t.Errorf("expected 20, got %d", o.pageSize)
	}
}

func TestWithMultiSelectValidate_SetsValidate(t *testing.T) {
	o := &multiSelectOptions{}
	fn := func(selections []string) error {
		if len(selections) == 0 {
			return errors.New("select at least one")
		}
		return nil
	}
	WithMultiSelectValidate(fn)(o)
	if o.validate == nil {
		t.Fatal("validate should be set")
	}
	if err := o.validate([]string{}); err == nil {
		t.Error("expected error for empty selections")
	}
	if err := o.validate([]string{"a"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestErrNoOptions_IsDistinct(t *testing.T) {
	if ErrNoOptions == nil {
		t.Fatal("ErrNoOptions should not be nil")
	}
	if ErrNoOptions.Error() != "no options provided" {
		t.Errorf("unexpected error message: %s", ErrNoOptions.Error())
	}
	if errors.Is(ErrNoOptions, ErrCancelled) {
		t.Error("ErrNoOptions should not match ErrCancelled")
	}
}
