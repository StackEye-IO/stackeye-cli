package cmd

import (
	"testing"
)

func TestNewConfigCmd(t *testing.T) {
	cmd := NewConfigCmd()

	if cmd.Use != "config" {
		t.Errorf("expected Use='config', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify set-key subcommand is registered
	setKeyCmd, _, err := cmd.Find([]string{"set-key"})
	if err != nil {
		t.Errorf("set-key subcommand not found: %v", err)
	}
	if setKeyCmd.Use != "set-key [api-key]" {
		t.Errorf("expected set-key Use='set-key [api-key]', got %q", setKeyCmd.Use)
	}
}

func TestNewConfigSetKeyCmd_Flags(t *testing.T) {
	cmd := NewConfigCmd()
	setKeyCmd, _, _ := cmd.Find([]string{"set-key"})

	// Check --verify flag exists
	verifyFlag := setKeyCmd.Flags().Lookup("verify")
	if verifyFlag == nil {
		t.Fatal("expected --verify flag to be defined")
	}
	if verifyFlag.DefValue != "false" {
		t.Errorf("expected --verify default to be 'false', got %q", verifyFlag.DefValue)
	}

	// Check --context flag exists
	contextFlag := setKeyCmd.Flags().Lookup("context")
	if contextFlag == nil {
		t.Fatal("expected --context flag to be defined")
	}
	if contextFlag.DefValue != "" {
		t.Errorf("expected --context default to be empty, got %q", contextFlag.DefValue)
	}
}

func TestGetAPIKey_FromArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name:    "valid key from args",
			args:    []string{"se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
			want:    "se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			wantErr: false,
		},
		{
			name:    "empty args with no-input returns error",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "empty string arg with no-input returns error",
			args:    []string{""},
			wantErr: true,
		},
	}

	// Set noInput to true for testing non-interactive mode
	originalNoInput := noInput
	noInput = true
	defer func() { noInput = originalNoInput }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAPIKey(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getAPIKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunConfigSetKey_InvalidFormat(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing se_ prefix",
			args:    []string{"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
			wantErr: "invalid API key format",
		},
		{
			name:    "too short",
			args:    []string{"se_abc123"},
			wantErr: "invalid API key format",
		},
		{
			name:    "invalid characters",
			args:    []string{"se_ghij56789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
			wantErr: "invalid API key format",
		},
		{
			name:    "wrong prefix",
			args:    []string{"sk_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
			wantErr: "invalid API key format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &configSetKeyFlags{}
			err := runConfigSetKey(flags, tt.args)
			if err == nil {
				t.Error("expected error, got nil")
				return
			}
			if tt.wantErr != "" {
				if !contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   string
	}{
		{
			name:   "empty key",
			apiKey: "",
			want:   "(not set)",
		},
		{
			name:   "valid key shows prefix and last 4",
			apiKey: "se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			want:   "se_****...cdef",
		},
		{
			name:   "short key over 7 chars",
			apiKey: "se_abcd1234",
			want:   "se_****...1234",
		},
		{
			name:   "key between 5 and 7 chars",
			apiKey: "abc123",
			want:   "****c123",
		},
		{
			name:   "key exactly 4 chars",
			apiKey: "abcd",
			want:   "****",
		},
		{
			name:   "key less than 4 chars",
			apiKey: "abc",
			want:   "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskAPIKey(tt.apiKey)
			if got != tt.want {
				t.Errorf("maskAPIKey(%q) = %q, want %q", tt.apiKey, got, tt.want)
			}
		})
	}
}

func TestNewConfigGetCmd(t *testing.T) {
	cmd := NewConfigCmd()

	// Verify get subcommand is registered
	getCmd, _, err := cmd.Find([]string{"get"})
	if err != nil {
		t.Fatalf("get subcommand not found: %v", err)
	}
	if getCmd.Use != "get" {
		t.Errorf("expected get Use='get', got %q", getCmd.Use)
	}
	if getCmd.Short == "" {
		t.Error("expected get Short description to be set")
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
