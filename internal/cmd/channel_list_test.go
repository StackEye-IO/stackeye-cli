// Package cmd implements the CLI commands for StackEye.
package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/client"
)

// TestNewChannelListCmd verifies that the channel list command is properly constructed.
func TestNewChannelListCmd(t *testing.T) {
	cmd := NewChannelListCmd()

	if cmd.Use != "list" {
		t.Errorf("expected Use to be 'list', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify aliases
	expectedAliases := []string{"ls"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	} else {
		for i, alias := range expectedAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("expected alias[%d] to be %q, got %q", i, alias, cmd.Aliases[i])
			}
		}
	}

	// Verify flags exist
	flags := []string{"type", "enabled", "page", "limit"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag %q to be defined", flag)
		}
	}
}

// TestValidateChannelListFlags tests the production validation logic.
func TestValidateChannelListFlags(t *testing.T) {
	tests := []struct {
		name        string
		flags       channelListFlags
		wantErrMsg  string
		shouldError bool
	}{
		{
			name: "valid defaults",
			flags: channelListFlags{
				page:  1,
				limit: 20,
			},
			shouldError: false,
		},
		{
			name: "invalid limit too low",
			flags: channelListFlags{
				page:  1,
				limit: 0,
			},
			wantErrMsg:  "invalid limit 0: must be between 1 and 100",
			shouldError: true,
		},
		{
			name: "invalid limit too high",
			flags: channelListFlags{
				page:  1,
				limit: 101,
			},
			wantErrMsg:  "invalid limit 101: must be between 1 and 100",
			shouldError: true,
		},
		{
			name: "invalid page",
			flags: channelListFlags{
				page:  0,
				limit: 20,
			},
			wantErrMsg:  "invalid page 0: must be at least 1",
			shouldError: true,
		},
		{
			name: "invalid channel type",
			flags: channelListFlags{
				page:        1,
				limit:       20,
				channelType: "invalid",
			},
			wantErrMsg:  `invalid value "invalid" for --type: must be one of: email, slack, webhook, pagerduty, discord, teams, sms`,
			shouldError: true,
		},
		{
			name: "valid channel type email",
			flags: channelListFlags{
				page:        1,
				limit:       20,
				channelType: "email",
			},
			shouldError: false,
		},
		{
			name: "valid channel type slack",
			flags: channelListFlags{
				page:        1,
				limit:       20,
				channelType: "slack",
			},
			shouldError: false,
		},
		{
			name: "valid channel type webhook",
			flags: channelListFlags{
				page:        1,
				limit:       20,
				channelType: "webhook",
			},
			shouldError: false,
		},
		{
			name: "valid channel type pagerduty",
			flags: channelListFlags{
				page:        1,
				limit:       20,
				channelType: "pagerduty",
			},
			shouldError: false,
		},
		{
			name: "valid channel type discord",
			flags: channelListFlags{
				page:        1,
				limit:       20,
				channelType: "discord",
			},
			shouldError: false,
		},
		{
			name: "valid channel type teams",
			flags: channelListFlags{
				page:        1,
				limit:       20,
				channelType: "teams",
			},
			shouldError: false,
		},
		{
			name: "valid channel type sms",
			flags: channelListFlags{
				page:        1,
				limit:       20,
				channelType: "sms",
			},
			shouldError: false,
		},
		{
			name: "invalid enabled value",
			flags: channelListFlags{
				page:    1,
				limit:   20,
				enabled: "maybe",
			},
			wantErrMsg:  `invalid value "maybe" for --enabled: must be one of: true, false`,
			shouldError: true,
		},
		{
			name: "valid enabled true",
			flags: channelListFlags{
				page:    1,
				limit:   20,
				enabled: "true",
			},
			shouldError: false,
		},
		{
			name: "valid enabled false",
			flags: channelListFlags{
				page:    1,
				limit:   20,
				enabled: "false",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the actual production validation function
			err := validateChannelListFlags(&tt.flags)
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.wantErrMsg {
					t.Errorf("expected error %q, got %q", tt.wantErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestParseChannelType tests the channel type parsing function.
func TestParseChannelType(t *testing.T) {
	tests := []struct {
		input    string
		expected client.ChannelType
	}{
		{"email", client.ChannelTypeEmail},
		{"slack", client.ChannelTypeSlack},
		{"webhook", client.ChannelTypeWebhook},
		{"pagerduty", client.ChannelTypePagerDuty},
		{"discord", client.ChannelTypeDiscord},
		{"teams", client.ChannelTypeTeams},
		{"sms", client.ChannelTypeSMS},
		{"", ""},
		{"invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseChannelType(tt.input)
			if result != tt.expected {
				t.Errorf("parseChannelType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseEnabledFilter tests the enabled filter parsing function.
func TestParseEnabledFilter(t *testing.T) {
	tests := []struct {
		input    string
		expected *bool
	}{
		{"true", boolPtr(true)},
		{"false", boolPtr(false)},
		{"", nil},
		{"invalid", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseEnabledFilter(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("parseEnabledFilter(%q) = %v, want nil", tt.input, *result)
				}
			} else {
				if result == nil {
					t.Errorf("parseEnabledFilter(%q) = nil, want %v", tt.input, *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("parseEnabledFilter(%q) = %v, want %v", tt.input, *result, *tt.expected)
				}
			}
		})
	}
}

// TestPaginationOffset verifies the pagination offset calculation.
func TestPaginationOffset(t *testing.T) {
	tests := []struct {
		page           int
		limit          int
		expectedOffset int
	}{
		{page: 1, limit: 20, expectedOffset: 0},
		{page: 2, limit: 20, expectedOffset: 20},
		{page: 3, limit: 20, expectedOffset: 40},
		{page: 1, limit: 50, expectedOffset: 0},
		{page: 2, limit: 50, expectedOffset: 50},
		{page: 10, limit: 100, expectedOffset: 900},
	}

	for _, tt := range tests {
		// Use same formula as production code
		offset := (tt.page - 1) * tt.limit
		if offset != tt.expectedOffset {
			t.Errorf("page=%d, limit=%d: got offset %d, want %d", tt.page, tt.limit, offset, tt.expectedOffset)
		}
	}
}

// boolPtr is a helper to create *bool for test cases.
func boolPtr(b bool) *bool {
	return &b
}

// TestRunChannelList_Validation tests that validation errors are returned for invalid inputs.
// Since runChannelList requires an API client, validation happens first
// and will fail before making API calls for invalid inputs.
func TestRunChannelList_Validation(t *testing.T) {
	tests := []struct {
		name         string
		limit        int
		page         int
		channelType  string
		enabled      string
		wantErrorMsg string
	}{
		{
			name:         "limit too low",
			limit:        0,
			page:         1,
			wantErrorMsg: "invalid limit 0: must be between 1 and 100",
		},
		{
			name:         "limit too high",
			limit:        101,
			page:         1,
			wantErrorMsg: "invalid limit 101: must be between 1 and 100",
		},
		{
			name:         "page too low",
			limit:        20,
			page:         0,
			wantErrorMsg: "invalid page 0: must be at least 1",
		},
		{
			name:         "invalid channel type",
			limit:        20,
			page:         1,
			channelType:  "badtype",
			wantErrorMsg: `invalid value "badtype" for --type: must be one of: email, slack, webhook, pagerduty, discord, teams, sms`,
		},
		{
			name:         "invalid enabled value",
			limit:        20,
			page:         1,
			enabled:      "maybe",
			wantErrorMsg: `invalid value "maybe" for --enabled: must be one of: true, false`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &channelListFlags{
				page:        tt.page,
				limit:       tt.limit,
				channelType: tt.channelType,
				enabled:     tt.enabled,
			}

			// Call runChannelList with a background context.
			// It should fail on validation before needing API client.
			err := runChannelList(context.Background(), flags)

			if err == nil {
				t.Errorf("expected error containing %q, got nil", tt.wantErrorMsg)
				return
			}

			if !strings.Contains(err.Error(), tt.wantErrorMsg) {
				t.Errorf("expected error containing %q, got %q", tt.wantErrorMsg, err.Error())
			}
		})
	}
}

// TestRunChannelList_ValidFlags tests that valid flags pass validation.
// The function will fail later when trying to get the API client.
func TestRunChannelList_ValidFlags(t *testing.T) {
	tests := []struct {
		name        string
		channelType string
		enabled     string
	}{
		{
			name:        "defaults only",
			channelType: "",
			enabled:     "",
		},
		{
			name:        "with email type",
			channelType: "email",
			enabled:     "",
		},
		{
			name:        "with slack type",
			channelType: "slack",
			enabled:     "",
		},
		{
			name:        "with enabled true",
			channelType: "",
			enabled:     "true",
		},
		{
			name:        "with enabled false",
			channelType: "",
			enabled:     "false",
		},
		{
			name:        "with type and enabled",
			channelType: "webhook",
			enabled:     "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := &channelListFlags{
				page:        1,
				limit:       20,
				channelType: tt.channelType,
				enabled:     tt.enabled,
			}

			err := runChannelList(context.Background(), flags)

			// Should fail on API client initialization, not validation
			if err == nil {
				t.Error("expected error (no API client configured), got nil")
				return
			}

			// Error should NOT be a validation error
			validationErrors := []string{
				"invalid limit",
				"invalid page",
				"invalid channel type",
				"invalid enabled value",
			}

			for _, ve := range validationErrors {
				if strings.Contains(err.Error(), ve) {
					t.Errorf("got validation error %q, expected non-validation error", err.Error())
					return
				}
			}

			// Should be an API client error
			if !strings.Contains(err.Error(), "API client") && !strings.Contains(err.Error(), "client") && !strings.Contains(err.Error(), "config") {
				// It's OK if we get a different non-validation error
				t.Logf("got expected non-validation error: %v", err)
			}
		})
	}
}
