package update

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sdkupdate "github.com/StackEye-IO/stackeye-go-sdk/update"
)

func TestShouldCheck(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		disabled       bool
		want           bool
	}{
		{
			name:           "normal version enabled",
			currentVersion: "1.0.0",
			disabled:       false,
			want:           true,
		},
		{
			name:           "normal version disabled",
			currentVersion: "1.0.0",
			disabled:       true,
			want:           false,
		},
		{
			name:           "dev version",
			currentVersion: "dev",
			disabled:       false,
			want:           false,
		},
		{
			name:           "empty version",
			currentVersion: "",
			disabled:       false,
			want:           false,
		},
		{
			name:           "prerelease version enabled",
			currentVersion: "1.0.0-beta.1",
			disabled:       false,
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldCheck(tt.currentVersion, tt.disabled)
			if got != tt.want {
				t.Errorf("ShouldCheck(%q, %v) = %v, want %v",
					tt.currentVersion, tt.disabled, got, tt.want)
			}
		})
	}
}

func TestNotifier_PrintNotification_NoUpdate(t *testing.T) {
	// Create a mock server that returns a release with same version
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := map[string]any{
			"tag_name":     "v1.0.0",
			"html_url":     "https://github.com/stackeye-io/stackeye-cli/releases/v1.0.0",
			"body":         "Release notes",
			"published_at": time.Now().Format(time.RFC3339),
			"assets":       []any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Create notifier with test server and temp cache dir
	tmpDir := t.TempDir()
	updater := sdkupdate.NewUpdater("stackeye-io/stackeye-cli", "1.0.0",
		sdkupdate.WithAPIURL(server.URL))
	checker := sdkupdate.NewChecker(updater, sdkupdate.WithCacheDir(tmpDir))

	var buf bytes.Buffer
	notifier := &Notifier{
		checker:        checker,
		currentVersion: "1.0.0",
		writer:         &buf,
		colorEnabled:   false,
	}

	ctx := context.Background()
	notifier.StartCheck(ctx)

	// Wait for check to complete
	_, _ = checker.WaitForBackgroundResult(2 * time.Second)

	printed := notifier.PrintNotification()
	if printed {
		t.Error("PrintNotification() = true, want false (no update available)")
	}

	if buf.Len() > 0 {
		t.Errorf("Notification printed when no update available: %s", buf.String())
	}
}

func TestNotifier_PrintNotification_UpdateAvailable(t *testing.T) {
	// Create a mock server that returns a newer version
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := map[string]any{
			"tag_name":     "v2.0.0",
			"html_url":     "https://github.com/stackeye-io/stackeye-cli/releases/v2.0.0",
			"body":         "Release notes for v2.0.0",
			"published_at": time.Now().Format(time.RFC3339),
			"assets":       []any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Create notifier with test server and temp cache dir to avoid cache pollution
	tmpDir := t.TempDir()
	updater := sdkupdate.NewUpdater("stackeye-io/stackeye-cli", "1.0.0",
		sdkupdate.WithAPIURL(server.URL))
	checker := sdkupdate.NewChecker(updater, sdkupdate.WithCacheDir(tmpDir))

	var buf bytes.Buffer
	notifier := &Notifier{
		checker:        checker,
		currentVersion: "1.0.0",
		writer:         &buf,
		colorEnabled:   false,
	}

	ctx := context.Background()
	notifier.StartCheck(ctx)

	// Wait for check to complete - use GetResult which has proper timeout handling
	result, err := checker.WaitForBackgroundResult(2 * time.Second)
	if err != nil {
		t.Fatalf("WaitForBackgroundResult failed: %v", err)
	}
	if result == nil {
		t.Fatal("WaitForBackgroundResult returned nil")
	}

	printed := notifier.PrintNotification()
	if !printed {
		t.Errorf("PrintNotification() = false, want true (update available, result.HasUpdate=%v)", result.HasUpdate)
	}

	output := buf.String()
	if !strings.Contains(output, "new version") {
		t.Errorf("Notification missing 'new version': %s", output)
	}
	if !strings.Contains(output, "1.0.0") {
		t.Errorf("Notification missing current version '1.0.0': %s", output)
	}
	if !strings.Contains(output, "2.0.0") {
		t.Errorf("Notification missing new version '2.0.0': %s", output)
	}
	if !strings.Contains(output, "stackeye upgrade") {
		t.Errorf("Notification missing upgrade command: %s", output)
	}
}

func TestNotifier_PrintNotification_PlainText(t *testing.T) {
	// Create a mock server that returns a newer version
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := map[string]any{
			"tag_name":     "v2.0.0",
			"html_url":     "https://github.com/stackeye-io/stackeye-cli/releases/v2.0.0",
			"body":         "Release notes",
			"published_at": time.Now().Format(time.RFC3339),
			"assets":       []any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	updater := sdkupdate.NewUpdater("stackeye-io/stackeye-cli", "1.0.0",
		sdkupdate.WithAPIURL(server.URL))
	checker := sdkupdate.NewChecker(updater, sdkupdate.WithCacheDir(tmpDir))

	var buf bytes.Buffer
	notifier := &Notifier{
		checker:        checker,
		currentVersion: "1.0.0",
		writer:         &buf,
		colorEnabled:   false,
	}

	ctx := context.Background()
	notifier.StartCheck(ctx)

	// Wait for check to complete
	_, _ = checker.WaitForBackgroundResult(2 * time.Second)
	notifier.PrintNotification()

	output := buf.String()

	// If update was available, verify no ANSI escape codes in plain text mode
	if output != "" && strings.Contains(output, "\033[") {
		t.Errorf("Plain text notification contains ANSI codes: %s", output)
	}
}

func TestNotifier_PrintNotification_Colored(t *testing.T) {
	// Create a mock server that returns a newer version
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := map[string]any{
			"tag_name":     "v2.0.0",
			"html_url":     "https://github.com/stackeye-io/stackeye-cli/releases/v2.0.0",
			"body":         "Release notes",
			"published_at": time.Now().Format(time.RFC3339),
			"assets":       []any{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	updater := sdkupdate.NewUpdater("stackeye-io/stackeye-cli", "1.0.0",
		sdkupdate.WithAPIURL(server.URL))
	checker := sdkupdate.NewChecker(updater, sdkupdate.WithCacheDir(tmpDir))

	var buf bytes.Buffer
	notifier := &Notifier{
		checker:        checker,
		currentVersion: "1.0.0",
		writer:         &buf,
		colorEnabled:   true,
	}

	ctx := context.Background()
	notifier.StartCheck(ctx)

	// Wait for check to complete
	result, err := checker.WaitForBackgroundResult(2 * time.Second)
	if err != nil {
		t.Fatalf("WaitForBackgroundResult failed: %v", err)
	}
	if result == nil || !result.HasUpdate {
		t.Fatal("Expected update available")
	}

	notifier.PrintNotification()

	output := buf.String()

	// Verify ANSI escape codes are present in colored mode
	if !strings.Contains(output, "\033[") {
		t.Errorf("Colored notification missing ANSI codes: %s", output)
	}
}

func TestNotifier_BackgroundCheckTimeout(t *testing.T) {
	// Create a slow server that takes longer than the timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		release := map[string]any{
			"tag_name": "v2.0.0",
		}
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	updater := sdkupdate.NewUpdater("stackeye-io/stackeye-cli", "1.0.0",
		sdkupdate.WithAPIURL(server.URL))
	checker := sdkupdate.NewChecker(updater)

	var buf bytes.Buffer
	notifier := &Notifier{
		checker:        checker,
		currentVersion: "1.0.0",
		writer:         &buf,
		colorEnabled:   false,
	}

	ctx := context.Background()
	notifier.StartCheck(ctx)

	// PrintNotification should return quickly (within BackgroundTimeout)
	// even if the check is still in progress
	start := time.Now()
	printed := notifier.PrintNotification()
	duration := time.Since(start)

	if printed {
		t.Error("PrintNotification() = true, want false (timeout)")
	}

	// Should complete within the timeout + some margin
	if duration > BackgroundTimeout+100*time.Millisecond {
		t.Errorf("PrintNotification took %v, want <= %v", duration, BackgroundTimeout+100*time.Millisecond)
	}
}

func TestNewNotifier_Options(t *testing.T) {
	var buf bytes.Buffer

	notifier := NewNotifier("1.0.0",
		WithWriter(&buf),
		WithColor(false),
	)

	if notifier.writer != &buf {
		t.Error("WithWriter option not applied")
	}

	if notifier.colorEnabled {
		t.Error("WithColor(false) option not applied")
	}

	if notifier.currentVersion != "1.0.0" {
		t.Errorf("currentVersion = %q, want %q", notifier.currentVersion, "1.0.0")
	}
}
