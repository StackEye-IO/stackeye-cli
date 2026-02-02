// Package cmd implements the CLI commands for StackEye.
// Task #8066
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-cli/internal/api"
	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
)

func TestNewLabelCreateCmd(t *testing.T) {
	cmd := NewLabelCreateCmd()

	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	// Verify command structure
	if cmd.Use != "create <key>" {
		t.Errorf("expected Use='create <key>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}

	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}

	// Verify required positional args
	if cmd.Args == nil {
		t.Error("expected Args validator to be set")
	}

	// Verify RunE is set
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}

	// Verify flags exist
	flags := cmd.Flags()
	if flags.Lookup("display-name") == nil {
		t.Error("expected --display-name flag")
	}
	if flags.Lookup("description") == nil {
		t.Error("expected --description flag")
	}
	if flags.Lookup("color") == nil {
		t.Error("expected --color flag")
	}
}

func TestValidateLabelKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		// Valid keys
		{name: "simple key", key: "env", wantErr: false},
		{name: "key with hyphen", key: "service-tier", wantErr: false},
		{name: "key with numbers", key: "tier1", wantErr: false},
		{name: "numeric start", key: "1tier", wantErr: false},
		{name: "single char", key: "a", wantErr: false},
		{name: "single digit", key: "1", wantErr: false},
		{name: "multi hyphen", key: "my-service-tier", wantErr: false},

		// Invalid keys
		{name: "empty key", key: "", wantErr: true},
		{name: "uppercase", key: "Env", wantErr: true},
		{name: "underscore", key: "my_key", wantErr: true},
		{name: "starts with hyphen", key: "-env", wantErr: true},
		{name: "ends with hyphen", key: "env-", wantErr: true},
		{name: "special char", key: "env@prod", wantErr: true},
		{name: "space", key: "my key", wantErr: true},
		{name: "dot", key: "my.key", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLabelKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLabelKey(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
			}
		})
	}
}

func TestValidateLabelKey_MaxLength(t *testing.T) {
	// Test exact max length (63 chars)
	maxKey := "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz1"
	if len(maxKey) != 63 {
		t.Fatalf("test key should be 63 chars, got %d", len(maxKey))
	}
	if err := validateLabelKey(maxKey); err != nil {
		t.Errorf("expected 63-char key to be valid, got error: %v", err)
	}

	// Test over max length (64 chars)
	overMaxKey := maxKey + "a"
	if err := validateLabelKey(overMaxKey); err == nil {
		t.Error("expected 64-char key to be invalid")
	}
}

func TestValidateHexColor(t *testing.T) {
	tests := []struct {
		name    string
		color   string
		wantErr bool
	}{
		// Valid colors
		{name: "lowercase hex", color: "#10b981", wantErr: false},
		{name: "uppercase hex", color: "#10B981", wantErr: false},
		{name: "mixed case", color: "#AbCdEf", wantErr: false},
		{name: "all zeros", color: "#000000", wantErr: false},
		{name: "all f", color: "#ffffff", wantErr: false},
		{name: "red", color: "#ff0000", wantErr: false},
		{name: "green", color: "#00ff00", wantErr: false},
		{name: "blue", color: "#0000ff", wantErr: false},

		// Invalid colors
		{name: "no hash", color: "10B981", wantErr: true},
		{name: "too short", color: "#10B98", wantErr: true},
		{name: "too long", color: "#10B9811", wantErr: true},
		{name: "invalid char g", color: "#10B98g", wantErr: true},
		{name: "3-char shorthand", color: "#fff", wantErr: true},
		{name: "empty", color: "", wantErr: true},
		{name: "spaces", color: "# 10B98", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHexColor(tt.color)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHexColor(%q) error = %v, wantErr %v", tt.color, err, tt.wantErr)
			}
		})
	}
}

// setupMockAPIServer creates a mock server for testing label create operations.
// Returns the server and a cleanup function.
func setupMockAPIServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
	t.Helper()
	server := httptest.NewServer(handler)

	// Save current config getter
	oldGetter := api.SetConfigGetter(nil)

	// Create SDK config with mock server URL
	cfg := config.NewConfig()
	cfg.CurrentContext = "test-context"
	cfg.SetContext("test-context", &config.Context{
		APIURL: server.URL,
		APIKey: "se_test1234567890123456789012345678901234567890123456789012345678901234",
	})

	// Set config getter to return our test config
	api.SetConfigGetter(func() *config.Config {
		return cfg
	})

	cleanup := func() {
		server.Close()
		// Restore original config getter
		api.SetConfigGetter(oldGetter)
	}

	return server, cleanup
}

func TestRunLabelCreate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/label-keys" || r.Method != "POST" {
			http.NotFound(w, r)
			return
		}

		// Parse and verify request body
		var req client.CreateLabelKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if req.Key != "env" {
			t.Errorf("expected key 'env', got %q", req.Key)
		}

		// Return successful response
		displayName := "Environment"
		resp := map[string]interface{}{
			"label_key": map[string]interface{}{
				"id":              1,
				"organization_id": "test-org-123",
				"key":             req.Key,
				"display_name":    displayName,
				"description":     nil,
				"color":           "#6B7280",
				"values_in_use":   []string{},
				"probe_count":     0,
				"created_at":      "2026-01-15T10:30:00Z",
				"updated_at":      "2026-01-15T10:30:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	})

	_, cleanup := setupMockAPIServer(t, handler)
	defer cleanup()

	// Capture output - the real test is that it doesn't error
	err := runLabelCreate(context.Background(), "env", "Environment", "", "")
	if err != nil {
		t.Errorf("runLabelCreate returned unexpected error: %v", err)
	}
}

func TestRunLabelCreate_DuplicateKey(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/label-keys" || r.Method != "POST" {
			http.NotFound(w, r)
			return
		}

		// Return 409 Conflict for duplicate key
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error": map[string]interface{}{
				"code":    "key_exists",
				"message": "label key already exists for this organization",
			},
		})
	})

	_, cleanup := setupMockAPIServer(t, handler)
	defer cleanup()

	err := runLabelCreate(context.Background(), "env", "", "", "")
	if err == nil {
		t.Error("expected error for duplicate key, got nil")
	}

	// Verify error message indicates the issue
	if err != nil && !strings.Contains(err.Error(), "failed to create label key") {
		t.Errorf("expected error to contain 'failed to create label key', got: %v", err)
	}
}

func TestRunLabelCreate_InvalidKeyRejectedLocally(t *testing.T) {
	// This test verifies that invalid keys are rejected before any API call
	// No mock server needed - validation happens client-side

	tests := []struct {
		name    string
		key     string
		wantErr string
	}{
		{
			name:    "uppercase rejected",
			key:     "Env",
			wantErr: "invalid key format",
		},
		{
			name:    "empty rejected",
			key:     "",
			wantErr: "label key is required",
		},
		{
			name:    "starts with hyphen rejected",
			key:     "-env",
			wantErr: "invalid key format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runLabelCreate(context.Background(), tt.key, "", "", "")
			if err == nil {
				t.Error("expected error for invalid key")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error to contain %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestRunLabelCreate_InvalidColorRejectedLocally(t *testing.T) {
	// Verify invalid colors are rejected before API call
	tests := []struct {
		name    string
		color   string
		wantErr string
	}{
		{
			name:    "no hash",
			color:   "10B981",
			wantErr: "7-character hex code",
		},
		{
			name:    "too short",
			color:   "#fff",
			wantErr: "7-character hex code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runLabelCreate(context.Background(), "validkey", "", "", tt.color)
			if err == nil {
				t.Error("expected error for invalid color")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error to contain %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestRunLabelCreate_AllFieldsPassedToAPI(t *testing.T) {
	var receivedReq client.CreateLabelKeyRequest

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/label-keys" || r.Method != "POST" {
			http.NotFound(w, r)
			return
		}

		// Capture the request for verification
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r.Body)
		_ = json.Unmarshal(buf.Bytes(), &receivedReq)

		// Return successful response
		resp := map[string]interface{}{
			"label_key": map[string]interface{}{
				"id":              1,
				"organization_id": "test-org-123",
				"key":             receivedReq.Key,
				"display_name":    receivedReq.DisplayName,
				"description":     receivedReq.Description,
				"color":           "#10B981",
				"values_in_use":   []string{},
				"probe_count":     0,
				"created_at":      "2026-01-15T10:30:00Z",
				"updated_at":      "2026-01-15T10:30:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	})

	_, cleanup := setupMockAPIServer(t, handler)
	defer cleanup()

	err := runLabelCreate(context.Background(), "env", "Environment", "Deployment environment", "#10B981")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify all fields were passed to API
	if receivedReq.Key != "env" {
		t.Errorf("expected key 'env', got %q", receivedReq.Key)
	}
	if receivedReq.DisplayName == nil || *receivedReq.DisplayName != "Environment" {
		t.Errorf("expected display_name 'Environment', got %v", receivedReq.DisplayName)
	}
	if receivedReq.Description == nil || *receivedReq.Description != "Deployment environment" {
		t.Errorf("expected description 'Deployment environment', got %v", receivedReq.Description)
	}
	if receivedReq.Color == nil || *receivedReq.Color != "#10B981" {
		t.Errorf("expected color '#10B981', got %v", receivedReq.Color)
	}
}
