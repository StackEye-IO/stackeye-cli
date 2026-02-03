package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

func TestNewProbeExportCmd(t *testing.T) {
	cmd := NewProbeExportCmd()

	if cmd.Use != "export" {
		t.Errorf("expected Use='export', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if !strings.Contains(cmd.Short, "Export") {
		t.Errorf("expected Short to mention Export, got %q", cmd.Short)
	}
}

func TestNewProbeExportCmd_Flags(t *testing.T) {
	cmd := NewProbeExportCmd()

	expectedFlags := []struct {
		name         string
		defaultValue string
	}{
		{"format", "yaml"},
		{"file", ""},
		{"probe-ids", ""},
		{"status", ""},
		{"labels", ""},
	}

	for _, ef := range expectedFlags {
		f := cmd.Flags().Lookup(ef.name)
		if f == nil {
			t.Errorf("expected flag %q not found", ef.name)
			continue
		}
		if f.DefValue != ef.defaultValue {
			t.Errorf("flag %q: expected default %q, got %q", ef.name, ef.defaultValue, f.DefValue)
		}
	}
}

func TestNewProbeExportCmd_Aliases(t *testing.T) {
	cmd := NewProbeExportCmd()

	if len(cmd.Aliases) == 0 {
		t.Error("expected at least one alias")
		return
	}

	found := false
	for _, a := range cmd.Aliases {
		if a == "exp" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected alias 'exp' not found")
	}
}

func TestNewProbeExportCmd_Long(t *testing.T) {
	cmd := NewProbeExportCmd()

	long := cmd.Long
	keywords := []string{"backup", "migration", "YAML", "JSON", "--probe-ids", "--format", "--file"}
	for _, kw := range keywords {
		if !strings.Contains(long, kw) {
			t.Errorf("expected Long description to contain %q", kw)
		}
	}
}

func TestConvertProbeToExportConfig(t *testing.T) {
	probeID := uuid.New()
	channelID := uuid.New()
	body := "test body"
	kwCheck := "healthy"
	kwType := "contains"
	labelVal := "production"

	probe := &client.Probe{
		ID:                     probeID,
		Name:                   "Test Probe",
		URL:                    "https://example.com/health",
		CheckType:              client.CheckTypeHTTP,
		Method:                 "GET",
		Headers:                `{"Authorization":"Bearer token"}`,
		Body:                   &body,
		TimeoutMs:              10000,
		IntervalSeconds:        60,
		Regions:                []string{"us-east-1", "eu-west-1"},
		ExpectedStatusCodes:    []int{200, 201},
		KeywordCheck:           &kwCheck,
		KeywordCheckType:       &kwType,
		SSLCheckEnabled:        true,
		SSLExpiryThresholdDays: 30,
		FollowRedirects:        true,
		MaxRedirects:           5,
		AlertChannelIDs:        []uuid.UUID{channelID},
		Labels: []client.ProbeLabel{
			{Key: "env", Value: &labelVal},
		},
	}

	cfg := convertProbeToExportConfig(probe)

	if cfg.Name != "Test Probe" {
		t.Errorf("expected Name='Test Probe', got %q", cfg.Name)
	}
	if cfg.URL != "https://example.com/health" {
		t.Errorf("expected URL='https://example.com/health', got %q", cfg.URL)
	}
	if cfg.CheckType != "http" {
		t.Errorf("expected CheckType='http', got %q", cfg.CheckType)
	}
	if cfg.Method != "GET" {
		t.Errorf("expected Method='GET', got %q", cfg.Method)
	}
	if cfg.TimeoutMs != 10000 {
		t.Errorf("expected TimeoutMs=10000, got %d", cfg.TimeoutMs)
	}
	if cfg.IntervalSeconds != 60 {
		t.Errorf("expected IntervalSeconds=60, got %d", cfg.IntervalSeconds)
	}
	if len(cfg.Regions) != 2 {
		t.Errorf("expected 2 regions, got %d", len(cfg.Regions))
	}
	if len(cfg.ExpectedStatusCodes) != 2 {
		t.Errorf("expected 2 status codes, got %d", len(cfg.ExpectedStatusCodes))
	}
	if cfg.Body == nil || *cfg.Body != "test body" {
		t.Error("expected Body='test body'")
	}
	if cfg.KeywordCheck == nil || *cfg.KeywordCheck != "healthy" {
		t.Error("expected KeywordCheck='healthy'")
	}
	if cfg.KeywordCheckType == nil || *cfg.KeywordCheckType != "contains" {
		t.Error("expected KeywordCheckType='contains'")
	}
	if !cfg.SSLCheckEnabled {
		t.Error("expected SSLCheckEnabled=true")
	}
	if cfg.SSLExpiryThresholdDays != 30 {
		t.Errorf("expected SSLExpiryThresholdDays=30, got %d", cfg.SSLExpiryThresholdDays)
	}
	if !cfg.FollowRedirects {
		t.Error("expected FollowRedirects=true")
	}
	if cfg.MaxRedirects != 5 {
		t.Errorf("expected MaxRedirects=5, got %d", cfg.MaxRedirects)
	}

	// Check headers were parsed
	if len(cfg.Headers) != 1 {
		t.Errorf("expected 1 header, got %d", len(cfg.Headers))
	}
	if cfg.Headers["Authorization"] != "Bearer token" {
		t.Errorf("expected Authorization header, got %v", cfg.Headers)
	}

	// Check channel IDs
	if len(cfg.AlertChannelIDs) != 1 {
		t.Errorf("expected 1 channel ID, got %d", len(cfg.AlertChannelIDs))
	}
	if cfg.AlertChannelIDs[0] != channelID.String() {
		t.Errorf("expected channel ID %s, got %s", channelID, cfg.AlertChannelIDs[0])
	}

	// Check labels
	if len(cfg.Labels) != 1 {
		t.Errorf("expected 1 label, got %d", len(cfg.Labels))
	}
	if cfg.Labels[0].Key != "env" {
		t.Errorf("expected label key='env', got %q", cfg.Labels[0].Key)
	}
	if cfg.Labels[0].Value == nil || *cfg.Labels[0].Value != "production" {
		t.Error("expected label value='production'")
	}
}

func TestConvertProbeToExportConfig_EmptyHeaders(t *testing.T) {
	probe := &client.Probe{
		Name:                "Minimal Probe",
		URL:                 "https://example.com",
		CheckType:           client.CheckTypeHTTP,
		Method:              "GET",
		Headers:             "",
		TimeoutMs:           10000,
		IntervalSeconds:     60,
		ExpectedStatusCodes: []int{200},
	}

	cfg := convertProbeToExportConfig(probe)
	if cfg.Headers != nil {
		t.Errorf("expected nil headers for empty header string, got %v", cfg.Headers)
	}
}

func TestConvertProbeToExportConfig_EmptyBracesHeaders(t *testing.T) {
	probe := &client.Probe{
		Name:                "Minimal Probe",
		URL:                 "https://example.com",
		CheckType:           client.CheckTypeHTTP,
		Method:              "GET",
		Headers:             "{}",
		TimeoutMs:           10000,
		IntervalSeconds:     60,
		ExpectedStatusCodes: []int{200},
	}

	cfg := convertProbeToExportConfig(probe)
	if cfg.Headers != nil {
		t.Errorf("expected nil headers for '{}' header string, got %v", cfg.Headers)
	}
}

func TestConvertProbeToExportConfig_NoChannels(t *testing.T) {
	probe := &client.Probe{
		Name:                "No Channels",
		URL:                 "https://example.com",
		CheckType:           client.CheckTypeHTTP,
		Method:              "GET",
		TimeoutMs:           10000,
		IntervalSeconds:     60,
		ExpectedStatusCodes: []int{200},
	}

	cfg := convertProbeToExportConfig(probe)
	if cfg.AlertChannelIDs != nil {
		t.Errorf("expected nil channel IDs, got %v", cfg.AlertChannelIDs)
	}
}

func TestProbeExportConfig_JSONMarshal(t *testing.T) {
	cfg := probeExportConfig{
		Name:                "Test",
		URL:                 "https://example.com",
		CheckType:           "http",
		Method:              "GET",
		TimeoutMs:           10000,
		IntervalSeconds:     60,
		ExpectedStatusCodes: []int{200},
		SSLCheckEnabled:     true,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	var result probeExportConfig
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if result.Name != cfg.Name {
		t.Errorf("expected Name=%q, got %q", cfg.Name, result.Name)
	}
	if result.CheckType != cfg.CheckType {
		t.Errorf("expected CheckType=%q, got %q", cfg.CheckType, result.CheckType)
	}
}

func TestProbeExportConfig_YAMLMarshal(t *testing.T) {
	cfg := probeExportConfig{
		Name:                "Test",
		URL:                 "https://example.com",
		CheckType:           "http",
		Method:              "GET",
		TimeoutMs:           10000,
		IntervalSeconds:     60,
		ExpectedStatusCodes: []int{200},
		SSLCheckEnabled:     true,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal YAML: %v", err)
	}

	var result probeExportConfig
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}

	if result.Name != cfg.Name {
		t.Errorf("expected Name=%q, got %q", cfg.Name, result.Name)
	}
	if result.CheckType != cfg.CheckType {
		t.Errorf("expected CheckType=%q, got %q", cfg.CheckType, result.CheckType)
	}
}

func TestRunProbeExport_InvalidFormat(t *testing.T) {
	flags := &probeExportFlags{
		format: "xml",
	}

	err := runProbeExport(t.Context(), flags)
	if err == nil {
		t.Error("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid value") || !strings.Contains(err.Error(), "--format") {
		t.Errorf("expected 'invalid value' error for --format, got: %v", err)
	}
}

func TestRunProbeExport_InvalidStatus(t *testing.T) {
	flags := &probeExportFlags{
		format: "yaml",
		status: "invalid",
	}

	err := runProbeExport(t.Context(), flags)
	if err == nil {
		t.Error("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid value") || !strings.Contains(err.Error(), "--status") {
		t.Errorf("expected 'invalid value' error for --status, got: %v", err)
	}
}

func TestRunProbeExport_InvalidProbeID(t *testing.T) {
	flags := &probeExportFlags{
		format:   "yaml",
		probeIDs: "not-a-uuid",
	}

	err := runProbeExport(t.Context(), flags)
	if err == nil {
		t.Error("expected error for invalid probe ID")
	}
	if !strings.Contains(err.Error(), "invalid probe ID") {
		t.Errorf("expected 'invalid probe ID' error, got: %v", err)
	}
}

func TestRunProbeExport_EmptyProbeIDs(t *testing.T) {
	flags := &probeExportFlags{
		format:   "yaml",
		probeIDs: ",,,",
	}

	err := runProbeExport(t.Context(), flags)
	if err == nil {
		t.Error("expected error for empty probe IDs")
	}
	if !strings.Contains(err.Error(), "no valid IDs") {
		t.Errorf("expected 'no valid IDs' error, got: %v", err)
	}
}

func TestRunProbeExport_ValidStatusValues(t *testing.T) {
	// These should fail at the API client step (no config), not at validation
	statuses := []string{"up", "down", "degraded", "paused", "pending"}

	for _, status := range statuses {
		flags := &probeExportFlags{
			format: "yaml",
			status: status,
		}

		err := runProbeExport(t.Context(), flags)
		if err == nil {
			t.Errorf("expected error for status %q (no API config), got nil", status)
			continue
		}
		// Should fail at API client initialization, not status validation
		if strings.Contains(err.Error(), "invalid value") && strings.Contains(err.Error(), "--status") {
			t.Errorf("status %q should be valid but got validation error: %v", status, err)
		}
	}
}
