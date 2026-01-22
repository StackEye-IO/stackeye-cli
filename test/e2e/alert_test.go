package e2e

import (
	"testing"

	"github.com/StackEye-IO/stackeye-cli/test/e2e/testutil"
)

// TestAlertList tests the alert list command.
func TestAlertList(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("default table output", func(t *testing.T) {
		result := env.RunSuccess("alert", "list")

		// Should contain fixture alert data
		// Note: Table output uses uppercase for severity
		env.AssertContains(result, "API Health Check")
		env.AssertContains(result, "CRITICAL")
		env.AssertAPICall("GET", "/v1/alerts")
	})

	t.Run("json output format", func(t *testing.T) {
		env.ClearAPICalls()
		result := env.RunSuccess("alert", "list", "--output", "json")

		var data any
		env.AssertJSONOutput(result, &data)
		env.AssertAPICall("GET", "/v1/alerts")
	})

	t.Run("yaml output format", func(t *testing.T) {
		env.ClearAPICalls()
		result := env.RunSuccess("alert", "list", "--output", "yaml")

		env.AssertContains(result, "severity:")
		env.AssertAPICall("GET", "/v1/alerts")
	})
}

// TestAlertGet tests the alert get command.
func TestAlertGet(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("get existing alert", func(t *testing.T) {
		result := env.RunSuccess("alert", "get", testutil.AlertID1)

		// Verify the alert ID is in output (table output shows AlertGetResponse wrapper)
		env.AssertContains(result, testutil.AlertID1)
		env.AssertAPICall("GET", "/v1/alerts/"+testutil.AlertID1)
	})

	t.Run("get non-existent alert", func(t *testing.T) {
		result := env.RunError("alert", "get", "99999999-9999-9999-9999-999999999999")

		env.AssertStderrContains(result, "not found")
	})

	t.Run("json output format", func(t *testing.T) {
		env.ClearAPICalls()
		result := env.RunSuccess("alert", "get", testutil.AlertID1, "--output", "json")

		var data any
		env.AssertJSONOutput(result, &data)
	})
}

// TestAlertAcknowledge tests the alert acknowledge command.
func TestAlertAcknowledge(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("acknowledge existing alert", func(t *testing.T) {
		result := env.RunSuccess("alert", "ack", testutil.AlertID1)

		env.AssertContains(result, "acknowledged")
		env.AssertAPICall("PUT", "/v1/alerts/"+testutil.AlertID1+"/acknowledge")
	})

	t.Run("acknowledge non-existent alert", func(t *testing.T) {
		result := env.RunError("alert", "ack", "99999999-9999-9999-9999-999999999999")

		env.AssertStderrContains(result, "not found")
	})
}

// TestAlertResolve tests the alert resolve command.
func TestAlertResolve(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("resolve existing alert", func(t *testing.T) {
		result := env.RunSuccess("alert", "resolve", testutil.AlertID1)

		env.AssertContains(result, "resolved")
		env.AssertAPICall("POST", "/v1/alerts/"+testutil.AlertID1+"/resolve")
	})

	t.Run("resolve non-existent alert", func(t *testing.T) {
		result := env.RunError("alert", "resolve", "99999999-9999-9999-9999-999999999999")

		env.AssertStderrContains(result, "not found")
	})
}
