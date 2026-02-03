package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

func TestNewProbeImportCmd(t *testing.T) {
	cmd := NewProbeImportCmd()

	if cmd.Use != "import" {
		t.Errorf("expected Use='import', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if !strings.Contains(cmd.Short, "Import") {
		t.Errorf("expected Short to mention Import, got %q", cmd.Short)
	}
}

func TestNewProbeImportCmd_Flags(t *testing.T) {
	cmd := NewProbeImportCmd()

	expectedFlags := []struct {
		name         string
		defaultValue string
	}{
		{"file", ""},
		{"format", ""},
		{"dry-run", "false"},
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

func TestNewProbeImportCmd_Aliases(t *testing.T) {
	cmd := NewProbeImportCmd()

	if len(cmd.Aliases) == 0 {
		t.Error("expected at least one alias")
		return
	}

	found := false
	for _, a := range cmd.Aliases {
		if a == "imp" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected alias 'imp' not found")
	}
}

func TestNewProbeImportCmd_Long(t *testing.T) {
	cmd := NewProbeImportCmd()

	long := cmd.Long
	keywords := []string{"Import", "JSON", "YAML", "--file", "--dry-run", "Duplicate"}
	for _, kw := range keywords {
		if !strings.Contains(long, kw) {
			t.Errorf("expected Long description to contain %q", kw)
		}
	}
}

func TestNewProbeImportCmd_FileRequired(t *testing.T) {
	cmd := NewProbeImportCmd()

	f := cmd.Flags().Lookup("file")
	if f == nil {
		t.Fatal("expected flag 'file' not found")
	}

	// The flag should be marked as required via annotations
	ann := f.Annotations
	if ann == nil {
		t.Error("expected annotations on required flag")
	}
}

func TestResolveImportFormat_FromFlag(t *testing.T) {
	tests := []struct {
		flag     string
		expected string
	}{
		{"yaml", "yaml"},
		{"YAML", "yaml"},
		{"json", "json"},
		{"JSON", "json"},
	}

	for _, tt := range tests {
		format, err := resolveImportFormat("file.txt", tt.flag)
		if err != nil {
			t.Errorf("resolveImportFormat(%q): unexpected error: %v", tt.flag, err)
			continue
		}
		if format != tt.expected {
			t.Errorf("resolveImportFormat(%q): expected %q, got %q", tt.flag, tt.expected, format)
		}
	}
}

func TestResolveImportFormat_InvalidFlag(t *testing.T) {
	_, err := resolveImportFormat("file.txt", "xml")
	if err == nil {
		t.Error("expected error for invalid format flag")
	}
	if !strings.Contains(err.Error(), "invalid value") {
		t.Errorf("expected 'invalid value' error, got: %v", err)
	}
}

func TestResolveImportFormat_FromExtension(t *testing.T) {
	tests := []struct {
		file     string
		expected string
	}{
		{"probes.yaml", "yaml"},
		{"probes.yml", "yaml"},
		{"probes.YAML", "yaml"},
		{"probes.json", "json"},
		{"probes.JSON", "json"},
	}

	for _, tt := range tests {
		format, err := resolveImportFormat(tt.file, "")
		if err != nil {
			t.Errorf("resolveImportFormat(%q): unexpected error: %v", tt.file, err)
			continue
		}
		if format != tt.expected {
			t.Errorf("resolveImportFormat(%q): expected %q, got %q", tt.file, tt.expected, format)
		}
	}
}

func TestResolveImportFormat_UnknownExtension(t *testing.T) {
	_, err := resolveImportFormat("probes.txt", "")
	if err == nil {
		t.Error("expected error for unknown extension")
	}
	if !strings.Contains(err.Error(), "cannot detect format") {
		t.Errorf("expected 'cannot detect format' error, got: %v", err)
	}
}

func TestReadProbeConfigs_JSON(t *testing.T) {
	configs := []probeExportConfig{
		{
			Name:                "Test Probe",
			URL:                 "https://example.com",
			CheckType:           "http",
			Method:              "GET",
			TimeoutMs:           10000,
			IntervalSeconds:     60,
			ExpectedStatusCodes: []int{200},
		},
	}

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	tmpFile := filepath.Join(t.TempDir(), "probes.json")
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	result, err := readProbeConfigs(tmpFile, "json")
	if err != nil {
		t.Fatalf("readProbeConfigs: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 config, got %d", len(result))
	}
	if result[0].Name != "Test Probe" {
		t.Errorf("expected Name='Test Probe', got %q", result[0].Name)
	}
}

func TestReadProbeConfigs_YAML(t *testing.T) {
	configs := []probeExportConfig{
		{
			Name:                "YAML Probe",
			URL:                 "https://example.com",
			CheckType:           "http",
			Method:              "GET",
			TimeoutMs:           10000,
			IntervalSeconds:     60,
			ExpectedStatusCodes: []int{200},
		},
	}

	data, err := yaml.Marshal(configs)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	tmpFile := filepath.Join(t.TempDir(), "probes.yaml")
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	result, err := readProbeConfigs(tmpFile, "yaml")
	if err != nil {
		t.Fatalf("readProbeConfigs: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 config, got %d", len(result))
	}
	if result[0].Name != "YAML Probe" {
		t.Errorf("expected Name='YAML Probe', got %q", result[0].Name)
	}
}

func TestReadProbeConfigs_FileNotFound(t *testing.T) {
	_, err := readProbeConfigs("/nonexistent/file.json", "json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("expected 'failed to read file' error, got: %v", err)
	}
}

func TestReadProbeConfigs_EmptyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.json")
	if err := os.WriteFile(tmpFile, []byte{}, 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	_, err := readProbeConfigs(tmpFile, "json")
	if err == nil {
		t.Error("expected error for empty file")
	}
	if !strings.Contains(err.Error(), "is empty") {
		t.Errorf("expected 'is empty' error, got: %v", err)
	}
}

func TestReadProbeConfigs_InvalidJSON(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(tmpFile, []byte("not json"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	_, err := readProbeConfigs(tmpFile, "json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse JSON") {
		t.Errorf("expected 'failed to parse JSON' error, got: %v", err)
	}
}

func TestReadProbeConfigs_InvalidYAML(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "bad.yaml")
	// Write invalid YAML that can't be parsed as a list
	if err := os.WriteFile(tmpFile, []byte(":\n  :\n    - [invalid"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	_, err := readProbeConfigs(tmpFile, "yaml")
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "failed to parse YAML") {
		t.Errorf("expected 'failed to parse YAML' error, got: %v", err)
	}
}

func TestValidateProbeConfigs_Valid(t *testing.T) {
	configs := []probeExportConfig{
		{
			Name:                "Valid Probe",
			URL:                 "https://example.com",
			CheckType:           "http",
			Method:              "GET",
			TimeoutMs:           10000,
			IntervalSeconds:     60,
			ExpectedStatusCodes: []int{200},
		},
	}

	err := validateProbeConfigs(configs)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateProbeConfigs_MissingName(t *testing.T) {
	configs := []probeExportConfig{
		{
			URL:       "https://example.com",
			CheckType: "http",
		},
	}

	err := validateProbeConfigs(configs)
	if err == nil {
		t.Error("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("expected 'name is required' error, got: %v", err)
	}
}

func TestValidateProbeConfigs_MissingURL(t *testing.T) {
	configs := []probeExportConfig{
		{
			Name:      "No URL",
			CheckType: "http",
		},
	}

	err := validateProbeConfigs(configs)
	if err == nil {
		t.Error("expected error for missing URL")
	}
	if !strings.Contains(err.Error(), "url is required") {
		t.Errorf("expected 'url is required' error, got: %v", err)
	}
}

func TestValidateProbeConfigs_MissingCheckType(t *testing.T) {
	configs := []probeExportConfig{
		{
			Name: "No Type",
			URL:  "https://example.com",
		},
	}

	err := validateProbeConfigs(configs)
	if err == nil {
		t.Error("expected error for missing check_type")
	}
	if !strings.Contains(err.Error(), "check_type is required") {
		t.Errorf("expected 'check_type is required' error, got: %v", err)
	}
}

func TestValidateProbeConfigs_InvalidCheckType(t *testing.T) {
	configs := []probeExportConfig{
		{
			Name:      "Bad Type",
			URL:       "https://example.com",
			CheckType: "invalid",
		},
	}

	err := validateProbeConfigs(configs)
	if err == nil {
		t.Error("expected error for invalid check type")
	}
	if !strings.Contains(err.Error(), "invalid value") {
		t.Errorf("expected 'invalid value' error, got: %v", err)
	}
}

func TestValidateProbeConfigs_InvalidURL(t *testing.T) {
	configs := []probeExportConfig{
		{
			Name:      "Bad URL",
			URL:       "not-a-url",
			CheckType: "http",
		},
	}

	err := validateProbeConfigs(configs)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestValidateProbeConfigs_InvalidInterval(t *testing.T) {
	configs := []probeExportConfig{
		{
			Name:            "Bad Interval",
			URL:             "https://example.com",
			CheckType:       "http",
			IntervalSeconds: 5, // Too low
		},
	}

	err := validateProbeConfigs(configs)
	if err == nil {
		t.Error("expected error for invalid interval")
	}
	if !strings.Contains(err.Error(), "interval_seconds") {
		t.Errorf("expected 'interval_seconds' in error, got: %v", err)
	}
}

func TestValidateProbeConfigs_InvalidTimeout(t *testing.T) {
	configs := []probeExportConfig{
		{
			Name:      "Bad Timeout",
			URL:       "https://example.com",
			CheckType: "http",
			TimeoutMs: 500, // Less than 1000
		},
	}

	err := validateProbeConfigs(configs)
	if err == nil {
		t.Error("expected error for invalid timeout")
	}
	if !strings.Contains(err.Error(), "timeout_ms") {
		t.Errorf("expected 'timeout_ms' in error, got: %v", err)
	}
}

func TestValidateProbeConfigs_ZeroTimeoutAndInterval(t *testing.T) {
	// Zero values should pass (defaults will be applied during conversion)
	configs := []probeExportConfig{
		{
			Name:      "Zero Defaults",
			URL:       "https://example.com",
			CheckType: "http",
		},
	}

	err := validateProbeConfigs(configs)
	if err != nil {
		t.Errorf("expected no error for zero timeout/interval (defaults apply), got: %v", err)
	}
}

func TestConvertExportConfigToCreateRequest(t *testing.T) {
	channelID := uuid.New()
	body := "test body"
	kwCheck := "healthy"
	kwType := "contains"
	jpCheck := "$.status"
	jpExpected := "ok"
	labelVal := "production"

	cfg := &probeExportConfig{
		Name:                   "Import Test",
		URL:                    "https://example.com/health",
		CheckType:              "http",
		Method:                 "POST",
		Headers:                map[string]string{"Authorization": "Bearer token"},
		Body:                   &body,
		TimeoutMs:              15000,
		IntervalSeconds:        120,
		Regions:                []string{"us-east-1", "eu-west-1"},
		ExpectedStatusCodes:    []int{200, 201},
		KeywordCheck:           &kwCheck,
		KeywordCheckType:       &kwType,
		JSONPathCheck:          &jpCheck,
		JSONPathExpected:       &jpExpected,
		SSLCheckEnabled:        true,
		SSLExpiryThresholdDays: 30,
		FollowRedirects:        true,
		MaxRedirects:           5,
		AlertChannelIDs:        []string{channelID.String()},
		Labels:                 []probeExportLabel{{Key: "env", Value: &labelVal}},
	}

	req := convertExportConfigToCreateRequest(cfg)

	if req.Name != "Import Test" {
		t.Errorf("expected Name='Import Test', got %q", req.Name)
	}
	if req.URL != "https://example.com/health" {
		t.Errorf("expected URL, got %q", req.URL)
	}
	if req.CheckType != client.CheckType("http") {
		t.Errorf("expected CheckType='http', got %q", req.CheckType)
	}
	if req.Method != "POST" {
		t.Errorf("expected Method='POST', got %q", req.Method)
	}
	if req.TimeoutMs != 15000 {
		t.Errorf("expected TimeoutMs=15000, got %d", req.TimeoutMs)
	}
	if req.IntervalSeconds != 120 {
		t.Errorf("expected IntervalSeconds=120, got %d", req.IntervalSeconds)
	}
	if len(req.Regions) != 2 {
		t.Errorf("expected 2 regions, got %d", len(req.Regions))
	}
	if len(req.ExpectedStatusCodes) != 2 {
		t.Errorf("expected 2 status codes, got %d", len(req.ExpectedStatusCodes))
	}
	if req.Body == nil || *req.Body != "test body" {
		t.Error("expected Body='test body'")
	}
	if req.KeywordCheck == nil || *req.KeywordCheck != "healthy" {
		t.Error("expected KeywordCheck='healthy'")
	}
	if req.KeywordCheckType == nil || *req.KeywordCheckType != "contains" {
		t.Error("expected KeywordCheckType='contains'")
	}
	if req.JSONPathCheck == nil || *req.JSONPathCheck != "$.status" {
		t.Error("expected JSONPathCheck='$.status'")
	}
	if req.JSONPathExpected == nil || *req.JSONPathExpected != "ok" {
		t.Error("expected JSONPathExpected='ok'")
	}
	if !req.SSLCheckEnabled {
		t.Error("expected SSLCheckEnabled=true")
	}
	if req.SSLExpiryThresholdDays != 30 {
		t.Errorf("expected SSLExpiryThresholdDays=30, got %d", req.SSLExpiryThresholdDays)
	}
	if req.FollowRedirects == nil || !*req.FollowRedirects {
		t.Error("expected FollowRedirects=true")
	}
	if req.MaxRedirects != 5 {
		t.Errorf("expected MaxRedirects=5, got %d", req.MaxRedirects)
	}

	// Check headers were serialized to JSON
	if req.Headers == "" {
		t.Error("expected non-empty Headers JSON string")
	}
	var parsedHeaders map[string]string
	if err := json.Unmarshal([]byte(req.Headers), &parsedHeaders); err != nil {
		t.Errorf("failed to parse headers JSON: %v", err)
	}
	if parsedHeaders["Authorization"] != "Bearer token" {
		t.Errorf("expected Authorization header, got %v", parsedHeaders)
	}

	// Check channel IDs were converted
	if len(req.AlertChannelIDs) != 1 {
		t.Errorf("expected 1 channel ID, got %d", len(req.AlertChannelIDs))
	}
	if req.AlertChannelIDs[0] != channelID {
		t.Errorf("expected channel ID %s, got %s", channelID, req.AlertChannelIDs[0])
	}
}

func TestConvertExportConfigToCreateRequest_Defaults(t *testing.T) {
	cfg := &probeExportConfig{
		Name:      "Minimal",
		URL:       "https://example.com",
		CheckType: "http",
	}

	req := convertExportConfigToCreateRequest(cfg)

	if req.Method != "GET" {
		t.Errorf("expected default Method='GET', got %q", req.Method)
	}
	if req.TimeoutMs != 10000 {
		t.Errorf("expected default TimeoutMs=10000, got %d", req.TimeoutMs)
	}
	if req.IntervalSeconds != 60 {
		t.Errorf("expected default IntervalSeconds=60, got %d", req.IntervalSeconds)
	}
	if len(req.ExpectedStatusCodes) != 1 || req.ExpectedStatusCodes[0] != 200 {
		t.Errorf("expected default ExpectedStatusCodes=[200], got %v", req.ExpectedStatusCodes)
	}
	if req.Headers != "" {
		t.Errorf("expected empty Headers for no headers, got %q", req.Headers)
	}
	if req.Body != nil {
		t.Error("expected nil Body")
	}
	if req.AlertChannelIDs != nil {
		t.Error("expected nil AlertChannelIDs")
	}
}

func TestConvertExportConfigToCreateRequest_InvalidChannelID(t *testing.T) {
	cfg := &probeExportConfig{
		Name:            "Bad Channel",
		URL:             "https://example.com",
		CheckType:       "http",
		AlertChannelIDs: []string{"not-a-uuid"},
	}

	req := convertExportConfigToCreateRequest(cfg)

	// Invalid UUIDs should be skipped with a warning
	if req.AlertChannelIDs != nil {
		t.Errorf("expected nil AlertChannelIDs for invalid UUID, got %v", req.AlertChannelIDs)
	}
}

func TestRunProbeImport_FileNotFound(t *testing.T) {
	flags := &probeImportFlags{
		file: "/nonexistent/file.yaml",
	}

	err := runProbeImport(t.Context(), flags)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestRunProbeImport_EmptyConfigs(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.yaml")
	if err := os.WriteFile(tmpFile, []byte("[]"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	flags := &probeImportFlags{
		file: tmpFile,
	}

	err := runProbeImport(t.Context(), flags)
	if err == nil {
		t.Error("expected error for empty configs")
	}
	if !strings.Contains(err.Error(), "no probe configurations found") {
		t.Errorf("expected 'no probe configurations found' error, got: %v", err)
	}
}

func TestRunProbeImport_ValidationError(t *testing.T) {
	configs := []probeExportConfig{
		{
			Name:      "Invalid",
			URL:       "not-a-url",
			CheckType: "http",
		},
	}

	data, _ := yaml.Marshal(configs)
	tmpFile := filepath.Join(t.TempDir(), "invalid.yaml")
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	flags := &probeImportFlags{
		file: tmpFile,
	}

	err := runProbeImport(t.Context(), flags)
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestRunProbeImport_DryRun(t *testing.T) {
	labelVal := "prod"
	configs := []probeExportConfig{
		{
			Name:                "DryRun Probe",
			URL:                 "https://example.com",
			CheckType:           "http",
			Method:              "GET",
			TimeoutMs:           10000,
			IntervalSeconds:     60,
			ExpectedStatusCodes: []int{200},
			Regions:             []string{"us-east-1"},
			Labels: []probeExportLabel{
				{Key: "env", Value: &labelVal},
			},
		},
	}

	data, _ := yaml.Marshal(configs)
	tmpFile := filepath.Join(t.TempDir(), "dryrun.yaml")
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	flags := &probeImportFlags{
		file:   tmpFile,
		dryRun: true,
	}

	// Dry run should succeed without API client
	err := runProbeImport(t.Context(), flags)
	if err != nil {
		t.Errorf("expected dry run to succeed, got: %v", err)
	}
}

func TestRunProbeImport_DryRunMultiple(t *testing.T) {
	configs := []probeExportConfig{
		{
			Name:                "Probe 1",
			URL:                 "https://one.example.com",
			CheckType:           "http",
			Method:              "GET",
			TimeoutMs:           10000,
			IntervalSeconds:     60,
			ExpectedStatusCodes: []int{200},
		},
		{
			Name:                "Probe 2",
			URL:                 "https://two.example.com",
			CheckType:           "http",
			Method:              "POST",
			TimeoutMs:           5000,
			IntervalSeconds:     120,
			ExpectedStatusCodes: []int{200, 201},
		},
	}

	data, _ := json.MarshalIndent(configs, "", "  ")
	tmpFile := filepath.Join(t.TempDir(), "multi.json")
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	flags := &probeImportFlags{
		file:   tmpFile,
		dryRun: true,
	}

	err := runProbeImport(t.Context(), flags)
	if err != nil {
		t.Errorf("expected dry run to succeed, got: %v", err)
	}
}

func TestRunProbeImport_NoAPIClientDryRunSucceeds(t *testing.T) {
	// Without API config, dry run should still work (no API calls)
	configs := []probeExportConfig{
		{
			Name:                "No API",
			URL:                 "https://example.com",
			CheckType:           "http",
			Method:              "GET",
			TimeoutMs:           10000,
			IntervalSeconds:     60,
			ExpectedStatusCodes: []int{200},
		},
	}

	data, _ := yaml.Marshal(configs)
	tmpFile := filepath.Join(t.TempDir(), "noapi.yaml")
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	flags := &probeImportFlags{
		file:   tmpFile,
		dryRun: true,
	}

	err := runProbeImport(t.Context(), flags)
	if err != nil {
		t.Errorf("dry run should not require API: %v", err)
	}
}

func TestRunProbeImport_NoAPIClientNonDryRunFails(t *testing.T) {
	// Without API config, non-dry-run should fail at API client init
	configs := []probeExportConfig{
		{
			Name:                "No API",
			URL:                 "https://example.com",
			CheckType:           "http",
			Method:              "GET",
			TimeoutMs:           10000,
			IntervalSeconds:     60,
			ExpectedStatusCodes: []int{200},
		},
	}

	data, _ := yaml.Marshal(configs)
	tmpFile := filepath.Join(t.TempDir(), "noapi.yaml")
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	flags := &probeImportFlags{
		file: tmpFile,
	}

	err := runProbeImport(t.Context(), flags)
	if err == nil {
		t.Error("expected error without API config")
	}
}

func TestProbeExportConfig_RoundTrip_JSON(t *testing.T) {
	// Verify that export -> JSON -> import produces equivalent configs
	body := "test body"
	kwCheck := "healthy"
	kwType := "contains"

	original := []probeExportConfig{
		{
			Name:                   "Round Trip",
			URL:                    "https://example.com",
			CheckType:              "http",
			Method:                 "GET",
			Headers:                map[string]string{"X-Custom": "value"},
			Body:                   &body,
			TimeoutMs:              10000,
			IntervalSeconds:        60,
			Regions:                []string{"us-east-1"},
			ExpectedStatusCodes:    []int{200, 201},
			KeywordCheck:           &kwCheck,
			KeywordCheckType:       &kwType,
			SSLCheckEnabled:        true,
			SSLExpiryThresholdDays: 14,
			FollowRedirects:        true,
			MaxRedirects:           10,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var imported []probeExportConfig
	if err := json.Unmarshal(data, &imported); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(imported) != 1 {
		t.Fatalf("expected 1 config, got %d", len(imported))
	}
	if imported[0].Name != original[0].Name {
		t.Errorf("Name mismatch: %q vs %q", imported[0].Name, original[0].Name)
	}
	if imported[0].CheckType != original[0].CheckType {
		t.Errorf("CheckType mismatch: %q vs %q", imported[0].CheckType, original[0].CheckType)
	}
}

func TestProbeExportConfig_RoundTrip_YAML(t *testing.T) {
	original := []probeExportConfig{
		{
			Name:                "YAML Round Trip",
			URL:                 "https://example.com",
			CheckType:           "http",
			Method:              "GET",
			TimeoutMs:           10000,
			IntervalSeconds:     60,
			ExpectedStatusCodes: []int{200},
		},
	}

	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var imported []probeExportConfig
	if err := yaml.Unmarshal(data, &imported); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(imported) != 1 {
		t.Fatalf("expected 1 config, got %d", len(imported))
	}
	if imported[0].Name != original[0].Name {
		t.Errorf("Name mismatch: %q vs %q", imported[0].Name, original[0].Name)
	}
}
