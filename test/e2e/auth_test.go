package e2e

import (
	"testing"

	"github.com/StackEye-IO/stackeye-cli/test/e2e/testutil"
	"github.com/StackEye-IO/stackeye-go-sdk/config"
	"github.com/stretchr/testify/require"
)

// TestWhoamiAuthenticated tests the whoami command with valid credentials.
func TestWhoamiAuthenticated(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	// Register extended routes including /v1/cli-auth/verify
	env.Server.RegisterAllRoutes()

	t.Run("default output shows user info", func(t *testing.T) {
		result := env.RunSuccess("whoami")

		env.AssertContains(result, "admin@example.com")
		env.AssertContains(result, "Admin User")
		env.AssertContains(result, "Test Organization")
		env.AssertContains(result, "api_key")
		env.AssertAPICall("GET", "/v1/cli-auth/verify")
	})

	t.Run("shows context name", func(t *testing.T) {
		env.ClearAPICalls()

		result := env.RunSuccess("whoami")

		env.AssertContains(result, "test")
		env.AssertAPICall("GET", "/v1/cli-auth/verify")
	})
}

// TestWhoamiUnauthenticated tests whoami when no API key is configured.
func TestWhoamiUnauthenticated(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	// Clear the API key from the test config
	ctx, err := env.Config.Config.GetCurrentContext()
	require.NoError(t, err)
	ctx.APIKey = ""
	err = env.Config.Config.SaveTo(env.Config.ConfigPath)
	require.NoError(t, err)

	t.Run("shows not logged in", func(t *testing.T) {
		result := env.RunSuccess("whoami")

		env.AssertContains(result, "Not logged in")
	})
}

// TestWhoamiInvalidCredentials tests whoami when the server rejects the API key.
func TestWhoamiInvalidCredentials(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	// Register extended routes so the verify endpoint exists
	env.Server.RegisterAllRoutes()

	// Force the verify endpoint to return 401
	env.Server.WithError("GET", "/v1/cli-auth/verify", 401, "invalid API key")

	t.Run("returns error on invalid credentials", func(t *testing.T) {
		result := env.RunError("whoami")

		combined := result.CombinedOutput()
		require.Contains(t, combined, "failed to verify credentials")
	})
}

// TestLogoutCurrentContext tests logging out from the current context.
func TestLogoutCurrentContext(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	// Register extended routes for whoami verification
	env.Server.RegisterAllRoutes()

	t.Run("logout clears credentials for current context", func(t *testing.T) {
		// Verify we're logged in first
		result := env.RunSuccess("whoami")
		env.AssertContains(result, "admin@example.com")

		// Logout
		result = env.RunSuccess("logout")
		env.AssertContains(result, "Logged out")
		env.AssertContains(result, "test")

		// Verify credentials are cleared - whoami should show not logged in
		result = env.RunSuccess("whoami")
		env.AssertContains(result, "Not logged in")
	})
}

// TestLogoutAll tests logging out from all contexts.
func TestLogoutAll(t *testing.T) {
	env := testutil.NewTestEnvWithMultipleContexts(t)
	defer env.Cleanup()

	t.Run("logout --all clears all context credentials", func(t *testing.T) {
		result := env.RunSuccess("logout", "--all")
		env.AssertContains(result, "Logged out")

		// Reload config and verify all API keys are cleared
		err := env.Config.Reload()
		require.NoError(t, err)

		for name, ctx := range env.Config.Config.Contexts {
			require.Empty(t, ctx.APIKey, "expected API key to be cleared for context %q", name)
		}
	})
}

// TestLogoutAlreadyLoggedOut tests logout when already logged out.
func TestLogoutAlreadyLoggedOut(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	// Clear the API key first
	ctx, err := env.Config.Config.GetCurrentContext()
	require.NoError(t, err)
	ctx.APIKey = ""
	err = env.Config.Config.SaveTo(env.Config.ConfigPath)
	require.NoError(t, err)

	t.Run("shows already logged out", func(t *testing.T) {
		result := env.RunSuccess("logout")
		env.AssertContains(result, "Already logged out")
	})
}

// TestLogoutPreservesContextConfig tests that logout preserves context configuration.
func TestLogoutPreservesContextConfig(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("logout preserves API URL and org name", func(t *testing.T) {
		// Logout
		env.RunSuccess("logout")

		// Reload config and check context still exists with config preserved
		err := env.Config.Reload()
		require.NoError(t, err)

		ctx, err := env.Config.Config.GetContext("test")
		require.NoError(t, err)
		require.Empty(t, ctx.APIKey, "API key should be cleared")
		require.NotEmpty(t, ctx.APIURL, "API URL should be preserved")
		require.NotEmpty(t, ctx.OrganizationName, "Organization name should be preserved")
	})
}

// TestUnauthenticatedAPIAccess tests that CLI commands give clear errors without auth.
func TestUnauthenticatedAPIAccess(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	// Remove the API key to simulate unauthenticated state
	ctx, err := env.Config.Config.GetCurrentContext()
	require.NoError(t, err)
	ctx.APIKey = ""
	err = env.Config.Config.SaveTo(env.Config.ConfigPath)
	require.NoError(t, err)

	t.Run("probe list without auth shows error", func(t *testing.T) {
		result := env.RunError("probe", "list")

		combined := result.CombinedOutput()
		// Should indicate authentication is needed (various possible error messages)
		require.True(t,
			result.Contains("login") || result.StderrContains("login") ||
				result.Contains("not logged in") || result.StderrContains("not logged in") ||
				result.Contains("No API key") || result.StderrContains("No API key") ||
				result.Contains("no API key") || result.StderrContains("no API key") ||
				result.Contains("authenticate") || result.StderrContains("authenticate"),
			"expected auth-related error message, got: %s", combined)
	})
}

// TestWhoamiNoConfig tests whoami when no config file exists at all.
func TestWhoamiNoConfig(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	// Create a new config with empty contexts
	emptyCfg := config.NewConfig()
	err := emptyCfg.SaveTo(env.Config.ConfigPath)
	require.NoError(t, err)

	t.Run("shows not logged in with no context", func(t *testing.T) {
		result := env.RunSuccess("whoami")
		env.AssertContains(result, "Not logged in")
	})
}
