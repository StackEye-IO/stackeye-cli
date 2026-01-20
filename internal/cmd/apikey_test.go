package cmd

import (
	"strings"
	"testing"
)

func TestNewAPIKeyCmd(t *testing.T) {
	cmd := NewAPIKeyCmd()

	t.Run("command properties", func(t *testing.T) {
		if cmd.Use != "api-key" {
			t.Errorf("Use = %q, want %q", cmd.Use, "api-key")
		}

		if cmd.Short == "" {
			t.Error("Short description should not be empty")
		}

		if cmd.Long == "" {
			t.Error("Long description should not be empty")
		}
	})

	t.Run("aliases", func(t *testing.T) {
		expectedAliases := []string{"apikey", "apikeys", "api-keys"}
		if len(cmd.Aliases) != len(expectedAliases) {
			t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
			return
		}

		for i, alias := range expectedAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	})

	t.Run("long description content", func(t *testing.T) {
		// Verify key concepts are documented
		longDesc := cmd.Long

		mustContain := []string{
			"API keys",
			"programmatic access",
			"organization",
			"se_",
			"Security Best Practices",
		}

		for _, phrase := range mustContain {
			if !strings.Contains(longDesc, phrase) {
				t.Errorf("Long description should contain %q", phrase)
			}
		}
	})
}
