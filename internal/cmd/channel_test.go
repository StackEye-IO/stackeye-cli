package cmd

import (
	"testing"
)

func TestNewChannelCmd(t *testing.T) {
	cmd := NewChannelCmd()

	t.Run("command configuration", func(t *testing.T) {
		if cmd.Use != "channel" {
			t.Errorf("Use = %q, want %q", cmd.Use, "channel")
		}

		if cmd.Short == "" {
			t.Error("Short description should not be empty")
		}

		if cmd.Long == "" {
			t.Error("Long description should not be empty")
		}
	})

	t.Run("aliases", func(t *testing.T) {
		expectedAliases := []string{"channels", "ch"}
		if len(cmd.Aliases) != len(expectedAliases) {
			t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
			return
		}

		for i, alias := range expectedAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Alias[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	})

	t.Run("help content includes channel types", func(t *testing.T) {
		channelTypes := []string{"email", "slack", "webhook", "pagerduty", "discord", "teams", "sms"}

		for _, ct := range channelTypes {
			if !containsString(cmd.Long, ct) {
				t.Errorf("Long description should mention channel type %q", ct)
			}
		}
	})

	t.Run("help content includes available commands", func(t *testing.T) {
		// Only check for commands that are currently implemented
		commands := []string{"list", "get", "create", "update"}

		for _, c := range commands {
			if !containsString(cmd.Long, c) {
				t.Errorf("Long description should mention command %q", c)
			}
		}
	})
}

// containsString checks if s contains substr.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

// containsSubstring is a simple substring check.
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
