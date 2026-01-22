package e2e

import (
	"testing"

	"github.com/StackEye-IO/stackeye-cli/test/e2e/testutil"
)

// TestProbeList tests the probe list command.
func TestProbeList(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("default table output", func(t *testing.T) {
		result := env.RunSuccess("probe", "list")

		// Should contain fixture probe data
		env.AssertContains(result, "API Health Check")
		env.AssertContains(result, "Website Monitor")
		env.AssertAPICall("GET", "/v1/probes")
	})

	t.Run("json output format", func(t *testing.T) {
		env.ClearAPICalls()
		result := env.RunSuccess("probe", "list", "--output", "json")

		// Should be valid JSON
		var data any
		env.AssertJSONOutput(result, &data)
		env.AssertAPICall("GET", "/v1/probes")
	})

	t.Run("yaml output format", func(t *testing.T) {
		env.ClearAPICalls()
		result := env.RunSuccess("probe", "list", "--output", "yaml")

		// Should contain YAML markers
		env.AssertContains(result, "name:")
		env.AssertAPICall("GET", "/v1/probes")
	})
}

// TestProbeGet tests the probe get command.
func TestProbeGet(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("get existing probe", func(t *testing.T) {
		result := env.RunSuccess("probe", "get", testutil.ProbeID1)

		env.AssertContains(result, "API Health Check")
		env.AssertAPICall("GET", "/v1/probes/"+testutil.ProbeID1)
	})

	t.Run("get non-existent probe", func(t *testing.T) {
		result := env.RunError("probe", "get", "99999999-9999-9999-9999-999999999999")

		env.AssertStderrContains(result, "not found")
	})

	t.Run("json output format", func(t *testing.T) {
		env.ClearAPICalls()
		result := env.RunSuccess("probe", "get", testutil.ProbeID1, "--output", "json")

		var data any
		env.AssertJSONOutput(result, &data)
	})
}

// TestProbeCreate tests the probe create command.
func TestProbeCreate(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("create http probe", func(t *testing.T) {
		result := env.RunSuccess(
			"probe", "create",
			"--name", "New Test Probe",
			"--check-type", "http",
			"--url", "https://example.com/health",
			"--interval", "60",
		)

		// CLI outputs probe details on success - check for probe name in output
		env.AssertContains(result, "New Probe")
		env.AssertAPICall("POST", "/v1/probes")
	})

	t.Run("missing required flags", func(t *testing.T) {
		env.ClearAPICalls()
		_ = env.RunError("probe", "create", "--name", "Test")

		// Should fail without required fields
		env.AssertNoAPICall("POST", "/v1/probes")
	})
}

// TestProbeDelete tests the probe delete command.
func TestProbeDelete(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("delete existing probe", func(t *testing.T) {
		_ = env.RunSuccess("probe", "delete", testutil.ProbeID1, "--yes")

		env.AssertAPICall("DELETE", "/v1/probes/"+testutil.ProbeID1)
	})

	t.Run("delete non-existent probe", func(t *testing.T) {
		result := env.RunError("probe", "delete", "99999999-9999-9999-9999-999999999999", "--yes")

		// CLI wraps the error message - check for the wrapped error format
		env.AssertStderrContains(result, "failed to delete")
	})
}

// TestProbePause tests the probe pause command.
func TestProbePause(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("pause existing probe", func(t *testing.T) {
		result := env.RunSuccess("probe", "pause", testutil.ProbeID1, "--yes")

		env.AssertContains(result, "Paused")
		env.AssertAPICall("POST", "/v1/probes/"+testutil.ProbeID1+"/pause")
	})
}

// TestProbeResume tests the probe resume command.
func TestProbeResume(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("resume existing probe", func(t *testing.T) {
		result := env.RunSuccess("probe", "resume", testutil.ProbeID1, "--yes")

		env.AssertContains(result, "Resumed")
		env.AssertAPICall("POST", "/v1/probes/"+testutil.ProbeID1+"/resume")
	})
}

// TestProbeTest tests the probe test command.
func TestProbeTest(t *testing.T) {
	env := testutil.NewTestEnv(t)
	defer env.Cleanup()

	t.Run("test existing probe", func(t *testing.T) {
		result := env.RunSuccess("probe", "test", testutil.ProbeID1)

		env.AssertContains(result, "success")
		// CLI fetches probe first, then calls ad-hoc test endpoint
		env.AssertAPICall("GET", "/v1/probes/"+testutil.ProbeID1)
		env.AssertAPICall("POST", "/v1/probes/test")
	})
}
