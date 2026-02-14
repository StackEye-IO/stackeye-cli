// Package testutil provides test fixtures and mock data for StackEye CLI tests.
// This package contains pre-defined sample data that matches API response shapes,
// making it easy to write consistent and realistic tests.
package testutil

import (
	"encoding/json"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
	"github.com/google/uuid"
)

// Fixed UUIDs for reproducible test assertions.
// These are valid UUIDv4 values that can be used across tests.
var (
	// Probe UUIDs
	ProbeID1 = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	ProbeID2 = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ProbeID3 = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	ProbeID4 = uuid.MustParse("44444444-4444-4444-4444-444444444444")

	// Alert UUIDs
	AlertID1 = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	AlertID2 = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	AlertID3 = uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	// Channel UUIDs
	ChannelID1 = uuid.MustParse("c1111111-1111-1111-1111-111111111111")
	ChannelID2 = uuid.MustParse("c2222222-2222-2222-2222-222222222222")
	ChannelID3 = uuid.MustParse("c3333333-3333-3333-3333-333333333333")
	ChannelID4 = uuid.MustParse("c4444444-4444-4444-4444-444444444444")

	// Organization UUIDs
	OrgID1 = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	OrgID2 = uuid.MustParse("00000000-0000-0000-0000-000000000002")

	// User UUIDs
	UserID1 = uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	UserID2 = uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
)

// Fixed timestamps for reproducible test assertions.
var (
	// BaseTime is a fixed reference time for tests (2024-01-15 10:30:00 UTC)
	BaseTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	// TimeOneHourAgo is BaseTime minus 1 hour
	TimeOneHourAgo = BaseTime.Add(-1 * time.Hour)

	// TimeOneDayAgo is BaseTime minus 24 hours
	TimeOneDayAgo = BaseTime.Add(-24 * time.Hour)

	// TimeOneWeekAgo is BaseTime minus 7 days
	TimeOneWeekAgo = BaseTime.Add(-7 * 24 * time.Hour)
)

// ============================================================================
// Probe Fixtures
// ============================================================================

// SampleHTTPProbe returns a sample HTTP probe with default values.
func SampleHTTPProbe() client.Probe {
	return client.Probe{
		ID:                     ProbeID1,
		Name:                   "API Health Check",
		URL:                    "https://api.example.com/health",
		Method:                 "GET",
		Headers:                "",
		TimeoutMs:              10000,
		IntervalSeconds:        60,
		Regions:                []string{"nyc3", "sfo3", "lon1"},
		ExpectedStatusCodes:    []int{200},
		SSLCheckEnabled:        true,
		SSLExpiryThresholdDays: 30,
		FollowRedirects:        true,
		MaxRedirects:           5,
		CheckType:              client.CheckTypeHTTP,
		Status:                 "up",
		IsPaused:               false,
		LastCheckedAt:          &BaseTime,
		CreatedAt:              TimeOneWeekAgo,
		UpdatedAt:              BaseTime,
		AlertChannelIDs:        []uuid.UUID{ChannelID1, ChannelID2},
		IsUnreachable:          false,
		ParentCount:            0,
		ChildCount:             2,
		Uptime:                 99.95,
		AvgResponseTimeMs:      125.5,
	}
}

// SamplePingProbe returns a sample Ping probe.
func SamplePingProbe() client.Probe {
	return client.Probe{
		ID:                ProbeID2,
		Name:              "Database Server Ping",
		URL:               "db.internal.example.com",
		TimeoutMs:         5000,
		IntervalSeconds:   30,
		Regions:           []string{"nyc3"},
		CheckType:         client.CheckTypePing,
		Status:            "up",
		IsPaused:          false,
		LastCheckedAt:     &BaseTime,
		CreatedAt:         TimeOneWeekAgo,
		UpdatedAt:         BaseTime,
		IsUnreachable:     false,
		ParentCount:       1,
		ChildCount:        0,
		Uptime:            100.0,
		AvgResponseTimeMs: 15.2,
	}
}

// SampleTCPProbe returns a sample TCP probe.
func SampleTCPProbe() client.Probe {
	return client.Probe{
		ID:                ProbeID3,
		Name:              "Redis Connection",
		URL:               "redis.example.com:6379",
		TimeoutMs:         3000,
		IntervalSeconds:   60,
		Regions:           []string{"nyc3", "sfo3"},
		CheckType:         client.CheckTypeTCP,
		Status:            "up",
		IsPaused:          false,
		LastCheckedAt:     &BaseTime,
		CreatedAt:         TimeOneWeekAgo,
		UpdatedAt:         BaseTime,
		IsUnreachable:     false,
		ParentCount:       0,
		ChildCount:        0,
		Uptime:            99.99,
		AvgResponseTimeMs: 8.5,
	}
}

// SampleDNSProbe returns a sample DNS resolve probe.
func SampleDNSProbe() client.Probe {
	keyword := "93.184.216.34"
	keywordType := "contains"
	return client.Probe{
		ID:                ProbeID4,
		Name:              "DNS Resolution Check",
		URL:               "example.com",
		TimeoutMs:         5000,
		IntervalSeconds:   300,
		Regions:           []string{"nyc3", "lon1", "sgp1"},
		CheckType:         client.CheckTypeDNSResolve,
		KeywordCheck:      &keyword,
		KeywordCheckType:  &keywordType,
		Status:            "up",
		IsPaused:          false,
		LastCheckedAt:     &BaseTime,
		CreatedAt:         TimeOneWeekAgo,
		UpdatedAt:         BaseTime,
		IsUnreachable:     false,
		ParentCount:       0,
		ChildCount:        0,
		Uptime:            100.0,
		AvgResponseTimeMs: 45.0,
	}
}

// SampleDownProbe returns a probe in DOWN status.
func SampleDownProbe() client.Probe {
	probe := SampleHTTPProbe()
	probe.ID = uuid.MustParse("dddddddd-1111-1111-1111-111111111111")
	probe.Name = "Failing Service"
	probe.URL = "https://failing.example.com/health"
	probe.Status = "down"
	probe.Uptime = 85.5
	return probe
}

// SamplePausedProbe returns a paused probe.
func SamplePausedProbe() client.Probe {
	probe := SampleHTTPProbe()
	probe.ID = uuid.MustParse("eeeeeeee-1111-1111-1111-111111111111")
	probe.Name = "Maintenance Mode Service"
	probe.IsPaused = true
	probe.Status = "paused"
	return probe
}

// SampleProbeList returns a list of sample probes for pagination tests.
func SampleProbeList() []client.Probe {
	return []client.Probe{
		SampleHTTPProbe(),
		SamplePingProbe(),
		SampleTCPProbe(),
		SampleDNSProbe(),
	}
}

// SampleProbeListResponse returns a complete probe list API response.
func SampleProbeListResponse() client.ProbeListResponse {
	probes := SampleProbeList()
	return client.ProbeListResponse{
		Probes: probes,
		Total:  int64(len(probes)),
		Page:   1,
		Limit:  20,
	}
}

// ============================================================================
// Alert Fixtures
// ============================================================================

// SampleActiveAlert returns an active (untriggered) alert.
func SampleActiveAlert() client.Alert {
	msg := "Probe has been down for 5 minutes"
	return client.Alert{
		ID:               AlertID1,
		ProbeID:          ProbeID1,
		OrganizationID:   OrgID1,
		Status:           client.AlertStatusActive,
		Severity:         client.AlertSeverityCritical,
		AlertType:        client.AlertTypeStatusDown,
		Message:          &msg,
		TriggeredAt:      TimeOneHourAgo,
		NotifiedChannels: []string{"Slack #alerts", "Email: team@example.com"},
		CreatedAt:        TimeOneHourAgo,
		UpdatedAt:        TimeOneHourAgo,
		Probe: &client.AlertProbe{
			ID:        ProbeID1,
			Name:      "API Health Check",
			URL:       "https://api.example.com/health",
			CheckType: client.CheckTypeHTTP,
		},
	}
}

// SampleAcknowledgedAlert returns an acknowledged alert.
func SampleAcknowledgedAlert() client.Alert {
	msg := "SSL certificate expires in 7 days"
	ackTime := TimeOneHourAgo.Add(30 * time.Minute)
	return client.Alert{
		ID:               AlertID2,
		ProbeID:          ProbeID1,
		OrganizationID:   OrgID1,
		Status:           client.AlertStatusAcknowledged,
		Severity:         client.AlertSeverityWarning,
		AlertType:        client.AlertTypeSSLExpiry,
		Message:          &msg,
		TriggeredAt:      TimeOneHourAgo,
		AcknowledgedAt:   &ackTime,
		AcknowledgedBy:   &UserID1,
		NotifiedChannels: []string{"Email: security@example.com"},
		CreatedAt:        TimeOneHourAgo,
		UpdatedAt:        ackTime,
		Probe: &client.AlertProbe{
			ID:        ProbeID1,
			Name:      "API Health Check",
			URL:       "https://api.example.com/health",
			CheckType: client.CheckTypeHTTP,
		},
	}
}

// SampleResolvedAlert returns a resolved alert.
func SampleResolvedAlert() client.Alert {
	msg := "Response time exceeded threshold"
	ackTime := TimeOneDayAgo.Add(10 * time.Minute)
	resolveTime := TimeOneDayAgo.Add(45 * time.Minute)
	duration := 45 * 60 // 45 minutes in seconds
	return client.Alert{
		ID:               AlertID3,
		ProbeID:          ProbeID2,
		OrganizationID:   OrgID1,
		Status:           client.AlertStatusResolved,
		Severity:         client.AlertSeverityInfo,
		AlertType:        client.AlertTypeSlowResponse,
		Message:          &msg,
		TriggeredAt:      TimeOneDayAgo,
		AcknowledgedAt:   &ackTime,
		AcknowledgedBy:   &UserID2,
		ResolvedAt:       &resolveTime,
		DurationSeconds:  &duration,
		NotifiedChannels: []string{"Slack #alerts"},
		CreatedAt:        TimeOneDayAgo,
		UpdatedAt:        resolveTime,
		Probe: &client.AlertProbe{
			ID:        ProbeID2,
			Name:      "Database Server Ping",
			URL:       "db.internal.example.com",
			CheckType: client.CheckTypePing,
		},
	}
}

// SampleAlertList returns a list of sample alerts.
func SampleAlertList() []client.Alert {
	return []client.Alert{
		SampleActiveAlert(),
		SampleAcknowledgedAlert(),
		SampleResolvedAlert(),
	}
}

// SampleAlertListResponse returns a complete alert list API response.
func SampleAlertListResponse() client.AlertListResponse {
	alerts := SampleAlertList()
	return client.AlertListResponse{
		Alerts: alerts,
		Total:  int64(len(alerts)),
	}
}

// SampleAlertStats returns sample alert statistics.
func SampleAlertStats() client.AlertStats {
	mttr := int64(1800) // 30 minutes
	mtta := int64(300)  // 5 minutes
	return client.AlertStats{
		Period:             "24h",
		TotalAlerts:        15,
		ActiveAlerts:       3,
		AcknowledgedAlerts: 2,
		ResolvedAlerts:     10,
		CriticalAlerts:     5,
		WarningAlerts:      7,
		InfoAlerts:         3,
		MTTR:               &mttr,
		MTTA:               &mtta,
	}
}

// ============================================================================
// Channel Fixtures
// ============================================================================

// SampleEmailChannel returns a sample email notification channel.
func SampleEmailChannel() client.Channel {
	config, _ := json.Marshal(client.EmailChannelConfig{
		Address: "alerts@example.com",
	})
	return client.Channel{
		ID:             ChannelID1,
		OrganizationID: OrgID1,
		Name:           "Team Alerts Email",
		Type:           client.ChannelTypeEmail,
		Config:         config,
		Enabled:        true,
		ProbeCount:     5,
		CreatedAt:      TimeOneWeekAgo,
		UpdatedAt:      TimeOneDayAgo,
	}
}

// SampleSlackChannel returns a sample Slack notification channel.
func SampleSlackChannel() client.Channel {
	config, _ := json.Marshal(client.SlackChannelConfig{
		WebhookURL: "https://hooks.slack.com/services/test/test/test",
	})
	return client.Channel{
		ID:             ChannelID2,
		OrganizationID: OrgID1,
		Name:           "Slack #alerts",
		Type:           client.ChannelTypeSlack,
		Config:         config,
		Enabled:        true,
		ProbeCount:     10,
		CreatedAt:      TimeOneWeekAgo,
		UpdatedAt:      TimeOneDayAgo,
	}
}

// SampleWebhookChannel returns a sample webhook notification channel.
func SampleWebhookChannel() client.Channel {
	config, _ := json.Marshal(client.WebhookChannelConfig{
		URL:    "https://api.example.com/webhooks/alerts",
		Method: "POST",
		Headers: map[string]string{
			"Authorization": "Bearer token123",
			"Content-Type":  "application/json",
		},
	})
	return client.Channel{
		ID:             ChannelID3,
		OrganizationID: OrgID1,
		Name:           "Custom Webhook",
		Type:           client.ChannelTypeWebhook,
		Config:         config,
		Enabled:        true,
		ProbeCount:     2,
		CreatedAt:      TimeOneWeekAgo,
		UpdatedAt:      TimeOneDayAgo,
	}
}

// SamplePagerDutyChannel returns a sample PagerDuty notification channel.
func SamplePagerDutyChannel() client.Channel {
	config, _ := json.Marshal(client.PagerDutyChannelConfig{
		RoutingKey: "R0123456789ABCDEF0123456789ABCDEF",
		Severity:   "critical",
	})
	return client.Channel{
		ID:             ChannelID4,
		OrganizationID: OrgID1,
		Name:           "PagerDuty On-Call",
		Type:           client.ChannelTypePagerDuty,
		Config:         config,
		Enabled:        true,
		ProbeCount:     3,
		CreatedAt:      TimeOneWeekAgo,
		UpdatedAt:      TimeOneDayAgo,
	}
}

// SampleDisabledChannel returns a disabled channel.
func SampleDisabledChannel() client.Channel {
	ch := SampleEmailChannel()
	ch.ID = uuid.MustParse("c5555555-5555-5555-5555-555555555555")
	ch.Name = "Old Email (disabled)"
	ch.Enabled = false
	ch.ProbeCount = 0
	return ch
}

// SampleChannelList returns a list of sample channels.
func SampleChannelList() []client.Channel {
	return []client.Channel{
		SampleEmailChannel(),
		SampleSlackChannel(),
		SampleWebhookChannel(),
		SamplePagerDutyChannel(),
	}
}

// SampleChannelListResponse returns a complete channel list API response.
func SampleChannelListResponse() client.ChannelListResponse {
	channels := SampleChannelList()
	return client.ChannelListResponse{
		Channels: channels,
		Total:    int64(len(channels)),
	}
}

// ============================================================================
// Team Fixtures
// ============================================================================

// SampleOwnerMember returns a sample team owner.
func SampleOwnerMember() client.TeamMember {
	return client.TeamMember{
		ID:        1,
		UserID:    UserID1.String(),
		Email:     "owner@example.com",
		Name:      "Alice Owner",
		Role:      client.TeamRoleOwner,
		JoinedAt:  TimeOneWeekAgo.Add(-30 * 24 * time.Hour), // 30 days before base
		AvatarURL: "https://avatars.example.com/alice.png",
	}
}

// SampleAdminMember returns a sample team admin.
func SampleAdminMember() client.TeamMember {
	return client.TeamMember{
		ID:        2,
		UserID:    UserID2.String(),
		Email:     "admin@example.com",
		Name:      "Bob Admin",
		Role:      client.TeamRoleAdmin,
		JoinedAt:  TimeOneWeekAgo,
		AvatarURL: "https://avatars.example.com/bob.png",
	}
}

// SampleRegularMember returns a sample regular team member.
func SampleRegularMember() client.TeamMember {
	return client.TeamMember{
		ID:       3,
		UserID:   "33333333-3333-3333-3333-333333333333",
		Email:    "member@example.com",
		Name:     "Charlie Member",
		Role:     client.TeamRoleMember,
		JoinedAt: TimeOneDayAgo,
	}
}

// SampleViewerMember returns a sample viewer team member.
func SampleViewerMember() client.TeamMember {
	return client.TeamMember{
		ID:       4,
		UserID:   "44444444-4444-4444-4444-444444444444",
		Email:    "viewer@example.com",
		Name:     "Dana Viewer",
		Role:     client.TeamRoleViewer,
		JoinedAt: BaseTime,
	}
}

// SampleTeamMemberList returns a list of sample team members.
func SampleTeamMemberList() []client.TeamMember {
	return []client.TeamMember{
		SampleOwnerMember(),
		SampleAdminMember(),
		SampleRegularMember(),
		SampleViewerMember(),
	}
}

// SampleTeamMemberListResponse returns a complete team member list API response.
func SampleTeamMemberListResponse() client.TeamMemberListResponse {
	members := SampleTeamMemberList()
	return client.TeamMemberListResponse{
		Members: members,
		Total:   int64(len(members)),
	}
}

// SampleInvitation returns a sample pending invitation.
func SampleInvitation() client.Invitation {
	return client.Invitation{
		ID:             "inv-001",
		Email:          "newuser@example.com",
		Role:           client.TeamRoleMember,
		InviteCode:     "ABC123",
		ExpiresAt:      BaseTime.Add(7 * 24 * time.Hour),
		InvitedBy:      UserID1.String(),
		OrganizationID: OrgID1.String(),
	}
}

// SampleInvitationList returns a list of sample invitations.
func SampleInvitationList() []client.Invitation {
	return []client.Invitation{
		SampleInvitation(),
		{
			ID:             "inv-002",
			Email:          "another@example.com",
			Role:           client.TeamRoleViewer,
			InviteCode:     "XYZ789",
			ExpiresAt:      BaseTime.Add(3 * 24 * time.Hour),
			InvitedBy:      UserID2.String(),
			OrganizationID: OrgID1.String(),
		},
	}
}

// ============================================================================
// Region Fixtures
// ============================================================================

// SampleRegions returns regions grouped by continent.
func SampleRegions() map[string][]client.Region {
	return map[string][]client.Region{
		"north_america": {
			{ID: "nyc3", Name: "New York 3", DisplayName: "New York", CountryCode: "US"},
			{ID: "sfo3", Name: "San Francisco 3", DisplayName: "San Francisco", CountryCode: "US"},
			{ID: "tor1", Name: "Toronto 1", DisplayName: "Toronto", CountryCode: "CA"},
		},
		"europe": {
			{ID: "lon1", Name: "London 1", DisplayName: "London", CountryCode: "GB"},
			{ID: "fra1", Name: "Frankfurt 1", DisplayName: "Frankfurt", CountryCode: "DE"},
			{ID: "ams3", Name: "Amsterdam 3", DisplayName: "Amsterdam", CountryCode: "NL"},
		},
		"asia_pacific": {
			{ID: "sgp1", Name: "Singapore 1", DisplayName: "Singapore", CountryCode: "SG"},
			{ID: "blr1", Name: "Bangalore 1", DisplayName: "Bangalore", CountryCode: "IN"},
			{ID: "syd1", Name: "Sydney 1", DisplayName: "Sydney", CountryCode: "AU"},
		},
	}
}

// SampleRegionListResponse returns a complete region list API response.
func SampleRegionListResponse() client.RegionListResponse {
	return client.RegionListResponse{
		Status: "success",
		Data:   SampleRegions(),
	}
}

// SampleRegionStatus returns a sample region status.
func SampleRegionStatus() client.RegionStatus {
	return client.RegionStatus{
		ID:           "nyc3",
		Name:         "New York 3",
		Status:       "active",
		HealthStatus: "healthy",
	}
}

// SampleRegionStatusMaintenance returns a region in maintenance.
func SampleRegionStatusMaintenance() client.RegionStatus {
	reason := "Scheduled hardware upgrade"
	ends := BaseTime.Add(4 * time.Hour)
	return client.RegionStatus{
		ID:                "fra1",
		Name:              "Frankfurt 1",
		Status:            "maintenance",
		HealthStatus:      "warning",
		MaintenanceReason: &reason,
		MaintenanceEndsAt: &ends,
	}
}

// ============================================================================
// Organization Fixtures
// ============================================================================

// SampleOrganization returns a sample organization.
func SampleOrganization() client.Organization {
	return client.Organization{
		ID:        OrgID1.String(),
		Name:      "Acme Corp",
		Slug:      "acme-corp",
		Role:      "owner",
		IsCurrent: true,
	}
}

// SampleSecondOrganization returns a second sample organization.
func SampleSecondOrganization() client.Organization {
	return client.Organization{
		ID:        OrgID2.String(),
		Name:      "Beta Inc",
		Slug:      "beta-inc",
		Role:      "member",
		IsCurrent: false,
	}
}

// SampleOrganizationList returns a list of sample organizations.
func SampleOrganizationList() []client.Organization {
	return []client.Organization{
		SampleOrganization(),
		SampleSecondOrganization(),
	}
}

// SampleListOrganizationsResponse returns a complete organizations list API response.
func SampleListOrganizationsResponse() client.ListOrganizationsResponse {
	orgs := SampleOrganizationList()
	return client.ListOrganizationsResponse{
		Organizations: orgs,
		Total:         len(orgs),
	}
}
