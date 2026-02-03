package output

import (
	"testing"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
	sdkoutput "github.com/StackEye-IO/stackeye-go-sdk/output"
)

func TestNewColorManager_NoConfig(t *testing.T) {
	// Save and restore original config getter
	orig := configGetter
	defer func() { configGetter = orig }()

	configGetter = nil

	cm := NewColorManager()
	if cm == nil {
		t.Fatal("NewColorManager() returned nil without config getter")
	}
	if cm.Mode() != sdkoutput.ColorAuto {
		t.Errorf("Mode() = %v, want %v", cm.Mode(), sdkoutput.ColorAuto)
	}
}

func TestNewColorManager_NilConfig(t *testing.T) {
	orig := configGetter
	defer func() { configGetter = orig }()

	configGetter = func() *config.Config { return nil }

	cm := NewColorManager()
	if cm == nil {
		t.Fatal("NewColorManager() returned nil with nil config")
	}
	if cm.Mode() != sdkoutput.ColorAuto {
		t.Errorf("Mode() = %v, want %v", cm.Mode(), sdkoutput.ColorAuto)
	}
}

func TestNewColorManager_ColorNever(t *testing.T) {
	orig := configGetter
	defer func() { configGetter = orig }()

	configGetter = func() *config.Config {
		return &config.Config{
			Preferences: &config.Preferences{
				Color: config.ColorModeNever,
			},
		}
	}

	cm := NewColorManager()
	if cm.Enabled() {
		t.Error("NewColorManager() should be disabled when config says ColorNever")
	}
}

func TestNewColorManager_ColorAlways(t *testing.T) {
	orig := configGetter
	defer func() { configGetter = orig }()

	configGetter = func() *config.Config {
		return &config.Config{
			Preferences: &config.Preferences{
				Color: config.ColorModeAlways,
			},
		}
	}

	cm := NewColorManager()
	if !cm.Enabled() {
		t.Error("NewColorManager() should be enabled when config says ColorAlways")
	}
}

func TestNewColorManagerFromConfig_Nil(t *testing.T) {
	cm := NewColorManagerFromConfig(nil)
	if cm == nil {
		t.Fatal("NewColorManagerFromConfig(nil) returned nil")
	}
	if cm.Mode() != sdkoutput.ColorAuto {
		t.Errorf("Mode() = %v, want %v", cm.Mode(), sdkoutput.ColorAuto)
	}
}

func TestNewColorManagerFromConfig_Never(t *testing.T) {
	cfg := &config.Config{
		Preferences: &config.Preferences{
			Color: config.ColorModeNever,
		},
	}

	cm := NewColorManagerFromConfig(cfg)
	if cm.Enabled() {
		t.Error("NewColorManagerFromConfig() should be disabled with ColorNever")
	}
}

func TestNewColorManager_NoColorEnv(t *testing.T) {
	orig := configGetter
	defer func() { configGetter = orig }()

	// Config says auto, but NO_COLOR env should cause SDK to disable colors
	configGetter = func() *config.Config {
		return &config.Config{
			Preferences: &config.Preferences{
				Color: config.ColorModeAuto,
			},
		}
	}

	t.Setenv("NO_COLOR", "1")

	cm := NewColorManager()
	if cm.Enabled() {
		t.Error("NewColorManager() should be disabled when NO_COLOR is set")
	}
}
