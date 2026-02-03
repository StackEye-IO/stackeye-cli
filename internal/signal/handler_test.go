package signal

import (
	"syscall"
	"testing"
	"time"
)

func TestSetup_ReturnsNonNilContextAndHandler(t *testing.T) {
	ctx, h := Setup()
	defer h.Cancel()

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestHandler_SignaledFalseInitially(t *testing.T) {
	_, h := Setup()
	defer h.Cancel()

	if h.Signaled() {
		t.Error("expected Signaled() == false before any signal")
	}
	if h.Signal() != nil {
		t.Error("expected Signal() == nil before any signal")
	}
}

func TestHandler_ExitCodePassthroughWithoutSignal(t *testing.T) {
	_, h := Setup()
	defer h.Cancel()

	tests := []int{0, 1, 2, 5, 10}
	for _, code := range tests {
		if got := h.ExitCode(code); got != code {
			t.Errorf("ExitCode(%d) = %d; want %d", code, got, code)
		}
	}
}

func TestHandler_ExitCodeSIGINT(t *testing.T) {
	_, h := Setup()
	defer h.Cancel()

	// Simulate signal catch
	h.caught.Store(syscall.SIGINT)

	if !h.Signaled() {
		t.Error("expected Signaled() == true after SIGINT")
	}
	if got := h.ExitCode(0); got != ExitSIGINT {
		t.Errorf("ExitCode(0) = %d; want %d", got, ExitSIGINT)
	}
}

func TestHandler_ExitCodeSIGTERM(t *testing.T) {
	_, h := Setup()
	defer h.Cancel()

	h.caught.Store(syscall.SIGTERM)

	if !h.Signaled() {
		t.Error("expected Signaled() == true after SIGTERM")
	}
	if got := h.ExitCode(0); got != ExitSIGTERM {
		t.Errorf("ExitCode(0) = %d; want %d", got, ExitSIGTERM)
	}
}

func TestHandler_CleanupLIFO(t *testing.T) {
	_, h := Setup()
	defer h.Cancel()

	var order []int
	h.OnCleanup(func() { order = append(order, 1) })
	h.OnCleanup(func() { order = append(order, 2) })
	h.OnCleanup(func() { order = append(order, 3) })

	h.RunCleanups()

	if len(order) != 3 {
		t.Fatalf("expected 3 cleanups to run, got %d", len(order))
	}
	// LIFO: last registered runs first
	expected := []int{3, 2, 1}
	for i, v := range order {
		if v != expected[i] {
			t.Errorf("cleanup[%d] = %d; want %d", i, v, expected[i])
		}
	}
}

func TestHandler_RunCleanupsIdempotent(t *testing.T) {
	_, h := Setup()
	defer h.Cancel()

	count := 0
	h.OnCleanup(func() { count++ })

	h.RunCleanups()
	h.RunCleanups() // second call should be no-op

	if count != 1 {
		t.Errorf("cleanup ran %d times; want 1", count)
	}
}

func TestHandler_ContextCanceledOnSignal(t *testing.T) {
	ctx, h := Setup()
	defer h.Cancel()

	// Send SIGINT to self
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	select {
	case <-ctx.Done():
		// Context was canceled as expected
	case <-time.After(2 * time.Second):
		t.Fatal("context was not canceled within 2 seconds after SIGINT")
	}

	if !h.Signaled() {
		t.Error("expected Signaled() == true after sending SIGINT")
	}
}

func TestExitCodeConstants(t *testing.T) {
	if ExitSIGINT != 130 {
		t.Errorf("ExitSIGINT = %d; want 130", ExitSIGINT)
	}
	if ExitSIGTERM != 143 {
		t.Errorf("ExitSIGTERM = %d; want 143", ExitSIGTERM)
	}
}
