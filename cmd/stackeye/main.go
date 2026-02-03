// Package main is the entry point for the StackEye CLI tool.
//
// The StackEye CLI provides command-line access to the StackEye uptime
// monitoring platform. It supports authentication, probe management,
// alert handling, and organization administration.
//
// Usage:
//
//	stackeye [command] [flags]
//
// For more information, run:
//
//	stackeye --help
package main

import (
	"os"

	"github.com/StackEye-IO/stackeye-cli/internal/cmd"
	clisignal "github.com/StackEye-IO/stackeye-cli/internal/signal"
)

func main() {
	ctx, handler := clisignal.Setup()
	exitCode := cmd.ExecuteWithContext(ctx)
	handler.RunCleanups()
	os.Exit(handler.ExitCode(exitCode))
}
