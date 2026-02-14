// Package signal provides graceful signal handling for the StackEye CLI.
//
// It intercepts SIGINT (Ctrl+C) and SIGTERM, cancels in-progress operations
// via context, runs registered cleanup functions, and exits with POSIX-standard
// codes (130 for SIGINT, 143 for SIGTERM).
//
// Usage:
//
//	ctx, handler := signal.Setup()
//	handler.OnCleanup(func() { telemetry.Flush(2 * time.Second) })
//	exitCode := cmd.ExecuteWithContext(ctx)
//	os.Exit(handler.ExitCode(exitCode))
package signal

import (
	"context"
	"os"
	ossignal "os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/StackEye-IO/stackeye-cli/internal/errors"
)

// Handler manages signal interception and cleanup coordination.
type Handler struct {
	cancel   context.CancelFunc
	caught   atomic.Value // stores os.Signal or nil
	cleanups []func()
	mu       sync.Mutex
}

// Setup registers SIGINT and SIGTERM handlers and returns a cancellable context
// along with a Handler for cleanup coordination. The returned context is canceled
// when a signal is received, allowing in-progress operations to terminate gracefully.
func Setup() (context.Context, *Handler) {
	ctx, cancel := context.WithCancel(context.Background())
	h := &Handler{cancel: cancel}

	sigCh := make(chan os.Signal, 1)
	ossignal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigCh:
			h.caught.Store(sig)
			cancel()
		case <-ctx.Done():
			// Context was canceled by normal exit; nothing to do.
		}
		ossignal.Stop(sigCh)
	}()

	return ctx, h
}

// OnCleanup registers a function to be called during graceful shutdown.
// Cleanup functions run in LIFO order (last registered runs first).
func (h *Handler) OnCleanup(fn func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cleanups = append(h.cleanups, fn)
}

// RunCleanups executes all registered cleanup functions in LIFO order.
// Safe to call multiple times; subsequent calls are no-ops.
func (h *Handler) RunCleanups() {
	h.mu.Lock()
	fns := make([]func(), len(h.cleanups))
	copy(fns, h.cleanups)
	h.cleanups = nil
	h.mu.Unlock()

	for i := len(fns) - 1; i >= 0; i-- {
		fns[i]()
	}
}

// Signaled returns true if a signal was caught.
func (h *Handler) Signaled() bool {
	return h.caught.Load() != nil
}

// Signal returns the caught signal, or nil if none was received.
func (h *Handler) Signal() os.Signal {
	v := h.caught.Load()
	if v == nil {
		return nil
	}
	sig, ok := v.(os.Signal)
	if !ok {
		return nil
	}
	return sig
}

// ExitCode returns the appropriate exit code. If a signal was caught, it returns
// the POSIX signal exit code (130 for SIGINT, 143 for SIGTERM). Otherwise it
// returns the provided command exit code unchanged.
func (h *Handler) ExitCode(cmdExitCode int) int {
	sig := h.Signal()
	if sig == nil {
		return cmdExitCode
	}
	switch sig {
	case syscall.SIGINT:
		return errors.ExitSIGINT
	case syscall.SIGTERM:
		return errors.ExitSIGTERM
	default:
		return cmdExitCode
	}
}

// Cancel cancels the handler's context. Called during normal shutdown to
// release the signal-watching goroutine.
func (h *Handler) Cancel() {
	h.cancel()
}
