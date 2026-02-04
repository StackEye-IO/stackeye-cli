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
	MuteID1        = "eeeeeeee-1111-1111-1111-111111111111"
	StatusPageID1  = "55555555-1111-1111-1111-111111111111"
	IncidentID1    = "66666666-1111-1111-1111-111111111111"
	TeamMemberID1  = "77777777-1111-1111-1111-111111111111"
	TeamMemberID2  = "77777777-2222-2222-2222-222222222222"
	InvitationID1  = "88888888-1111-1111-1111-111111111111"
	APIKeyID1      = "99999999-1111-1111-1111-111111111111"
	LabelID1       = "bbbbbbbb-1111-1111-1111-111111111111"
	LabelKeyID1    = "dddddddd-1111-1111-1111-111111111111"
	DependencyID1  = "ffffffff-1111-1111-1111-111111111111"
	UserID1        = "44444444-1111-1111-1111-111111111111"
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

// --- Alert Stats ---

// AlertStatsFixture returns alert statistics fixture.
func AlertStatsFixture() map[string]interface{} {
	return map[string]interface{}{
		"period":          "24h",
		"total_alerts":    5,
		"active_alerts":   2,
		"resolved_alerts": 3,
		"mtta_seconds":    120,
		"mttr_seconds":    600,
	}
}

// --- Alert Mutes ---

// AlertMuteListFixture returns a list of alert mutes for testing.
func AlertMuteListFixture() []map[string]interface{} {
	now := time.Now()
	return []map[string]interface{}{
		{
			"id":              MuteID1,
			"organization_id": OrganizationID,
			"scope":           "probe",
			"probe_id":        ProbeID1,
			"reason":          "Scheduled maintenance",
			"status":          "active",
			"starts_at":       now.Add(-1 * time.Hour).Format(time.RFC3339),
			"ends_at":         now.Add(1 * time.Hour).Format(time.RFC3339),
			"created_at":      now.Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}
}

// NewAlertMuteFixture returns a newly created alert mute fixture.
func NewAlertMuteFixture() map[string]interface{} {
	now := time.Now()
	return map[string]interface{}{
		"id":              "eeeeeeee-2222-2222-2222-222222222222",
		"organization_id": OrganizationID,
		"scope":           "organization",
		"reason":          "Emergency maintenance window",
		"status":          "active",
		"starts_at":       now.Format(time.RFC3339),
		"ends_at":         now.Add(2 * time.Hour).Format(time.RFC3339),
		"created_at":      now.Format(time.RFC3339),
	}
}

// --- Labels ---

// LabelListFixture returns a list of labels for a probe.
func LabelListFixture() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":    LabelID1,
			"key":   "env",
			"value": "production",
		},
		{
			"id":    "bbbbbbbb-2222-2222-2222-222222222222",
			"key":   "team",
			"value": "platform",
		},
	}
}

// NewLabelFixture returns a newly created label fixture.
func NewLabelFixture() map[string]interface{} {
	return map[string]interface{}{
		"id":    "bbbbbbbb-3333-3333-3333-333333333333",
		"key":   "service",
		"value": "api",
	}
}

// LabelKeyListFixture returns a list of label keys for an organization.
func LabelKeyListFixture() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":    LabelKeyID1,
			"key":   "env",
			"count": 5,
		},
		{
			"id":    "dddddddd-2222-2222-2222-222222222222",
			"key":   "team",
			"count": 3,
		},
	}
}

// NewLabelKeyFixture returns a newly created label key fixture.
func NewLabelKeyFixture() map[string]interface{} {
	return map[string]interface{}{
		"id":    "dddddddd-3333-3333-3333-333333333333",
		"key":   "region",
		"count": 0,
	}
}

// --- Probe Dependencies ---

// ProbeDependencyListFixture returns a list of probe dependencies.
func ProbeDependencyListFixture() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":              DependencyID1,
			"probe_id":        ProbeID1,
			"parent_probe_id": ProbeID2,
			"parent_probe": map[string]interface{}{
				"id":   ProbeID2,
				"name": "Website Monitor",
			},
		},
	}
}

// NewProbeDependencyFixture returns a newly created dependency fixture.
func NewProbeDependencyFixture() map[string]interface{} {
	return map[string]interface{}{
		"id":              "ffffffff-2222-2222-2222-222222222222",
		"probe_id":        ProbeID1,
		"parent_probe_id": ProbeID2,
	}
}

// DependencyTreeFixture returns a dependency tree for an organization.
func DependencyTreeFixture() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"probe_id":   ProbeID2,
			"probe_name": "Website Monitor",
			"children": []map[string]interface{}{
				{
					"probe_id":   ProbeID1,
					"probe_name": "API Health Check",
					"children":   []map[string]interface{}{},
				},
			},
		},
	}
}

// --- Team Members ---

// TeamMemberListFixture returns a list of team members.
func TeamMemberListFixture() []map[string]interface{} {
	now := time.Now()
	return []map[string]interface{}{
		{
			"id":         TeamMemberID1,
			"user_id":    UserID1,
			"email":      "admin@example.com",
			"name":       "Admin User",
			"role":       "owner",
			"status":     "active",
			"joined_at":  now.Add(-30 * 24 * time.Hour).Format(time.RFC3339),
			"created_at": now.Add(-30 * 24 * time.Hour).Format(time.RFC3339),
		},
		{
			"id":         TeamMemberID2,
			"user_id":    "44444444-2222-2222-2222-222222222222",
			"email":      "dev@example.com",
			"name":       "Dev User",
			"role":       "member",
			"status":     "active",
			"joined_at":  now.Add(-7 * 24 * time.Hour).Format(time.RFC3339),
			"created_at": now.Add(-7 * 24 * time.Hour).Format(time.RFC3339),
		},
	}
}

// NewTeamMemberFixture returns a newly invited team member fixture.
func NewTeamMemberFixture() map[string]interface{} {
	return map[string]interface{}{
		"id":         "77777777-3333-3333-3333-333333333333",
		"email":      "newmember@example.com",
		"name":       "New Member",
		"role":       "member",
		"status":     "invited",
		"created_at": time.Now().Format(time.RFC3339),
	}
}

// --- Team Invitations ---

// TeamInvitationListFixture returns a list of pending invitations.
func TeamInvitationListFixture() []map[string]interface{} {
	now := time.Now()
	return []map[string]interface{}{
		{
			"id":         InvitationID1,
			"email":      "pending@example.com",
			"role":       "member",
			"status":     "pending",
			"invited_by": "admin@example.com",
			"expires_at": now.Add(7 * 24 * time.Hour).Format(time.RFC3339),
			"created_at": now.Add(-1 * 24 * time.Hour).Format(time.RFC3339),
		},
	}
}

// --- Status Pages ---

// StatusPageListFixture returns a list of status pages.
func StatusPageListFixture() []map[string]interface{} {
	now := time.Now()
	return []map[string]interface{}{
		{
			"id":               StatusPageID1,
			"organization_id":  OrganizationID,
			"name":             "Main Status Page",
			"slug":             "main",
			"custom_domain":    "",
			"is_published":     true,
			"subscriber_count": 42,
			"probe_count":      3,
			"created_at":       now.Add(-60 * 24 * time.Hour).Format(time.RFC3339),
			"updated_at":       now.Format(time.RFC3339),
		},
	}
}

// StatusPageFixture returns a status page by ID.
func StatusPageFixture(id string) map[string]interface{} {
	pages := StatusPageListFixture()
	for _, page := range pages {
		if page["id"] == id {
			return page
		}
	}
	return nil
}

// NewStatusPageFixture returns a newly created status page fixture.
func NewStatusPageFixture() map[string]interface{} {
	return map[string]interface{}{
		"id":               "55555555-2222-2222-2222-222222222222",
		"organization_id":  OrganizationID,
		"name":             "New Status Page",
		"slug":             "new-page",
		"custom_domain":    "",
		"is_published":     false,
		"subscriber_count": 0,
		"probe_count":      0,
		"created_at":       time.Now().Format(time.RFC3339),
		"updated_at":       time.Now().Format(time.RFC3339),
	}
}

// --- Status Page Incidents ---

// StatusPageIncidentListFixture returns a list of incidents for a status page.
func StatusPageIncidentListFixture() []map[string]interface{} {
	now := time.Now()
	return []map[string]interface{}{
		{
			"id":             IncidentID1,
			"status_page_id": StatusPageID1,
			"title":          "API Degraded Performance",
			"status":         "investigating",
			"impact":         "minor",
			"message":        "We are investigating reports of increased latency.",
			"created_at":     now.Add(-2 * time.Hour).Format(time.RFC3339),
			"updated_at":     now.Format(time.RFC3339),
		},
	}
}

// NewStatusPageIncidentFixture returns a newly created incident fixture.
func NewStatusPageIncidentFixture() map[string]interface{} {
	return map[string]interface{}{
		"id":             "66666666-2222-2222-2222-222222222222",
		"status_page_id": StatusPageID1,
		"title":          "New Incident",
		"status":         "investigating",
		"impact":         "major",
		"message":        "We are aware of the issue and working on a fix.",
		"created_at":     time.Now().Format(time.RFC3339),
		"updated_at":     time.Now().Format(time.RFC3339),
	}
}

// --- Dashboard ---

// DashboardStatsFixture returns dashboard statistics.
func DashboardStatsFixture() map[string]interface{} {
	return map[string]interface{}{
		"total_probes":      10,
		"probes_up":         8,
		"probes_down":       1,
		"probes_paused":     1,
		"open_alerts":       2,
		"response_time_p95": 180,
		"uptime_percentage": 99.95,
		"period":            "24h",
	}
}

// --- Billing ---

// BillingInfoFixture returns billing info.
func BillingInfoFixture() map[string]interface{} {
	return map[string]interface{}{
		"plan":                 "pro",
		"status":               "active",
		"billing_email":        "billing@example.com",
		"current_period_start": time.Now().Add(-15 * 24 * time.Hour).Format(time.RFC3339),
		"current_period_end":   time.Now().Add(15 * 24 * time.Hour).Format(time.RFC3339),
	}
}

// BillingStatusFixture returns billing/subscription status.
func BillingStatusFixture() map[string]interface{} {
	return map[string]interface{}{
		"plan":      "pro",
		"status":    "active",
		"trial":     false,
		"trial_end": nil,
		"cancel_at": nil,
	}
}

// BillingUsageFixture returns current billing usage.
func BillingUsageFixture() map[string]interface{} {
	return map[string]interface{}{
		"monitors_used":  10,
		"monitors_limit": 100,
		"team_members":   2,
		"status_pages":   1,
		"api_keys":       1,
		"period":         "current",
	}
}

// --- API Keys ---

// APIKeyListFixture returns a list of API keys.
func APIKeyListFixture() []map[string]interface{} {
	now := time.Now()
	return []map[string]interface{}{
		{
			"id":           APIKeyID1,
			"name":         "CI/CD Pipeline",
			"prefix":       "se_abc1",
			"last_used_at": now.Add(-1 * time.Hour).Format(time.RFC3339),
			"created_at":   now.Add(-30 * 24 * time.Hour).Format(time.RFC3339),
		},
	}
}

// NewAPIKeyFixture returns a newly created API key fixture.
func NewAPIKeyFixture() map[string]interface{} {
	return map[string]interface{}{
		"id":         "99999999-2222-2222-2222-222222222222",
		"name":       "New API Key",
		"prefix":     "se_xyz9",
		"key":        "se_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"created_at": time.Now().Format(time.RFC3339),
	}
}

// --- User Profile ---

// UserProfileFixture returns a user profile.
func UserProfileFixture() map[string]interface{} {
	return map[string]interface{}{
		"id":              UserID1,
		"email":           "admin@example.com",
		"name":            "Admin User",
		"organization_id": OrganizationID,
		"role":            "owner",
		"created_at":      time.Now().Add(-90 * 24 * time.Hour).Format(time.RFC3339),
	}
}

// CLIVerifyFixture returns a CLI credential verification response.
// Matches the CLIVerifyResponse struct in stackeye-go-sdk/client/cli_auth.go.
func CLIVerifyFixture() map[string]interface{} {
	return map[string]interface{}{
		"valid":             true,
		"organization_id":   OrganizationID,
		"organization_name": "Test Organization",
		"auth_type":         "api_key",
		"user_email":        "admin@example.com",
		"user_name":         "Admin User",
		"is_platform_admin": false,
	}
}

// UserOrganizationListFixture returns a list of organizations the user belongs to.
func UserOrganizationListFixture() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":   OrganizationID,
			"name": "Example Org",
			"slug": "example-org",
			"role": "owner",
		},
	}
}
