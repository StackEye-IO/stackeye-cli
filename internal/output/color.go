// Package output provides CLI output helpers for StackEye commands.
package output

import (
	"github.com/StackEye-IO/stackeye-go-sdk/config"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

// NewColorManager creates a ColorManager from the current CLI configuration.
// It respects the user's color preference from config, --no-color flag,
// NO_COLOR environment variable, and TERM=dumb detection (via the SDK).
//
// If no config getter is set or config is nil, returns a ColorManager
// in auto mode (colors enabled when stdout is a terminal).
//
// NO_COLOR Handling (Intentional Duplication)
//
// The NO_COLOR environment variable (https://no-color.org/) is checked in two
// places by design:
//
//  1. CLI layer (cmd/root.go initConfig): Checks NO_COLOR early during config
//     initialization and sets ColorModeNever on the config preferences. This
//     ensures the CLI's --no-color flag and NO_COLOR env var are handled
//     consistently at the config level before any output is produced.
//
//  2. SDK layer (stackeye-go-sdk/output/color.go NewColorManager): Checks
//     NO_COLOR in ColorAuto mode as a safety net. This ensures any direct SDK
//     consumer (not just the CLI) also respects the NO_COLOR convention.
//
// This separation of concerns means the SDK works correctly standalone (e.g.,
// in third-party tools using the Go SDK) without depending on CLI config
// initialization. The CLI's early check converts NO_COLOR into a config
// preference (ColorModeNever), which then propagates through NewColorManager
// as an explicit mode rather than relying on the SDK's auto-detection.
func NewColorManager() *sdkoutput.ColorManager {
	mode := sdkoutput.ColorAuto

	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			mode = sdkoutput.ColorMode(cfg.Preferences.Color)
		}
	}

	return sdkoutput.NewColorManager(mode)
}

// NewColorManagerFromConfig creates a ColorManager from an explicit config.
// This is useful in commands that need a ColorManager before the global
// config getter is available.
func NewColorManagerFromConfig(cfg *config.Config) *sdkoutput.ColorManager {
	mode := sdkoutput.ColorAuto

	if cfg != nil && cfg.Preferences != nil {
		mode = sdkoutput.ColorMode(cfg.Preferences.Color)
	}

	return sdkoutput.NewColorManager(mode)
}
