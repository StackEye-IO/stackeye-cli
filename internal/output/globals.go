// Package output provides CLI output helpers.
package output

import (
	"sync/atomic"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
)

// Package-level globals stored as atomic.Values to ensure thread-safe access.
// These are written once during CLI initialization (cmd/root.go init) and read
// during runtime from any goroutine. Using atomic.Value prevents race detector
// flags when tests run with t.Parallel() or when concurrent goroutines read
// these values.
//
// Each global has a typed load function (returns the function pointer or nil)
// and a store function (called by the public Set* API).

var (
	atomicConfigGetter          atomic.Value // stores func() *config.Config
	atomicNoInputGetter         atomic.Value // stores func() bool
	atomicIsPipedOverride       atomic.Value // stores func() bool
	atomicIsStderrPipedOverride atomic.Value // stores func() bool
)

// sentinelFunc is stored to distinguish "explicitly set to nil" from "never set".
// atomic.Value does not allow storing nil directly, so we use a wrapper.
type configGetterBox struct {
	fn func() *config.Config
}

type boolGetterBox struct {
	fn func() bool
}

// loadConfigGetter returns the current config getter function, or nil if unset.
func loadConfigGetter() func() *config.Config {
	v := atomicConfigGetter.Load()
	if v == nil {
		return nil
	}
	return v.(configGetterBox).fn
}

// storeConfigGetter atomically stores the config getter function.
func storeConfigGetter(fn func() *config.Config) {
	atomicConfigGetter.Store(configGetterBox{fn: fn})
}

// loadNoInputGetter returns the current no-input getter function, or nil if unset.
func loadNoInputGetter() func() bool {
	v := atomicNoInputGetter.Load()
	if v == nil {
		return nil
	}
	return v.(boolGetterBox).fn
}

// storeNoInputGetter atomically stores the no-input getter function.
func storeNoInputGetter(fn func() bool) {
	atomicNoInputGetter.Store(boolGetterBox{fn: fn})
}

// loadIsPipedOverride returns the current piped override function, or nil if unset.
func loadIsPipedOverride() func() bool {
	v := atomicIsPipedOverride.Load()
	if v == nil {
		return nil
	}
	return v.(boolGetterBox).fn
}

// storeIsPipedOverride atomically stores the piped override function.
func storeIsPipedOverride(fn func() bool) {
	atomicIsPipedOverride.Store(boolGetterBox{fn: fn})
}

// loadIsStderrPipedOverride returns the current stderr piped override function, or nil if unset.
func loadIsStderrPipedOverride() func() bool {
	v := atomicIsStderrPipedOverride.Load()
	if v == nil {
		return nil
	}
	return v.(boolGetterBox).fn
}

// storeIsStderrPipedOverride atomically stores the stderr piped override function.
func storeIsStderrPipedOverride(fn func() bool) {
	atomicIsStderrPipedOverride.Store(boolGetterBox{fn: fn})
}
