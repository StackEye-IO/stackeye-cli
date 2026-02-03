package browser

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestOpen_EmptyURL(t *testing.T) {
	err := Open("")
	if !errors.Is(err, ErrEmptyURL) {
		t.Errorf("Open(\"\") = %v, want ErrEmptyURL", err)
	}
}

func TestOpen_InvalidURL(t *testing.T) {
	err := Open("not a url")
	if err == nil {
		t.Error("Open(\"not a url\") = nil, want error")
	}
}

func TestOpen_Success(t *testing.T) {
	var calledWith string
	SetOpener(func(url string) error {
		calledWith = url
		return nil
	})
	defer SetOpener(nil)

	const testURL = "https://app.stackeye.io/dashboard"
	if err := Open(testURL); err != nil {
		t.Errorf("Open(%q) = %v, want nil", testURL, err)
	}
	if calledWith != testURL {
		t.Errorf("opener called with %q, want %q", calledWith, testURL)
	}
}

func TestOpen_OpenerError(t *testing.T) {
	openerErr := errors.New("browser failed")
	SetOpener(func(string) error { return openerErr })
	defer SetOpener(nil)

	err := Open("https://example.com")
	if !errors.Is(err, openerErr) {
		t.Errorf("Open() = %v, want %v", err, openerErr)
	}
}

func TestOpenOrPrint_EmptyURL(t *testing.T) {
	err := OpenOrPrint("")
	if !errors.Is(err, ErrEmptyURL) {
		t.Errorf("OpenOrPrint(\"\") = %v, want ErrEmptyURL", err)
	}
}

func TestOpenOrPrint_InvalidURL(t *testing.T) {
	err := OpenOrPrint("not a url")
	if err == nil {
		t.Error("OpenOrPrint(\"not a url\") = nil, want error")
	}
}

func TestOpenOrPrint_Success(t *testing.T) {
	var called bool
	SetOpener(func(string) error {
		called = true
		return nil
	})
	defer SetOpener(nil)

	if err := OpenOrPrint("https://example.com"); err != nil {
		t.Errorf("OpenOrPrint() = %v, want nil", err)
	}
	if !called {
		t.Error("opener was not called")
	}
}

func TestOpenOrPrint_FallbackPrintsURL(t *testing.T) {
	SetOpener(func(string) error { return errors.New("no browser") })
	defer SetOpener(nil)

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	const testURL = "https://app.stackeye.io"
	err := OpenOrPrint(testURL)

	w.Close()
	os.Stderr = oldStderr

	if err != nil {
		t.Errorf("OpenOrPrint() = %v, want nil (fallback should absorb error)", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read stderr: %v", err)
	}
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte(testURL)) {
		t.Errorf("fallback output does not contain URL %q:\n%s", testURL, output)
	}
	if !bytes.Contains([]byte(output), []byte("Could not open browser")) {
		t.Errorf("fallback output missing expected message:\n%s", output)
	}
}

func TestSetOpener_NilRestoresDefault(t *testing.T) {
	// Override opener
	SetOpener(func(string) error { return errors.New("mock") })

	// Restore default
	SetOpener(nil)

	// opener should now be defaultOpen (we can't call it in tests since
	// it would actually try to open a browser, but we verify it's not nil)
	if opener == nil {
		t.Error("SetOpener(nil) set opener to nil, should restore default")
	}
}

func TestOpen_ValidURLSchemes(t *testing.T) {
	SetOpener(func(string) error { return nil })
	defer SetOpener(nil)

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https URL", "https://example.com", false},
		{"http URL", "http://localhost:3000", false},
		{"https with path", "https://app.stackeye.io/dashboard?tab=probes", false},
		{"empty string", "", true},
		{"bare word", "notaurl", true},
		{"spaces only", "   ", true},
		{"javascript scheme", "javascript:alert('xss')", true},
		{"file scheme", "file:///etc/passwd", true},
		{"data scheme", "data:text/html,hello", true},
		{"ftp scheme", "ftp://files.example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Open(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Open(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}
