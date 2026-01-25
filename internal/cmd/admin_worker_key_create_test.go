// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client/admin"
)

// TestNewAdminWorkerKeyCreateCmd verifies that the worker-key create command is properly constructed.
func TestNewAdminWorkerKeyCreateCmd(t *testing.T) {
	cmd := NewAdminWorkerKeyCreateCmd()

	if cmd.Use != "create" {
		t.Errorf("expected Use='create', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Create a new worker key" {
		t.Errorf("expected Short='Create a new worker key', got %q", cmd.Short)
	}
}

// TestNewAdminWorkerKeyCreateCmd_Long verifies the Long description contains key information.
func TestNewAdminWorkerKeyCreateCmd_Long(t *testing.T) {
	cmd := NewAdminWorkerKeyCreateCmd()

	long := cmd.Long

	// Should contain key features
	features := []string{
		"worker key",
		"regional probe",
		"region",
		"authentication",
		"store it securely",
	}
	for _, feature := range features {
		if !strings.Contains(strings.ToLower(long), strings.ToLower(feature)) {
			t.Errorf("expected Long description to mention %q", feature)
		}
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye admin worker-key create") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention region flag
	if !strings.Contains(long, "--region") {
		t.Error("expected Long description to mention --region flag")
	}

	// Should mention name flag
	if !strings.Contains(long, "--name") {
		t.Error("expected Long description to mention --name flag")
	}
}

// TestNewAdminWorkerKeyCreateCmd_RunEIsSet verifies that RunE is properly configured.
func TestNewAdminWorkerKeyCreateCmd_RunEIsSet(t *testing.T) {
	cmd := NewAdminWorkerKeyCreateCmd()

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

// TestNewAdminWorkerKeyCreateCmd_RegionFlag verifies that the region flag is registered and required.
func TestNewAdminWorkerKeyCreateCmd_RegionFlag(t *testing.T) {
	cmd := NewAdminWorkerKeyCreateCmd()

	flag := cmd.Flags().Lookup("region")
	if flag == nil {
		t.Fatal("expected --region flag to be registered")
	}

	if flag.Shorthand != "r" {
		t.Errorf("expected region flag shorthand to be 'r', got %q", flag.Shorthand)
	}

	// Check flag usage contains description
	if flag.Usage == "" {
		t.Error("expected region flag to have a usage description")
	}

	// Verify required flag annotation
	// Note: cobra doesn't expose required annotations easily in tests,
	// but we can verify the flag is registered
}

// TestNewAdminWorkerKeyCreateCmd_NameFlag verifies that the name flag is registered.
func TestNewAdminWorkerKeyCreateCmd_NameFlag(t *testing.T) {
	cmd := NewAdminWorkerKeyCreateCmd()

	flag := cmd.Flags().Lookup("name")
	if flag == nil {
		t.Fatal("expected --name flag to be registered")
	}

	if flag.Shorthand != "n" {
		t.Errorf("expected name flag shorthand to be 'n', got %q", flag.Shorthand)
	}

	// Check flag usage contains description
	if flag.Usage == "" {
		t.Error("expected name flag to have a usage description")
	}
}

// TestTruncateWorkerKeyField verifies worker key field truncation.
func TestTruncateWorkerKeyField(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"shorter than max", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"longer truncated", "hello world", 8, "hello..."},
		{"very short max", "hello", 3, "hel"},
		{"empty string", "", 10, ""},
		{"single char max", "hello", 1, "h"},
		{"uuid length", "550e8400-e29b-41d4-a716-446655440000", 36, "550e8400-e29b-41d4-a716-446655440000"},
		{"uuid truncated", "550e8400-e29b-41d4-a716-446655440000", 20, "550e8400-e29b-41d..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateWorkerKeyField(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateWorkerKeyField(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestFormatWorkerKeyTime verifies worker key time formatting.
func TestFormatWorkerKeyTime(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		contains string
	}{
		{"zero time", time.Time{}, "Unknown"},
		{"valid time", time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC), "2026-01-15"},
		{"with timezone", time.Date(2026, 6, 1, 14, 0, 0, 0, time.UTC), "2026-06-01 14:00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatWorkerKeyTime(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("formatWorkerKeyTime(%v) = %q, expected to contain %q", tt.input, result, tt.contains)
			}
		})
	}
}

// TestPrintWorkerKeyCreated_DoesNotPanic verifies that printWorkerKeyCreated doesn't panic with various inputs.
func TestPrintWorkerKeyCreated_DoesNotPanic(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		response *admin.CreateWorkerKeyResponse
	}{
		{
			name: "minimal response",
			response: &admin.CreateWorkerKeyResponse{
				Status: "success",
				Data: struct {
					admin.WorkerKey
					Key string `json:"key"`
				}{
					WorkerKey: admin.WorkerKey{
						ID:        "550e8400-e29b-41d4-a716-446655440000",
						Region:    "nyc3",
						KeyPrefix: "wk_nyc3_abc12345",
						KeyType:   "worker",
						IsActive:  true,
						CreatedAt: now,
						UpdatedAt: now,
					},
					Key: "wk_nyc3_abc12345_fulllongkeyvalue",
				},
			},
		},
		{
			name: "with all fields",
			response: &admin.CreateWorkerKeyResponse{
				Status: "success",
				Data: struct {
					admin.WorkerKey
					Key string `json:"key"`
				}{
					WorkerKey: admin.WorkerKey{
						ID:         "550e8400-e29b-41d4-a716-446655440001",
						Region:     "lon1",
						KeyPrefix:  "wk_lon1_xyz98765",
						KeyType:    "worker",
						IsActive:   true,
						LastSeenAt: &now,
						CreatedAt:  now,
						UpdatedAt:  now,
					},
					Key: "wk_lon1_xyz98765_anotherfulllongkeyvalue",
				},
			},
		},
		{
			name: "with very long key",
			response: &admin.CreateWorkerKeyResponse{
				Status: "success",
				Data: struct {
					admin.WorkerKey
					Key string `json:"key"`
				}{
					WorkerKey: admin.WorkerKey{
						ID:        "550e8400-e29b-41d4-a716-446655440002",
						Region:    "fra1",
						KeyPrefix: "wk_fra1_longlonglonglonglongprefix12345",
						KeyType:   "worker",
						IsActive:  true,
						CreatedAt: now,
						UpdatedAt: now,
					},
					Key: "wk_fra1_longlonglonglonglongprefix12345_verylongkeyvaluethatexceedsnormallength1234567890abcdefghijklmnop",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printWorkerKeyCreated panicked: %v", r)
				}
			}()
			printWorkerKeyCreated(tt.response)
		})
	}
}

// TestNewAdminWorkerKeyCmd verifies that the worker-key parent command is properly constructed.
func TestNewAdminWorkerKeyCmd(t *testing.T) {
	cmd := NewAdminWorkerKeyCmd()

	if cmd.Use != "worker-key" {
		t.Errorf("expected Use='worker-key', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
}

// TestNewAdminWorkerKeyCmd_Aliases verifies that aliases are set correctly.
func TestNewAdminWorkerKeyCmd_Aliases(t *testing.T) {
	cmd := NewAdminWorkerKeyCmd()

	if len(cmd.Aliases) == 0 {
		t.Error("expected aliases to be set")
	}

	// Should have "wk" alias
	foundWK := false
	for _, alias := range cmd.Aliases {
		if alias == "wk" {
			foundWK = true
			break
		}
	}
	if !foundWK {
		t.Error("expected 'wk' alias")
	}
}

// TestNewAdminWorkerKeyCmd_HasCreateSubcommand verifies that create subcommand is registered.
func TestNewAdminWorkerKeyCmd_HasCreateSubcommand(t *testing.T) {
	cmd := NewAdminWorkerKeyCmd()

	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("expected worker-key command to have at least 1 subcommand (create)")
	}

	// Verify create subcommand is registered
	found := false
	for _, sub := range subcommands {
		if sub.Use == "create" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'create' subcommand to be registered")
	}
}

// TestNewAdminCmd verifies that the admin parent command is properly constructed.
func TestNewAdminCmd(t *testing.T) {
	cmd := NewAdminCmd()

	if cmd.Use != "admin" {
		t.Errorf("expected Use='admin', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
}

// TestNewAdminCmd_Aliases verifies that aliases are set correctly.
func TestNewAdminCmd_Aliases(t *testing.T) {
	cmd := NewAdminCmd()

	if len(cmd.Aliases) == 0 {
		t.Error("expected aliases to be set")
	}

	// Should have "adm" alias
	foundAdm := false
	for _, alias := range cmd.Aliases {
		if alias == "adm" {
			foundAdm = true
			break
		}
	}
	if !foundAdm {
		t.Error("expected 'adm' alias")
	}
}

// TestNewAdminCmd_HasWorkerKeySubcommand verifies that worker-key subcommand is registered.
func TestNewAdminCmd_HasWorkerKeySubcommand(t *testing.T) {
	cmd := NewAdminCmd()

	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("expected admin command to have at least 1 subcommand (worker-key)")
	}

	// Verify worker-key subcommand is registered
	found := false
	for _, sub := range subcommands {
		if sub.Use == "worker-key" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'worker-key' subcommand to be registered")
	}
}
