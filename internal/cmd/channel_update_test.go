package cmd

import (
	"strings"
	"testing"
)

func TestNewChannelUpdateCmd(t *testing.T) {
	cmd := NewChannelUpdateCmd()

	if cmd.Use != "update <id>" {
		t.Errorf("expected Use='update <id>', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Short != "Update an existing notification channel" {
		t.Errorf("expected Short='Update an existing notification channel', got %q", cmd.Short)
	}
}

func TestNewChannelUpdateCmd_Args(t *testing.T) {
	cmd := NewChannelUpdateCmd()

	// Should require exactly 1 argument (the channel ID)
	if cmd.Args == nil {
		t.Error("expected Args validator to be set")
	}
}

func TestNewChannelUpdateCmd_Flags(t *testing.T) {
	cmd := NewChannelUpdateCmd()

	// Basic update flags
	basicFlags := []string{"name", "enabled"}
	for _, name := range basicFlags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("expected flag --%s to exist", name)
		}
	}

	// Type-specific flags (same as create)
	typeFlags := []string{
		"email", "webhook-url", "url", "method", "headers",
		"routing-key", "severity", "phone-number",
	}
	for _, name := range typeFlags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("expected flag --%s to exist", name)
		}
	}

	// File input flag
	flag := cmd.Flags().Lookup("from-file")
	if flag == nil {
		t.Error("expected flag --from-file to exist")
	}
}

func TestNewChannelUpdateCmd_Long(t *testing.T) {
	cmd := NewChannelUpdateCmd()

	long := cmd.Long

	// Should mention partial updates
	if !strings.Contains(long, "partial") {
		t.Error("expected Long description to mention partial updates")
	}

	// Should explain that only specified flags are updated
	if !strings.Contains(long, "only") && !strings.Contains(long, "Only") {
		t.Error("expected Long description to explain that only specified flags are updated")
	}

	// Should have usage examples
	if !strings.Contains(long, "stackeye channel update") {
		t.Error("expected Long description to contain example commands")
	}

	// Should mention --from-file option
	if !strings.Contains(long, "--from-file") {
		t.Error("expected Long description to mention --from-file option")
	}

	// Should mention that type cannot be changed
	if !strings.Contains(long, "type cannot be changed") {
		t.Error("expected Long description to mention that type cannot be changed")
	}

	// Should have examples for common update scenarios
	updateExamples := []string{
		"--name",
		"--enabled",
		"--email",
		"--webhook-url",
		"--severity",
	}
	for _, example := range updateExamples {
		if !strings.Contains(long, example) {
			t.Errorf("expected Long description to have example using %s", example)
		}
	}
}

func TestHasConfigFlags_NoFlags(t *testing.T) {
	cmd := NewChannelUpdateCmd()

	// By default no flags are changed
	if hasConfigFlags(cmd) {
		t.Error("expected hasConfigFlags to return false when no flags are changed")
	}
}

func TestHasConfigFlags_ConfigFlags(t *testing.T) {
	// Test each config flag individually
	configFlags := []string{
		"email", "webhook-url", "url", "method", "headers",
		"routing-key", "severity", "phone-number",
	}

	for _, flagName := range configFlags {
		t.Run(flagName, func(t *testing.T) {
			cmd := NewChannelUpdateCmd()

			// Simulate setting the flag
			flag := cmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Fatalf("flag --%s not found", flagName)
			}
			// Mark as changed by setting a value
			if err := cmd.Flags().Set(flagName, "test-value"); err != nil {
				t.Fatalf("failed to set flag --%s: %v", flagName, err)
			}

			if !hasConfigFlags(cmd) {
				t.Errorf("expected hasConfigFlags to return true when --%s is changed", flagName)
			}
		})
	}
}

func TestHasConfigFlags_NonConfigFlags(t *testing.T) {
	// Name and enabled are not config flags
	tests := []struct {
		flagName string
		value    string
	}{
		{"name", "test-name"},
		{"enabled", "true"},
	}

	for _, tc := range tests {
		t.Run(tc.flagName, func(t *testing.T) {
			cmd := NewChannelUpdateCmd()

			// Mark as changed (use appropriate value for flag type)
			if err := cmd.Flags().Set(tc.flagName, tc.value); err != nil {
				t.Fatalf("failed to set flag --%s: %v", tc.flagName, err)
			}

			if hasConfigFlags(cmd) {
				t.Errorf("expected hasConfigFlags to return false when only --%s is changed", tc.flagName)
			}
		})
	}
}

func TestChannelUpdateFlags_Structure(t *testing.T) {
	// Verify the flags struct has all expected pointer fields
	flags := &channelUpdateFlags{}

	// All these should be nil by default (pointer types)
	if flags.name != nil {
		t.Error("expected name to be nil by default")
	}
	if flags.enabled != nil {
		t.Error("expected enabled to be nil by default")
	}
	if flags.email != nil {
		t.Error("expected email to be nil by default")
	}
	if flags.webhookURL != nil {
		t.Error("expected webhookURL to be nil by default")
	}
	if flags.url != nil {
		t.Error("expected url to be nil by default")
	}
	if flags.method != nil {
		t.Error("expected method to be nil by default")
	}
	if flags.headers != nil {
		t.Error("expected headers to be nil by default")
	}
	if flags.routingKey != nil {
		t.Error("expected routingKey to be nil by default")
	}
	if flags.severity != nil {
		t.Error("expected severity to be nil by default")
	}
	if flags.phoneNumber != nil {
		t.Error("expected phoneNumber to be nil by default")
	}

	// fromFile is a value type, not pointer
	if flags.fromFile != "" {
		t.Error("expected fromFile to be empty by default")
	}
}

func TestChannelUpdateYAMLConfig_Structure(t *testing.T) {
	// Verify YAML config has optional fields via pointers
	cfg := channelUpdateYAMLConfig{}

	if cfg.Name != nil {
		t.Error("expected Name to be nil by default")
	}
	if cfg.Enabled != nil {
		t.Error("expected Enabled to be nil by default")
	}
	if cfg.Config != nil {
		t.Error("expected Config to be nil by default")
	}
}

func TestChannelUpdateTimeout(t *testing.T) {
	// Verify timeout constant is reasonable
	if channelUpdateTimeout <= 0 {
		t.Error("expected channelUpdateTimeout to be positive")
	}

	// Should be at least 10 seconds for network operations
	if channelUpdateTimeout.Seconds() < 10 {
		t.Errorf("expected channelUpdateTimeout to be at least 10 seconds, got %v", channelUpdateTimeout)
	}

	// Should not be excessively long (more than 5 minutes)
	if channelUpdateTimeout.Minutes() > 5 {
		t.Errorf("expected channelUpdateTimeout to be at most 5 minutes, got %v", channelUpdateTimeout)
	}
}

func TestStringPtrVar(t *testing.T) {
	var ptr *string
	result := stringPtrVar(&ptr)

	// After stringPtrVar, ptr is allocated (not nil) and points to empty string
	if ptr == nil {
		t.Error("expected ptr to be allocated after stringPtrVar")
	}
	if *ptr != "" {
		t.Errorf("expected ptr to point to empty string, got %q", *ptr)
	}

	// result and ptr should point to the same memory
	if result != ptr {
		t.Error("expected result and ptr to point to same memory")
	}

	// Setting a value via result should update ptr
	*result = "test"
	if *ptr != "test" {
		t.Errorf("expected ptr to be updated to 'test', got %q", *ptr)
	}
}

func TestBoolPtrVar(t *testing.T) {
	var ptr *bool
	result := boolPtrVar(&ptr)

	// After boolPtrVar, ptr is allocated (not nil) and points to false (zero value)
	if ptr == nil {
		t.Error("expected ptr to be allocated after boolPtrVar")
	}
	if *ptr != false {
		t.Errorf("expected ptr to point to false, got %v", *ptr)
	}

	// result and ptr should point to the same memory
	if result != ptr {
		t.Error("expected result and ptr to point to same memory")
	}

	// Setting a value via result should update ptr
	*result = true
	if *ptr != true {
		t.Errorf("expected ptr to be updated to true, got %v", *ptr)
	}
}

func TestNewChannelUpdateCmd_NoTypeFlag(t *testing.T) {
	cmd := NewChannelUpdateCmd()

	// Update command should NOT have a --type flag
	// because type cannot be changed after creation
	flag := cmd.Flags().Lookup("type")
	if flag != nil {
		t.Error("update command should not have --type flag (type cannot be changed)")
	}
}

func TestNewChannelUpdateCmd_Examples(t *testing.T) {
	cmd := NewChannelUpdateCmd()

	long := cmd.Long

	// Should have examples for enabling/disabling
	if !strings.Contains(long, "--enabled=false") {
		t.Error("expected example showing how to disable a channel")
	}
	if !strings.Contains(long, "--enabled=true") {
		t.Error("expected example showing how to enable a channel")
	}

	// Should show example with UUID-like ID
	if !strings.Contains(long, "550e8400-e29b-41d4-a716-446655440000") {
		t.Error("expected example with UUID-format channel ID")
	}
}
