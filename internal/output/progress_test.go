package output

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

// safeBuffer wraps bytes.Buffer with a mutex for thread-safe access.
type safeBuffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (sb *safeBuffer) Write(p []byte) (n int, err error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Write(p)
}

func (sb *safeBuffer) String() string {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.String()
}

func (sb *safeBuffer) Len() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.buf.Len()
}

func TestSpinner_StartStop(t *testing.T) {
	buf := &safeBuffer{}
	spin := NewSpinner("Testing...", WithWriter(buf), WithInterval(10*time.Millisecond))

	spin.Start()
	time.Sleep(50 * time.Millisecond) // Let a few frames render
	spin.Stop()

	// Check that something was written
	if buf.Len() == 0 {
		t.Error("expected spinner output, got none")
	}

	// Check that the message appears
	if !strings.Contains(buf.String(), "Testing...") {
		t.Errorf("spinner output should contain message, got: %q", buf.String())
	}
}

func TestSpinner_DoubleStart(t *testing.T) {
	buf := &safeBuffer{}
	spin := NewSpinner("Test", WithWriter(buf), WithInterval(10*time.Millisecond))

	spin.Start()
	spin.Start() // Should not panic or cause issues
	time.Sleep(30 * time.Millisecond)
	spin.Stop()
	spin.Stop() // Should not panic
}

func TestSpinner_StopWithMessage(t *testing.T) {
	buf := &safeBuffer{}
	spin := NewSpinner("Loading", WithWriter(buf), WithInterval(10*time.Millisecond))

	spin.Start()
	time.Sleep(30 * time.Millisecond)
	spin.StopWithMessage("Done!")

	if !strings.Contains(buf.String(), "Done!") {
		t.Errorf("output should contain final message, got: %q", buf.String())
	}
}

func TestSpinner_StopWithSuccess(t *testing.T) {
	buf := &safeBuffer{}
	spin := NewSpinner("Processing", WithWriter(buf), WithInterval(10*time.Millisecond))

	spin.Start()
	time.Sleep(30 * time.Millisecond)
	spin.StopWithSuccess("Completed successfully")

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Errorf("output should contain success checkmark, got: %q", output)
	}
	if !strings.Contains(output, "Completed successfully") {
		t.Errorf("output should contain success message, got: %q", output)
	}
}

func TestSpinner_StopWithError(t *testing.T) {
	buf := &safeBuffer{}
	spin := NewSpinner("Processing", WithWriter(buf), WithInterval(10*time.Millisecond))

	spin.Start()
	time.Sleep(30 * time.Millisecond)
	spin.StopWithError("Failed to complete")

	output := buf.String()
	if !strings.Contains(output, "✗") {
		t.Errorf("output should contain error X, got: %q", output)
	}
	if !strings.Contains(output, "Failed to complete") {
		t.Errorf("output should contain error message, got: %q", output)
	}
}

func TestSpinner_UpdateMessage(t *testing.T) {
	buf := &safeBuffer{}
	spin := NewSpinner("Initial", WithWriter(buf), WithInterval(10*time.Millisecond))

	spin.Start()
	time.Sleep(30 * time.Millisecond)
	spin.UpdateMessage("Updated")
	time.Sleep(30 * time.Millisecond)
	spin.Stop()

	output := buf.String()
	if !strings.Contains(output, "Updated") {
		t.Errorf("output should contain updated message, got: %q", output)
	}
}

func TestSpinner_ASCIIFrames(t *testing.T) {
	buf := &safeBuffer{}
	spin := NewSpinner("Test", WithWriter(buf), WithFrames(ASCIISpinnerFrames), WithInterval(10*time.Millisecond))

	spin.Start()
	time.Sleep(50 * time.Millisecond)
	spin.Stop()

	// Verify ASCII frames were used (check for at least one ASCII frame)
	output := buf.String()
	hasASCII := false
	for _, frame := range ASCIISpinnerFrames {
		if strings.Contains(output, frame) {
			hasASCII = true
			break
		}
	}
	if !hasASCII {
		t.Errorf("expected ASCII frames in output, got: %q", output)
	}
}

func TestRunWithSpinner(t *testing.T) {
	// Test successful execution
	result, err := RunWithSpinner("Test operation", func() (string, error) {
		time.Sleep(10 * time.Millisecond)
		return "success", nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %q", result)
	}
}

func TestProgressBar_Basic(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Processing")
	bar.SetWriter(&buf)

	for i := 0; i < 10; i++ {
		bar.Increment()
	}
	bar.Complete()

	output := buf.String()

	// Should show 100%
	if !strings.Contains(output, "100%") {
		t.Errorf("output should show 100%%, got: %q", output)
	}

	// Should show completion count
	if !strings.Contains(output, "10/10") {
		t.Errorf("output should show 10/10, got: %q", output)
	}

	// Should have progress bar characters
	if !strings.Contains(output, "█") {
		t.Errorf("output should contain filled bar characters, got: %q", output)
	}
}

func TestProgressBar_Set(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(100, "Downloading")
	bar.SetWriter(&buf)

	bar.Set(50)

	output := buf.String()
	if !strings.Contains(output, "50/100") {
		t.Errorf("output should show 50/100, got: %q", output)
	}
	if !strings.Contains(output, "50%") {
		t.Errorf("output should show 50%%, got: %q", output)
	}
}

func TestProgressBar_ZeroTotal(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(0, "Test")
	bar.SetWriter(&buf)

	// Should not panic
	bar.Increment()
	bar.Set(5)
	bar.Complete()

	// With zero total, nothing should be rendered
	if buf.Len() > 1 { // Allow for newline from Complete
		t.Errorf("expected minimal output for zero total, got: %q", buf.String())
	}
}

func TestProgressBar_CompleteEarly(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Test")
	bar.SetWriter(&buf)

	bar.Set(5)
	bar.Complete()

	// After complete, further updates should be ignored
	bar.Increment()
	bar.Set(3)

	output := buf.String()
	// Should still show 10/10 from Complete
	if !strings.Contains(output, "10/10") {
		t.Errorf("should show completed state, got: %q", output)
	}
}
