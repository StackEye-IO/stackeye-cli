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
