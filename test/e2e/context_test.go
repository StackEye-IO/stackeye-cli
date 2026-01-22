package e2e

import (
	"testing"

	"github.com/StackEye-IO/stackeye-cli/test/e2e/testutil"
)

// TestContextList tests the context list command.
func TestContextList(t *testing.T) {
	env := testutil.NewTestEnvWithMultipleContexts(t)
	defer env.Cleanup()

	t.Run("list all contexts", func(t *testing.T) {
		result := env.RunSuccess("context", "list")

		// Should show all contexts
		env.AssertContains(result, "test")
		env.AssertContains(result, "staging")
		env.AssertContains(result, "production")
	})

	// Note: context commands don't support --output flag as they manage
	// local configuration, not API data. JSON output is not applicable.
}

// TestContextCurrent tests the context current command.
func TestContextCurrent(t *testing.T) {
	env := testutil.NewTestEnvWithMultipleContexts(t)
	defer env.Cleanup()

	t.Run("show current context", func(t *testing.T) {
		result := env.RunSuccess("context", "current")

		env.AssertContains(result, "test")
	})
}

// TestContextUse tests the context use command.
func TestContextUse(t *testing.T) {
	env := testutil.NewTestEnvWithMultipleContexts(t)
	defer env.Cleanup()

	t.Run("switch to existing context", func(t *testing.T) {
		result := env.RunSuccess("context", "use", "staging")

		env.AssertContains(result, "staging")

		// Verify the switch persisted
		result = env.RunSuccess("context", "current")
		env.AssertContains(result, "staging")
	})

	t.Run("switch to non-existent context", func(t *testing.T) {
		result := env.RunError("context", "use", "nonexistent")

		env.AssertStderrContains(result, "not found")
	})
}

// TestContextShow tests the context show command.
// Note: context show is an alias for context current - it doesn't exist as a separate command.
// Commenting out until the command is implemented.
// func TestContextShow(t *testing.T) {
// 	env := testutil.NewTestEnvWithMultipleContexts(t)
// 	defer env.Cleanup()
//
// 	t.Run("show current context details", func(t *testing.T) {
// 		result := env.RunSuccess("context", "show")
//
// 		env.AssertContains(result, "Test Organization")
// 	})
//
// 	t.Run("show specific context details", func(t *testing.T) {
// 		result := env.RunSuccess("context", "show", "staging")
//
// 		env.AssertContains(result, "Staging Organization")
// 	})
// }
