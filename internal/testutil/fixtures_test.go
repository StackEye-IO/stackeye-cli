package testutil

import (
	"encoding/json"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

func TestProbeFixtures(t *testing.T) {
	t.Run("SampleHTTPProbe returns valid probe", func(t *testing.T) {
		probe := SampleHTTPProbe()
		if probe.ID != ProbeID1 {
			t.Errorf("expected ProbeID1, got %s", probe.ID)
		}
		if probe.CheckType != client.CheckTypeHTTP {
			t.Errorf("expected http check type, got %s", probe.CheckType)
		}
		if probe.Name == "" {
			t.Error("expected non-empty name")
		}
	})

	t.Run("SampleProbeList returns multiple probes", func(t *testing.T) {
		probes := SampleProbeList()
		if len(probes) < 4 {
			t.Errorf("expected at least 4 probes, got %d", len(probes))
		}
	})

	t.Run("SampleProbeListResponse is valid", func(t *testing.T) {
		resp := SampleProbeListResponse()
		if resp.Total != int64(len(resp.Probes)) {
			t.Errorf("total (%d) doesn't match probes length (%d)", resp.Total, len(resp.Probes))
		}
	})
}

func TestAlertFixtures(t *testing.T) {
	t.Run("SampleActiveAlert returns active alert", func(t *testing.T) {
		alert := SampleActiveAlert()
		if alert.Status != client.AlertStatusActive {
			t.Errorf("expected active status, got %s", alert.Status)
		}
	})

	t.Run("SampleAcknowledgedAlert has acknowledgement fields", func(t *testing.T) {
		alert := SampleAcknowledgedAlert()
		if alert.AcknowledgedAt == nil {
			t.Error("expected acknowledged_at to be set")
		}
		if alert.AcknowledgedBy == nil {
			t.Error("expected acknowledged_by to be set")
		}
	})

	t.Run("SampleResolvedAlert has resolution fields", func(t *testing.T) {
		alert := SampleResolvedAlert()
		if alert.ResolvedAt == nil {
			t.Error("expected resolved_at to be set")
		}
		if alert.DurationSeconds == nil {
			t.Error("expected duration_seconds to be set")
		}
	})
}

func TestChannelFixtures(t *testing.T) {
	t.Run("SampleEmailChannel has valid config", func(t *testing.T) {
		ch := SampleEmailChannel()
		if ch.Type != client.ChannelTypeEmail {
			t.Errorf("expected email type, got %s", ch.Type)
		}

		var config client.EmailChannelConfig
		if err := json.Unmarshal(ch.Config, &config); err != nil {
			t.Errorf("failed to unmarshal email config: %v", err)
		}
		if config.Address == "" {
			t.Error("expected non-empty address")
		}
	})

	t.Run("SampleSlackChannel has valid config", func(t *testing.T) {
		ch := SampleSlackChannel()
		if ch.Type != client.ChannelTypeSlack {
			t.Errorf("expected slack type, got %s", ch.Type)
		}

		var config client.SlackChannelConfig
		if err := json.Unmarshal(ch.Config, &config); err != nil {
			t.Errorf("failed to unmarshal slack config: %v", err)
		}
		if config.WebhookURL == "" {
			t.Error("expected non-empty webhook URL")
		}
	})

	t.Run("SampleChannelList returns multiple channels", func(t *testing.T) {
		channels := SampleChannelList()
		if len(channels) < 4 {
			t.Errorf("expected at least 4 channels, got %d", len(channels))
		}
	})
}

func TestTeamFixtures(t *testing.T) {
	t.Run("SampleOwnerMember has owner role", func(t *testing.T) {
		member := SampleOwnerMember()
		if member.Role != client.TeamRoleOwner {
			t.Errorf("expected owner role, got %s", member.Role)
		}
	})

	t.Run("SampleTeamMemberList has all roles", func(t *testing.T) {
		members := SampleTeamMemberList()
		roles := make(map[client.TeamRole]bool)
		for _, m := range members {
			roles[m.Role] = true
		}
		if !roles[client.TeamRoleOwner] {
			t.Error("expected owner role in list")
		}
		if !roles[client.TeamRoleAdmin] {
			t.Error("expected admin role in list")
		}
		if !roles[client.TeamRoleMember] {
			t.Error("expected member role in list")
		}
		if !roles[client.TeamRoleViewer] {
			t.Error("expected viewer role in list")
		}
	})
}

func TestRegionFixtures(t *testing.T) {
	t.Run("SampleRegions has all continents", func(t *testing.T) {
		regions := SampleRegions()
		if _, ok := regions["north_america"]; !ok {
			t.Error("expected north_america continent")
		}
		if _, ok := regions["europe"]; !ok {
			t.Error("expected europe continent")
		}
		if _, ok := regions["asia_pacific"]; !ok {
			t.Error("expected asia_pacific continent")
		}
	})

	t.Run("SampleRegionStatus is healthy", func(t *testing.T) {
		status := SampleRegionStatus()
		if status.Status != "active" {
			t.Errorf("expected active status, got %s", status.Status)
		}
		if status.HealthStatus != "healthy" {
			t.Errorf("expected healthy, got %s", status.HealthStatus)
		}
	})
}

func TestOrganizationFixtures(t *testing.T) {
	t.Run("SampleOrganization is current", func(t *testing.T) {
		org := SampleOrganization()
		if !org.IsCurrent {
			t.Error("expected is_current to be true")
		}
		if org.Role != "owner" {
			t.Errorf("expected owner role, got %s", org.Role)
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("StringPtr returns pointer", func(t *testing.T) {
		s := "test"
		ptr := StringPtr(s)
		if *ptr != s {
			t.Errorf("expected %q, got %q", s, *ptr)
		}
	})

	t.Run("IntPtr returns pointer", func(t *testing.T) {
		i := 42
		ptr := IntPtr(i)
		if *ptr != i {
			t.Errorf("expected %d, got %d", i, *ptr)
		}
	})

	t.Run("BoolPtr returns pointer", func(t *testing.T) {
		b := true
		ptr := BoolPtr(b)
		if *ptr != b {
			t.Errorf("expected %v, got %v", b, *ptr)
		}
	})

	t.Run("DaysAgo returns correct time", func(t *testing.T) {
		oneDay := DaysAgo(1)
		expected := BaseTime.AddDate(0, 0, -1)
		if !oneDay.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, oneDay)
		}
	})

	t.Run("HoursAgo returns correct time", func(t *testing.T) {
		twoHours := HoursAgo(2)
		if !twoHours.Before(BaseTime) {
			t.Error("expected time to be before BaseTime")
		}
	})
}
