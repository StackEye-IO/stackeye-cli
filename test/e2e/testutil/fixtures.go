package testutil

import (
	"time"
)

// Known fixture IDs for consistent testing
// Note: All IDs must be valid UUIDs (36 chars, proper hex format)
const (
	ProbeID1       = "11111111-1111-1111-1111-111111111111"
	ProbeID2       = "22222222-2222-2222-2222-222222222222"
	AlertID1       = "aaaaaaaa-1111-1111-1111-111111111111"
	ChannelID1     = "cccccccc-1111-1111-1111-111111111111"
	ChannelID2     = "cccccccc-2222-2222-2222-222222222222"
	OrganizationID = "00000000-0000-0000-0000-000000000001"
)

// ProbeListFixture returns a list of probes for testing.
// Note: Field names MUST match SDK's Probe struct JSON tags exactly.
func ProbeListFixture() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":                        ProbeID1,
			"name":                      "API Health Check",
			"check_type":                "http",
			"url":                       "https://api.example.com/health",
			"method":                    "GET",
			"status":                    "up",
			"interval_seconds":          60,
			"timeout_ms":                30000,
			"regions":                   []string{"us-east-1", "eu-west-1"},
			"expected_status_codes":     []int{200},
			"ssl_check_enabled":         true,
			"ssl_expiry_threshold_days": 14,
			"follow_redirects":          true,
			"max_redirects":             5,
			"is_paused":                 false,
			"created_at":                time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"updated_at":                time.Now().Format(time.RFC3339),
			"last_checked_at":           time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
			"uptime":                    99.95,
			"avg_response_time_ms":      145,
		},
		{
			"id":                        ProbeID2,
			"name":                      "Website Monitor",
			"check_type":                "http",
			"url":                       "https://www.example.com",
			"method":                    "GET",
			"status":                    "up",
			"interval_seconds":          120,
			"timeout_ms":                60000,
			"regions":                   []string{"us-west-2"},
			"expected_status_codes":     []int{200},
			"ssl_check_enabled":         true,
			"ssl_expiry_threshold_days": 14,
			"follow_redirects":          true,
			"max_redirects":             5,
			"is_paused":                 false,
			"created_at":                time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			"updated_at":                time.Now().Format(time.RFC3339),
			"last_checked_at":           time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
			"uptime":                    100.0,
			"avg_response_time_ms":      98,
		},
	}
}

// ProbeFixture returns a probe fixture by ID.
func ProbeFixture(id string) map[string]interface{} {
	probes := ProbeListFixture()
	for _, probe := range probes {
		if probe["id"] == id {
			return probe
		}
	}
	return nil
}

// NewProbeFixture returns a newly created probe fixture.
func NewProbeFixture() map[string]interface{} {
	return map[string]interface{}{
		"id":                        "33333333-3333-3333-3333-333333333333",
		"name":                      "New Probe",
		"check_type":                "http",
		"url":                       "https://new.example.com",
		"method":                    "GET",
		"status":                    "pending",
		"interval_seconds":          60,
		"timeout_ms":                30000,
		"regions":                   []string{"us-east-1"},
		"expected_status_codes":     []int{200},
		"ssl_check_enabled":         false,
		"ssl_expiry_threshold_days": 14,
		"follow_redirects":          true,
		"max_redirects":             5,
		"is_paused":                 false,
		"created_at":                time.Now().Format(time.RFC3339),
		"updated_at":                time.Now().Format(time.RFC3339),
	}
}

// ProbeTestResultFixture returns a probe test result fixture.
// Note: Field names MUST match SDK's ProbeTestResponse struct JSON tags exactly.
func ProbeTestResultFixture() map[string]interface{} {
	statusCode := 200
	return map[string]interface{}{
		"status":           "success",
		"response_time_ms": 145,
		"status_code":      statusCode,
		"checked_at":       time.Now().Format(time.RFC3339),
	}
}

// ProbeHistoryFixture returns probe history entries.
func ProbeHistoryFixture() []map[string]interface{} {
	now := time.Now()
	return []map[string]interface{}{
		{
			"id":          "hist-1",
			"probe_id":    ProbeID1,
			"status":      "up",
			"status_code": 200,
			"latency_ms":  145,
			"region":      "us-east-1",
			"checked_at":  now.Add(-1 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":          "hist-2",
			"probe_id":    ProbeID1,
			"status":      "up",
			"status_code": 200,
			"latency_ms":  152,
			"region":      "eu-west-1",
			"checked_at":  now.Add(-2 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":          "hist-3",
			"probe_id":    ProbeID1,
			"status":      "down",
			"status_code": 503,
			"latency_ms":  5000,
			"region":      "us-east-1",
			"error":       "Service Unavailable",
			"checked_at":  now.Add(-3 * time.Minute).Format(time.RFC3339),
		},
	}
}

// ProbeStatsFixture returns probe statistics fixture.
func ProbeStatsFixture() map[string]interface{} {
	return map[string]interface{}{
		"probe_id":        ProbeID1,
		"period":          "24h",
		"uptime_percent":  99.95,
		"avg_latency_ms":  145,
		"p50_latency_ms":  140,
		"p95_latency_ms":  180,
		"p99_latency_ms":  250,
		"total_checks":    1440,
		"successful":      1439,
		"failed":          1,
		"incidents":       1,
		"avg_recovery_ms": 180000,
	}
}

// AlertListFixture returns a list of alerts for testing.
// Note: Field names MUST match SDK's Alert struct JSON tags exactly.
func AlertListFixture() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":              AlertID1,
			"probe_id":        ProbeID1,
			"organization_id": OrganizationID,
			"status":          "active",
			"severity":        "critical",
			"alert_type":      "status_down",
			"message":         "Probe is down - Service Unavailable",
			"triggered_at":    time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			"created_at":      time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			"updated_at":      time.Now().Format(time.RFC3339),
			"probe": map[string]interface{}{
				"id":         ProbeID1,
				"name":       "API Health Check",
				"url":        "https://api.example.com/health",
				"check_type": "http",
			},
		},
	}
}

// AlertFixture returns an alert fixture by ID.
func AlertFixture(id string) map[string]interface{} {
	alerts := AlertListFixture()
	for _, alert := range alerts {
		if alert["id"] == id {
			return alert
		}
	}
	return nil
}

// AlertHistoryFixture returns alert history entries.
func AlertHistoryFixture() []map[string]interface{} {
	now := time.Now()
	return []map[string]interface{}{
		{
			"id":         "ahist-1",
			"alert_id":   AlertID1,
			"event":      "triggered",
			"message":    "Alert triggered: Probe is down",
			"created_at": now.Add(-30 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":         "ahist-2",
			"alert_id":   AlertID1,
			"event":      "notification_sent",
			"message":    "Notification sent to Slack channel",
			"created_at": now.Add(-29 * time.Minute).Format(time.RFC3339),
		},
	}
}

// ChannelListFixture returns a list of notification channels for testing.
// Note: Field names MUST match SDK's Channel struct JSON tags exactly.
func ChannelListFixture() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":              ChannelID1,
			"organization_id": OrganizationID,
			"name":            "Ops Slack",
			"type":            "slack",
			"enabled":         true,
			"probe_count":     2,
			"created_at":      time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339),
			"updated_at":      time.Now().Format(time.RFC3339),
			"config":          map[string]interface{}{"webhook_url": "https://hooks.slack.com/services/xxx"},
		},
		{
			"id":              ChannelID2,
			"organization_id": OrganizationID,
			"name":            "On-Call Email",
			"type":            "email",
			"enabled":         true,
			"probe_count":     1,
			"created_at":      time.Now().Add(-14 * 24 * time.Hour).Format(time.RFC3339),
			"updated_at":      time.Now().Format(time.RFC3339),
			"config":          map[string]interface{}{"address": "oncall@example.com"},
		},
	}
}

// ChannelFixture returns a channel fixture by ID.
func ChannelFixture(id string) map[string]interface{} {
	channels := ChannelListFixture()
	for _, channel := range channels {
		if channel["id"] == id {
			return channel
		}
	}
	return nil
}

// NewChannelFixture returns a newly created channel fixture.
func NewChannelFixture() map[string]interface{} {
	return map[string]interface{}{
		"id":              "dddd1111-1111-1111-1111-111111111111",
		"organization_id": OrganizationID,
		"name":            "New Channel",
		"type":            "email",
		"enabled":         true,
		"probe_count":     0,
		"created_at":      time.Now().Format(time.RFC3339),
		"updated_at":      time.Now().Format(time.RFC3339),
		"config":          map[string]interface{}{"address": "new@example.com"},
	}
}

// RegionListByContinent returns regions grouped by continent, matching SDK's RegionListResponse.Data format.
// SDK expects: map[string][]Region where key is continent slug like "north_america", "europe".
func RegionListByContinent() map[string][]map[string]interface{} {
	return map[string][]map[string]interface{}{
		"north_america": {
			{
				"id":           "us-east-1",
				"name":         "US East (N. Virginia)",
				"display_name": "US East",
				"country_code": "US",
			},
			{
				"id":           "us-west-2",
				"name":         "US West (Oregon)",
				"display_name": "US West",
				"country_code": "US",
			},
		},
		"europe": {
			{
				"id":           "eu-west-1",
				"name":         "EU West (Ireland)",
				"display_name": "Ireland",
				"country_code": "IE",
			},
		},
	}
}

// RegionListFixture returns a flat list of available regions (legacy format).
func RegionListFixture() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":           "us-east-1",
			"name":         "US East (N. Virginia)",
			"display_name": "US East",
			"country_code": "US",
		},
		{
			"id":           "us-west-2",
			"name":         "US West (Oregon)",
			"display_name": "US West",
			"country_code": "US",
		},
		{
			"id":           "eu-west-1",
			"name":         "EU West (Ireland)",
			"display_name": "Ireland",
			"country_code": "IE",
		},
	}
}
