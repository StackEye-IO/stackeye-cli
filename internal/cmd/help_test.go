package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/output"
	"github.com/spf13/cobra"
)

// setupTestConfig creates a test help config and returns a cleanup function.
func setupTestConfig(t *testing.T) {
	t.Helper()
	origConfig := helpConfig
	t.Cleanup(func() { helpConfig = origConfig })

	helpConfig = &HelpConfig{
		ColorManager: output.NewColorManager(output.ColorNever),
		Writer:       &bytes.Buffer{},
	}
}

// createTestRootWithCommands creates a root command with subcommands for testing.
// Commands must be added to a parent and have a Run/RunE for IsAvailableCommand() to return true.
func createTestRootWithCommands(commands ...*cobra.Command) *cobra.Command {
	root := &cobra.Command{Use: "test", Short: "Test root"}
	// Ensure each command has a Run function so IsAvailableCommand() returns true
	for _, cmd := range commands {
		if cmd.Run == nil && cmd.RunE == nil {
			cmd.Run = func(cmd *cobra.Command, args []string) {}
		}
	}
	root.AddCommand(commands...)
	return root
}

func TestGroupCommands(t *testing.T) {
	// Create commands and add them to a root command so IsAvailableCommand() works
	loginCmd := &cobra.Command{Use: "login", Short: "Log in"}
	logoutCmd := &cobra.Command{Use: "logout", Short: "Log out"}
	whoamiCmd := &cobra.Command{Use: "whoami", Short: "Show current user"}
	configCmd := &cobra.Command{Use: "config", Short: "Manage config"}
	versionCmd := &cobra.Command{Use: "version", Short: "Show version"}
	unknownCmd := &cobra.Command{Use: "unknown", Short: "Unknown command"}

	root := createTestRootWithCommands(loginCmd, logoutCmd, whoamiCmd, configCmd, versionCmd, unknownCmd)
	groups := GroupCommands(root.Commands())

	// Verify groups are created correctly
	if len(groups) == 0 {
		t.Fatal("expected at least one command group")
	}

	// Check that Authentication group exists and has correct commands
	var authGroup *CommandGroup
	for i := range groups {
		if groups[i].Name == "Authentication" {
			authGroup = &groups[i]
			break
		}
	}
	if authGroup == nil {
		t.Fatal("expected Authentication group")
	}
	if len(authGroup.Commands) != 3 {
		t.Errorf("expected 3 commands in Authentication group, got %d", len(authGroup.Commands))
	}

	// Check that unknown command goes to "Other Commands"
	var otherGroup *CommandGroup
	for i := range groups {
		if groups[i].Name == "Other Commands" {
			otherGroup = &groups[i]
			break
		}
	}
	if otherGroup == nil {
		t.Fatal("expected Other Commands group")
	}
	if len(otherGroup.Commands) != 1 {
		t.Errorf("expected 1 command in Other Commands group, got %d", len(otherGroup.Commands))
	}
}

func TestGroupCommandsOrder(t *testing.T) {
	// Create commands that span multiple groups
	loginCmd := &cobra.Command{Use: "login", Short: "Log in"}
	configCmd := &cobra.Command{Use: "config", Short: "Manage config"}
	probeCmd := &cobra.Command{Use: "probe", Short: "Manage probes"}
	versionCmd := &cobra.Command{Use: "version", Short: "Show version"}

	root := createTestRootWithCommands(versionCmd, configCmd, loginCmd, probeCmd)
	groups := GroupCommands(root.Commands())

	// Verify order matches groupOrder
	expectedOrder := []string{"Authentication", "Configuration", "Monitoring", "Utilities"}
	for i, g := range groups {
		if i >= len(expectedOrder) {
			break
		}
		if g.Name != expectedOrder[i] {
			t.Errorf("expected group %d to be %q, got %q", i, expectedOrder[i], g.Name)
		}
	}
}

func TestInitHelp(t *testing.T) {
	setupTestConfig(t)

	rootCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}
	subCmd := &cobra.Command{
		Use:   "sub",
		Short: "Subcommand",
	}
	rootCmd.AddCommand(subCmd)

	cfg := &HelpConfig{
		ColorManager: output.NewColorManager(output.ColorNever),
		Writer:       &bytes.Buffer{},
	}

	InitHelp(rootCmd, cfg)

	// Verify config is set
	if helpConfig == nil {
		t.Fatal("expected helpConfig to be set")
	}
	if helpConfig.ColorManager == nil {
		t.Fatal("expected ColorManager to be set")
	}
}

func TestSetNoColor(t *testing.T) {
	setupTestConfig(t)

	// Initialize with colors enabled
	cfg := &HelpConfig{
		ColorManager: output.NewColorManager(output.ColorAuto),
		Writer:       &bytes.Buffer{},
	}
	helpConfig = cfg

	// Disable colors
	SetNoColor(true)

	// Verify colors are disabled
	if helpConfig.ColorManager.Enabled() {
		t.Error("expected colors to be disabled")
	}

	// Re-enable colors
	SetNoColor(false)

	// Colors may or may not be enabled depending on terminal detection
	// Just verify it doesn't panic
}

func TestRenderHelp(t *testing.T) {
	setupTestConfig(t)

	// Create a test command with all features
	// Must have Run function for IsAvailableCommand() to work
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Long:  "This is a long description for the test command.",
		Example: `  # Run test
  test --flag value`,
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.Flags().StringP("flag", "f", "", "A test flag")

	helpOutput, err := RenderHelp(cmd)
	if err != nil {
		t.Fatalf("RenderHelp failed: %v", err)
	}

	// Verify help output contains expected sections
	expectedSections := []string{
		"This is a long description",
		"USAGE",
		"EXAMPLES",
		"FLAGS",
	}
	for _, section := range expectedSections {
		if !strings.Contains(helpOutput, section) {
			t.Errorf("expected help output to contain %q", section)
		}
	}
}

func TestRenderHelpWithSubcommands(t *testing.T) {
	setupTestConfig(t)

	rootCmd := &cobra.Command{
		Use:   "stackeye",
		Short: "StackEye CLI",
		Long:  "StackEye CLI for managing uptime monitoring.",
	}

	// Commands must have Run function for IsAvailableCommand() to return true
	noop := func(cmd *cobra.Command, args []string) {}
	loginCmd := &cobra.Command{Use: "login", Short: "Log in to StackEye", Run: noop}
	logoutCmd := &cobra.Command{Use: "logout", Short: "Log out of StackEye", Run: noop}
	versionCmd := &cobra.Command{Use: "version", Short: "Show version", Run: noop}
	configCmd := &cobra.Command{Use: "config", Short: "Manage configuration", Run: noop}

	rootCmd.AddCommand(loginCmd, logoutCmd, versionCmd, configCmd)

	helpOutput, err := RenderHelp(rootCmd)
	if err != nil {
		t.Fatalf("RenderHelp failed: %v", err)
	}

	// Verify grouped commands appear
	expectedGroups := []string{
		"Authentication:",
		"Configuration:",
		"Utilities:",
	}
	for _, group := range expectedGroups {
		if !strings.Contains(helpOutput, group) {
			t.Errorf("expected help output to contain group %q", group)
		}
	}

	// Verify commands appear
	expectedCommands := []string{"login", "logout", "version", "config"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(helpOutput, cmd) {
			t.Errorf("expected help output to contain command %q", cmd)
		}
	}
}

func TestGetHelpConfig(t *testing.T) {
	// Save and restore original config
	origConfig := helpConfig
	t.Cleanup(func() { helpConfig = origConfig })

	// Clear global config
	helpConfig = nil

	cfg := GetHelpConfig()
	if cfg == nil {
		t.Fatal("expected GetHelpConfig to return non-nil config")
	}
	if cfg.ColorManager == nil {
		t.Error("expected ColorManager to be set in default config")
	}
	if cfg.Writer == nil {
		t.Error("expected Writer to be set in default config")
	}
}

func TestTemplateFuncs(t *testing.T) {
	setupTestConfig(t)

	funcs := templateFuncs()

	// Test rpad
	rpad := funcs["rpad"].(func(string, int) string)
	padded := rpad("test", 10)
	if len(padded) != 10 {
		t.Errorf("expected padded string length 10, got %d", len(padded))
	}
	if !strings.HasPrefix(padded, "test") {
		t.Errorf("expected padded string to start with 'test'")
	}

	// Test bold (with colors disabled, should return unchanged)
	bold := funcs["bold"].(func(string) string)
	result := bold("test")
	if result != "test" {
		t.Errorf("expected bold('test') to return 'test' when colors disabled, got %q", result)
	}

	// Test cyan (with colors disabled)
	cyan := funcs["cyan"].(func(string) string)
	result = cyan("test")
	if result != "test" {
		t.Errorf("expected cyan('test') to return 'test' when colors disabled, got %q", result)
	}

	// Test trimTrailingWhitespaces
	trim := funcs["trimTrailingWhitespaces"].(func(string) string)
	trimmed := trim("test   \n\t")
	if trimmed != "test" {
		t.Errorf("expected trimTrailingWhitespaces to return 'test', got %q", trimmed)
	}
}

func TestHiddenCommandsNotGrouped(t *testing.T) {
	visibleCmd := &cobra.Command{Use: "login", Short: "Log in"}
	hiddenCmd := &cobra.Command{Use: "hidden", Short: "Hidden command", Hidden: true}

	root := createTestRootWithCommands(visibleCmd, hiddenCmd)
	groups := GroupCommands(root.Commands())

	// Count total commands in all groups
	totalCommands := 0
	for _, g := range groups {
		totalCommands += len(g.Commands)
	}

	// Hidden command should not be included
	if totalCommands != 1 {
		t.Errorf("expected 1 command in groups (hidden excluded), got %d", totalCommands)
	}
}

// TestCobraHelpIntegration verifies that the custom help templates work
// when invoked through Cobra's actual help system (not just RenderHelp).
// This is a regression test for the critical bug where template functions
// weren't registered with Cobra, causing panics on --help.
func TestCobraHelpIntegration(t *testing.T) {
	setupTestConfig(t)

	// Create a fresh command tree for testing
	rootCmd := &cobra.Command{
		Use:   "testcli",
		Short: "Test CLI",
		Long:  "A test CLI for integration testing.",
	}
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Log in",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	rootCmd.AddCommand(loginCmd)

	// Initialize help system (this registers template functions with Cobra)
	cfg := &HelpConfig{
		ColorManager: output.NewColorManager(output.ColorNever),
		Writer:       &bytes.Buffer{},
	}
	InitHelp(rootCmd, cfg)

	// Capture help output by setting command's output buffer
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	// Execute --help via Cobra's actual help system
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("command execution failed: %v", err)
	}

	output := buf.String()

	// Verify the custom template was used (grouped commands section)
	if !strings.Contains(output, "Authentication:") {
		t.Error("expected help output to contain grouped command section 'Authentication:'")
	}

	// Verify custom template sections
	expectedSections := []string{"USAGE", "COMMANDS", "FLAGS"}
	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("expected help output to contain section %q", section)
		}
	}
}

// TestCobraSubcommandHelpIntegration verifies subcommand help works via Cobra.
func TestCobraSubcommandHelpIntegration(t *testing.T) {
	setupTestConfig(t)

	rootCmd := &cobra.Command{
		Use:   "testcli",
		Short: "Test CLI",
	}
	loginCmd := &cobra.Command{
		Use:     "login",
		Short:   "Log in to the service",
		Long:    "Authenticate with the service via browser-based OAuth flow.",
		Example: "  testcli login --api-url https://api.example.com",
		Run:     func(cmd *cobra.Command, args []string) {},
	}
	loginCmd.Flags().String("api-url", "", "API URL")
	rootCmd.AddCommand(loginCmd)

	cfg := &HelpConfig{
		ColorManager: output.NewColorManager(output.ColorNever),
		Writer:       &bytes.Buffer{},
	}
	InitHelp(rootCmd, cfg)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	// Execute subcommand --help
	rootCmd.SetArgs([]string{"login", "--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("command execution failed: %v", err)
	}

	output := buf.String()

	// Verify long description appears
	if !strings.Contains(output, "Authenticate with the service") {
		t.Error("expected help output to contain long description")
	}

	// Verify example appears
	if !strings.Contains(output, "EXAMPLES") {
		t.Error("expected help output to contain EXAMPLES section")
	}

	// Verify flags appear
	if !strings.Contains(output, "api-url") {
		t.Error("expected help output to contain --api-url flag")
	}
}
