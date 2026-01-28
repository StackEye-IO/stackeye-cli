package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetInfo(t *testing.T) {
	info := GetInfo()

	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	if info.GoVersion != runtime.Version() {
		t.Errorf("GoVersion = %q, want %q", info.GoVersion, runtime.Version())
	}

	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if info.Platform != expectedPlatform {
		t.Errorf("Platform = %q, want %q", info.Platform, expectedPlatform)
	}
}

func TestInfoString(t *testing.T) {
	// Override package variables for testing
	origVersion := Version
	origCommit := Commit
	origDate := Date
	origBuiltBy := BuiltBy
	defer func() {
		Version = origVersion
		Commit = origCommit
		Date = origDate
		BuiltBy = origBuiltBy
	}()

	Version = "1.2.3"
	Commit = "abc1234"
	Date = "2026-01-15T12:00:00Z"
	BuiltBy = "test"

	info := GetInfo()
	str := info.String()

	tests := []struct {
		name   string
		want   string
		wantIn bool
	}{
		{"version header", "stackeye version 1.2.3", true},
		{"commit", "Commit:     abc1234", true},
		{"date", "Built:      2026-01-15T12:00:00Z", true},
		{"built by", "Built by:   test", true},
		{"go version", "Go version:", true},
		{"platform", "Platform:", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if contains := strings.Contains(str, tt.want); contains != tt.wantIn {
				t.Errorf("String() contains %q = %v, want %v\nGot:\n%s", tt.want, contains, tt.wantIn, str)
			}
		})
	}
}

func TestInfoShort(t *testing.T) {
	origVersion := Version
	defer func() { Version = origVersion }()

	Version = "1.2.3"

	info := GetInfo()
	if got := info.Short(); got != "1.2.3" {
		t.Errorf("Short() = %q, want %q", got, "1.2.3")
	}
}

func TestIsDev(t *testing.T) {
	origVersion := Version
	defer func() { Version = origVersion }()

	tests := []struct {
		version string
		want    bool
	}{
		{"dev", true},
		{"1.0.0", false},
		{"1.0.0-dev", false},
		{"dev-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			Version = tt.version
			if got := IsDev(); got != tt.want {
				t.Errorf("IsDev() with version %q = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestIsDirty(t *testing.T) {
	origVersion := Version
	defer func() { Version = origVersion }()

	tests := []struct {
		version string
		want    bool
	}{
		{"1.0.0-dirty", true},
		{"v1.2.3-g1234abc-dirty", true},
		{"1.0.0", false},
		{"dirty", false},
		{"dirty-1.0.0", false},
		{"dev", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			Version = tt.version
			if got := IsDirty(); got != tt.want {
				t.Errorf("IsDirty() with version %q = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestDefaultValues(t *testing.T) {
	// Test that default values are set
	// Note: These test the initial values before any ldflags injection
	if Version != "dev" {
		t.Logf("Note: Version is %q (may have been injected via ldflags)", Version)
	}
	if Commit == "" {
		t.Error("Commit should have a default value")
	}
	if Date == "" {
		t.Error("Date should have a default value")
	}
	if BuiltBy == "" {
		t.Error("BuiltBy should have a default value")
	}
}
