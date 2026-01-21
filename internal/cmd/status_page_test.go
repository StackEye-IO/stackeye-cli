package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusPageCmd(t *testing.T) {
	cmd := NewStatusPageCmd()

	require.NotNil(t, cmd, "NewStatusPageCmd should return a non-nil command")

	// Verify command configuration
	assert.Equal(t, "status-page", cmd.Use, "Use should be 'status-page'")
	assert.Equal(t, "Manage public status pages", cmd.Short, "Short description should match")
	assert.Contains(t, cmd.Long, "Status pages provide public-facing displays", "Long description should describe status pages")

	// Verify aliases
	expectedAliases := []string{"sp", "statuspage"}
	assert.Equal(t, expectedAliases, cmd.Aliases, "Aliases should include 'sp' and 'statuspage'")
}

func TestStatusPageCmd_HasNoRunFunction(t *testing.T) {
	cmd := NewStatusPageCmd()

	// Parent command should not have a Run function (subcommands do)
	assert.Nil(t, cmd.Run, "Parent command should not have Run function")
	assert.Nil(t, cmd.RunE, "Parent command should not have RunE function")
}

func TestStatusPageCmd_LongDescription(t *testing.T) {
	cmd := NewStatusPageCmd()

	// Verify key information is in the long description
	assert.Contains(t, cmd.Long, "Key Features:", "Should document key features")
	assert.Contains(t, cmd.Long, "Plan Limits:", "Should document plan limits")
	assert.Contains(t, cmd.Long, "Examples:", "Should include examples")
	assert.Contains(t, cmd.Long, "stackeye status-page list", "Should show list example")
	assert.Contains(t, cmd.Long, "stackeye status-page create", "Should show create example")
	assert.Contains(t, cmd.Long, "stackeye status-page get", "Should show get example")
	assert.Contains(t, cmd.Long, "stackeye status-page delete", "Should show delete example")
	assert.Contains(t, cmd.Long, "stackeye status-page add-probe", "Should show add-probe example")
	assert.Contains(t, cmd.Long, "stackeye status-page domain-verify", "Should show domain-verify example")
}
