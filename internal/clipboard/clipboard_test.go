package clipboard

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestCopy_EmptyText(t *testing.T) {
	err := Copy("")
	if !errors.Is(err, ErrEmptyText) {
		t.Errorf("Copy(\"\") = %v, want ErrEmptyText", err)
	}
}

func TestCopy_WhitespaceOnly(t *testing.T) {
	err := Copy("   ")
	if !errors.Is(err, ErrEmptyText) {
		t.Errorf("Copy(\"   \") = %v, want ErrEmptyText", err)
	}
}

func TestCopy_Success(t *testing.T) {
	var calledWith string
	SetCopier(func(text string) error {
		calledWith = text
		return nil
	})
	defer SetCopier(nil)

	const testText = "se_abc123def456"
	if err := Copy(testText); err != nil {
		t.Errorf("Copy(%q) = %v, want nil", testText, err)
	}
	if calledWith != testText {
		t.Errorf("copier called with %q, want %q", calledWith, testText)
	}
}

func TestCopy_CopierError(t *testing.T) {
	copierErr := errors.New("clipboard unavailable")
	SetCopier(func(string) error { return copierErr })
	defer SetCopier(nil)

	err := Copy("some text")
	if !errors.Is(err, copierErr) {
		t.Errorf("Copy() = %v, want %v", err, copierErr)
	}
}

func TestCopyOrPrint_EmptyText(t *testing.T) {
	err := CopyOrPrint("")
	if !errors.Is(err, ErrEmptyText) {
		t.Errorf("CopyOrPrint(\"\") = %v, want ErrEmptyText", err)
	}
}

func TestCopyOrPrint_Success(t *testing.T) {
	var called bool
	SetCopier(func(string) error {
		called = true
		return nil
	})
	defer SetCopier(nil)

	if err := CopyOrPrint("test value"); err != nil {
		t.Errorf("CopyOrPrint() = %v, want nil", err)
	}
	if !called {
		t.Error("copier was not called")
	}
}

func TestCopyOrPrint_FallbackPrintsText(t *testing.T) {
	SetCopier(func(string) error { return errors.New("no clipboard") })
	defer SetCopier(nil)

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	const testText = "se_secret_api_key_value"
	err := CopyOrPrint(testText)

	w.Close()
	os.Stderr = oldStderr

	if err != nil {
		t.Errorf("CopyOrPrint() = %v, want nil (fallback should absorb error)", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read stderr: %v", err)
	}
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte(testText)) {
		t.Errorf("fallback output does not contain text %q:\n%s", testText, output)
	}
	if !bytes.Contains([]byte(output), []byte("Could not copy to clipboard")) {
		t.Errorf("fallback output missing expected message:\n%s", output)
	}
}

func TestSetCopier_NilRestoresDefault(t *testing.T) {
	// Override copier
	SetCopier(func(string) error { return errors.New("mock") })

	// Restore default
	SetCopier(nil)

	// copier should now be defaultCopy (we can't call it in tests since
	// it would actually try to access the clipboard, but we verify it's not nil)
	if copier == nil {
		t.Error("SetCopier(nil) set copier to nil, should restore default")
	}
}

func TestCopy_PreservesContent(t *testing.T) {
	SetCopier(func(string) error { return nil })
	defer SetCopier(nil)

	tests := []struct {
		name string
		text string
	}{
		{"simple text", "hello"},
		{"API key format", "se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		{"URL", "https://app.stackeye.io/settings/api-keys"},
		{"multiline", "line1\nline2\nline3"},
		{"special chars", "text with spaces & symbols! @#$%"},
		{"unicode", "clipboard: \u2705 copied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured string
			SetCopier(func(text string) error {
				captured = text
				return nil
			})

			if err := Copy(tt.text); err != nil {
				t.Errorf("Copy(%q) = %v, want nil", tt.text, err)
			}
			if captured != tt.text {
				t.Errorf("copier received %q, want %q", captured, tt.text)
			}
		})
	}
}

func TestValidateText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"non-empty", "hello", false},
		{"empty", "", true},
		{"spaces only", "   ", true},
		{"tabs only", "\t\t", true},
		{"newlines only", "\n\n", true},
		{"text with leading spaces", "  hello", false},
		{"single char", "x", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateText(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateText(%q) error = %v, wantErr %v", tt.text, err, tt.wantErr)
			}
		})
	}
}
