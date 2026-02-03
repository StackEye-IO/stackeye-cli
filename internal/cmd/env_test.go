package cmd

import (
	"testing"
)

func TestNewEnvCmd(t *testing.T) {
	cmd := NewEnvCmd()

	if cmd.Use != "env" {
		t.Errorf("expected Use to be 'env', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}

func TestNewEnvCmd_SkipsConfigLoading(t *testing.T) {
	cmd := NewEnvCmd()

	if cmd.PersistentPreRunE == nil {
		t.Fatal("expected PersistentPreRunE to be set (config loading override)")
	}

	// PersistentPreRunE should return nil (skip config loading)
	err := cmd.PersistentPreRunE(cmd, nil)
	if err != nil {
		t.Errorf("expected PersistentPreRunE to return nil, got %v", err)
	}
}

func TestRunEnv_DoesNotPanic(t *testing.T) {
	// runEnv should work without any config loaded
	// It may produce output to stdout, which is fine for this test.
	// We just verify it doesn't panic or return an error.
	err := runEnv()
	if err != nil {
		t.Errorf("expected runEnv to succeed, got %v", err)
	}
}
