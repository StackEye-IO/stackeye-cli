package output

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/StackEye-IO/stackeye-go-sdk/config"
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
	spin := NewSpinner("Testing...", WithWriter(buf), WithDisabled(false), WithInterval(10*time.Millisecond))

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
	spin := NewSpinner("Test", WithWriter(buf), WithDisabled(false), WithInterval(10*time.Millisecond))

	spin.Start()
	spin.Start() // Should not panic or cause issues
	time.Sleep(30 * time.Millisecond)
	spin.Stop()
	spin.Stop() // Should not panic
}

func TestSpinner_StopWithMessage(t *testing.T) {
	buf := &safeBuffer{}
	spin := NewSpinner("Loading", WithWriter(buf), WithDisabled(false), WithInterval(10*time.Millisecond))

	spin.Start()
	time.Sleep(30 * time.Millisecond)
	spin.StopWithMessage("Done!")

	if !strings.Contains(buf.String(), "Done!") {
		t.Errorf("output should contain final message, got: %q", buf.String())
	}
}

func TestSpinner_StopWithSuccess(t *testing.T) {
	buf := &safeBuffer{}
	spin := NewSpinner("Processing", WithWriter(buf), WithDisabled(false), WithInterval(10*time.Millisecond))

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
	spin := NewSpinner("Processing", WithWriter(buf), WithDisabled(false), WithInterval(10*time.Millisecond))

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
	spin := NewSpinner("Initial", WithWriter(buf), WithDisabled(false), WithInterval(10*time.Millisecond))

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
	spin := NewSpinner("Test", WithWriter(buf), WithDisabled(false), WithFrames(ASCIISpinnerFrames), WithInterval(10*time.Millisecond))

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
	bar := NewProgressBar(10, "Processing", WithBarWriter(&buf), WithBarDisabled(false))

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
	bar := NewProgressBar(100, "Downloading", WithBarWriter(&buf), WithBarDisabled(false))

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
	bar := NewProgressBar(0, "Test", WithBarWriter(&buf), WithBarDisabled(false))

	// Should not panic
	bar.Increment()
	bar.Set(5)
	bar.Complete()

	// With zero total, nothing should be rendered (Complete is no-op for zero total)
	if buf.Len() != 0 {
		t.Errorf("expected no output for zero total, got: %q", buf.String())
	}
}

func TestProgressBar_CompleteEarly(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Test", WithBarWriter(&buf), WithBarDisabled(false))

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

func TestProgressBar_DisabledNoOutput(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Test", WithBarWriter(&buf), WithBarDisabled(true))

	if !bar.Disabled() {
		t.Error("progress bar should be disabled")
	}

	bar.Increment()
	bar.Set(5)
	bar.Complete()

	if buf.Len() != 0 {
		t.Errorf("disabled progress bar should produce no output, got: %q", buf.String())
	}
}

func TestProgressBar_WithBarDisabledFalseOverride(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Forced", WithBarWriter(&buf), WithBarDisabled(false))

	if bar.Disabled() {
		t.Error("progress bar should be force-enabled with WithBarDisabled(false)")
	}

	bar.Set(5)

	if buf.Len() == 0 {
		t.Error("force-enabled progress bar should produce output")
	}
}

func TestProgressBar_CompleteWithSuccess(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Processing", WithBarWriter(&buf), WithBarDisabled(false))

	for i := 0; i < 10; i++ {
		bar.Increment()
	}
	bar.CompleteWithSuccess("All items processed")

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Errorf("output should contain success checkmark, got: %q", output)
	}
	if !strings.Contains(output, "All items processed") {
		t.Errorf("output should contain success message, got: %q", output)
	}
}

func TestProgressBar_CompleteWithError(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Processing", WithBarWriter(&buf), WithBarDisabled(false))

	bar.Set(5)
	bar.CompleteWithError("connection timeout")

	output := buf.String()
	if !strings.Contains(output, "✗") {
		t.Errorf("output should contain error X, got: %q", output)
	}
	if !strings.Contains(output, "connection timeout") {
		t.Errorf("output should contain error message, got: %q", output)
	}
}

func TestProgressBar_DisabledCompleteMethods(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Test", WithBarWriter(&buf), WithBarDisabled(true))

	bar.Increment()
	bar.CompleteWithSuccess("OK")
	bar.CompleteWithError("Fail")

	if buf.Len() != 0 {
		t.Errorf("disabled progress bar complete methods should produce no output, got: %q", buf.String())
	}
}

func TestProgressBar_WithBarWidth(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Test", WithBarWriter(&buf), WithBarDisabled(false), WithBarWidth(10))

	bar.Set(5)

	output := buf.String()
	// With width 10 at 50%, expect 5 filled + 5 empty
	filledCount := strings.Count(output, "█")
	emptyCount := strings.Count(output, "░")
	if filledCount != 5 {
		t.Errorf("expected 5 filled chars with width=10 at 50%%, got %d", filledCount)
	}
	if emptyCount != 5 {
		t.Errorf("expected 5 empty chars with width=10 at 50%%, got %d", emptyCount)
	}
}

func TestProgressBar_WithBarWidthInvalid(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Test", WithBarWriter(&buf), WithBarDisabled(false), WithBarWidth(0))

	bar.Set(5)

	output := buf.String()
	// Width 0 should be ignored, default 30 should be used
	filledCount := strings.Count(output, "█")
	emptyCount := strings.Count(output, "░")
	if filledCount+emptyCount != 30 {
		t.Errorf("expected default width 30, got %d total chars", filledCount+emptyCount)
	}
}

func TestProgressBar_ETA(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(100, "Processing", WithBarWriter(&buf), WithBarDisabled(false))

	// Manually set start time to simulate elapsed time for ETA calculation
	bar.mu.Lock()
	bar.startTime = time.Now().Add(-5 * time.Second)
	bar.mu.Unlock()

	bar.Set(50)

	output := buf.String()
	if !strings.Contains(output, "ETA:") {
		t.Errorf("output should contain ETA after sufficient elapsed time, got: %q", output)
	}
}

func TestProgressBar_NoETAWhenComplete(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Test", WithBarWriter(&buf), WithBarDisabled(false))

	bar.mu.Lock()
	bar.startTime = time.Now().Add(-5 * time.Second)
	bar.mu.Unlock()

	bar.Complete()

	// The final render at Complete should not show ETA
	output := buf.String()
	// Count ETA occurrences - the Complete render should not have one
	// Split by \r to get individual renders
	renders := strings.Split(output, "\r")
	lastRender := renders[len(renders)-1]
	if strings.Contains(lastRender, "ETA:") {
		t.Errorf("completed progress bar should not show ETA, last render: %q", lastRender)
	}
}

func TestProgressBar_SetWriterDeprecated(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Test", WithBarDisabled(false))
	bar.SetWriter(&buf)

	bar.Set(5)

	if buf.Len() == 0 {
		t.Error("SetWriter should still work for backward compatibility")
	}
}

func TestRunWithProgressBar_Success(t *testing.T) {
	err := RunWithProgressBar(5, "Processing", func(bar *ProgressBar) error {
		for i := 0; i < 5; i++ {
			bar.Increment()
		}
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunWithProgressBar_Error(t *testing.T) {
	expectedErr := fmt.Errorf("test error")
	err := RunWithProgressBar(5, "Processing", func(bar *ProgressBar) error {
		bar.Increment()
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSpinner_DisabledNoOutput(t *testing.T) {
	buf := &safeBuffer{}
	spin := NewSpinner("Test", WithWriter(buf), WithDisabled(true), WithInterval(10*time.Millisecond))

	if !spin.Disabled() {
		t.Error("spinner should be disabled")
	}

	spin.Start()
	time.Sleep(50 * time.Millisecond)
	spin.Stop()

	if buf.Len() != 0 {
		t.Errorf("disabled spinner should produce no output, got: %q", buf.String())
	}
}

func TestSpinner_DisabledStopMethods(t *testing.T) {
	// Disabled spinner's stop methods should still work without panic
	buf := &safeBuffer{}
	spin := NewSpinner("Test", WithWriter(buf), WithDisabled(true))

	spin.Start()
	spin.StopWithMessage("Done")
	spin.StopWithSuccess("OK")
	spin.StopWithError("Fail")
	// No panic means success
}

func TestSpinner_WithDisabledFalseOverride(t *testing.T) {
	buf := &safeBuffer{}
	// Force-enable even though test environment may not be TTY
	spin := NewSpinner("Forced", WithWriter(buf), WithDisabled(false), WithInterval(10*time.Millisecond))

	if spin.Disabled() {
		t.Error("spinner should be force-enabled with WithDisabled(false)")
	}

	spin.Start()
	time.Sleep(30 * time.Millisecond)
	spin.Stop()

	if buf.Len() == 0 {
		t.Error("force-enabled spinner should produce output")
	}
}

func TestSpinner_NoInputDisables(t *testing.T) {
	// Save and restore
	origGetter := noInputGetter
	origConfig := configGetter
	defer func() {
		noInputGetter = origGetter
		configGetter = origConfig
	}()

	noInputGetter = func() bool { return true }
	configGetter = nil

	if isAnimationEnabled() {
		t.Error("spinner should be disabled when --no-input is set")
	}
}

func TestSpinner_JSONOutputDisables(t *testing.T) {
	// Save and restore
	origGetter := noInputGetter
	origConfig := configGetter
	defer func() {
		noInputGetter = origGetter
		configGetter = origConfig
	}()

	noInputGetter = nil
	configGetter = func() *config.Config {
		return &config.Config{
			Preferences: &config.Preferences{
				OutputFormat: "json",
			},
		}
	}

	if isAnimationEnabled() {
		t.Error("spinner should be disabled for JSON output")
	}
}

func TestSpinner_YAMLOutputDisables(t *testing.T) {
	// Save and restore
	origGetter := noInputGetter
	origConfig := configGetter
	defer func() {
		noInputGetter = origGetter
		configGetter = origConfig
	}()

	noInputGetter = nil
	configGetter = func() *config.Config {
		return &config.Config{
			Preferences: &config.Preferences{
				OutputFormat: "yaml",
			},
		}
	}

	if isAnimationEnabled() {
		t.Error("spinner should be disabled for YAML output")
	}
}

func TestSpinner_StackeyeNoInputEnvDisables(t *testing.T) {
	// Save and restore
	origGetter := noInputGetter
	origConfig := configGetter
	defer func() {
		noInputGetter = origGetter
		configGetter = origConfig
	}()

	noInputGetter = nil
	configGetter = nil
	t.Setenv("STACKEYE_NO_INPUT", "1")

	if isAnimationEnabled() {
		t.Error("spinner should be disabled when STACKEYE_NO_INPUT=1")
	}
}

func TestSpinner_TableOutputAllowed(t *testing.T) {
	// Save and restore
	origGetter := noInputGetter
	origConfig := configGetter
	defer func() {
		noInputGetter = origGetter
		configGetter = origConfig
	}()

	noInputGetter = func() bool { return false }
	configGetter = func() *config.Config {
		return &config.Config{
			Preferences: &config.Preferences{
				OutputFormat: "table",
			},
		}
	}

	// Note: isAnimationEnabled() may still return false due to TTY detection
	// in test environments, so we test the config/flag path specifically
	// by checking that table format doesn't trigger the format disable path

	// We can't directly test TTY in a test, but we verify the config logic
	// doesn't block table format
	buf := &safeBuffer{}
	spin := NewSpinner("Table test", WithWriter(buf), WithDisabled(false), WithInterval(10*time.Millisecond))

	spin.Start()
	time.Sleep(30 * time.Millisecond)
	spin.Stop()

	if buf.Len() == 0 {
		t.Error("spinner with table output and WithDisabled(false) should produce output")
	}
}

func TestRunWithSpinner_DisabledStillReturnsResult(t *testing.T) {
	// Even when spinner is disabled, RunWithSpinner must still execute the function
	// and return its result.
	// Save and force disable
	origGetter := noInputGetter
	origConfig := configGetter
	defer func() {
		noInputGetter = origGetter
		configGetter = origConfig
	}()
	noInputGetter = func() bool { return true }
	configGetter = nil

	result, err := RunWithSpinner("Disabled test", func() (string, error) {
		return "executed", nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "executed" {
		t.Errorf("expected 'executed', got %q", result)
	}
}

func TestProgressBar_ConcurrentIncrement(t *testing.T) {
	const numGoroutines = 200
	buf := &safeBuffer{}
	bar := NewProgressBar(numGoroutines, "Concurrent test", WithBarWriter(buf), WithBarDisabled(false))

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			bar.Increment()
		}()
	}

	wg.Wait()

	// Verify final count matches expected total
	bar.mu.Lock()
	current := bar.current
	bar.mu.Unlock()

	if current != numGoroutines {
		t.Errorf("expected current=%d after %d concurrent Increment() calls, got %d",
			numGoroutines, numGoroutines, current)
	}

	bar.Complete()

	output := buf.String()
	expectedCount := fmt.Sprintf("%d/%d", numGoroutines, numGoroutines)
	if !strings.Contains(output, expectedCount) {
		t.Errorf("output should contain %q after Complete(), got: %q", expectedCount, output)
	}
	if !strings.Contains(output, "100%") {
		t.Errorf("output should show 100%% after all increments, got: %q", output)
	}
}

func TestProgressBar_ConcurrentMixedOperations(t *testing.T) {
	const numGoroutines = 100
	buf := &safeBuffer{}
	bar := NewProgressBar(numGoroutines*2, "Mixed concurrent", WithBarWriter(buf), WithBarDisabled(false))

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Half the goroutines call Increment
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			bar.Increment()
		}()
	}

	// Other half also call Increment
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			bar.Increment()
		}()
	}

	wg.Wait()

	bar.mu.Lock()
	current := bar.current
	bar.mu.Unlock()

	if current != numGoroutines*2 {
		t.Errorf("expected current=%d after %d concurrent Increment() calls, got %d",
			numGoroutines*2, numGoroutines*2, current)
	}

	bar.Complete()
}

func TestProgressBar_AutoDisabledInNonTTY(t *testing.T) {
	// In test environments stderr is piped, so NewProgressBar should
	// auto-disable without WithBarDisabled(true) being explicitly set.
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Auto-disable test", WithBarWriter(&buf))

	if !bar.Disabled() {
		t.Error("ProgressBar should auto-disable when stderr is not a TTY")
	}

	bar.Increment()
	bar.Set(5)
	bar.Complete()

	if buf.Len() != 0 {
		t.Errorf("auto-disabled ProgressBar should produce no output, got: %q", buf.String())
	}
}

func TestProgressBar_AutoDisabledCompleteMethodsNoOutput(t *testing.T) {
	// Verify all complete variants produce no output when auto-disabled
	var buf bytes.Buffer
	bar := NewProgressBar(10, "Complete methods test", WithBarWriter(&buf))

	bar.Increment()
	bar.CompleteWithSuccess("should not appear")

	if buf.Len() != 0 {
		t.Errorf("auto-disabled ProgressBar CompleteWithSuccess should produce no output, got: %q", buf.String())
	}

	var buf2 bytes.Buffer
	bar2 := NewProgressBar(10, "Error test", WithBarWriter(&buf2))
	bar2.Increment()
	bar2.CompleteWithError("should not appear")

	if buf2.Len() != 0 {
		t.Errorf("auto-disabled ProgressBar CompleteWithError should produce no output, got: %q", buf2.String())
	}
}

func TestProgressBar_NoInputDisablesBar(t *testing.T) {
	origGetter := noInputGetter
	origConfig := configGetter
	origStderr := isStderrPipedOverride
	defer func() {
		noInputGetter = origGetter
		configGetter = origConfig
		isStderrPipedOverride = origStderr
	}()

	// Simulate TTY stderr so only --no-input triggers the disable
	isStderrPipedOverride = func() bool { return false }
	noInputGetter = func() bool { return true }
	configGetter = nil
	t.Setenv("TERM", "xterm-256color")

	var buf bytes.Buffer
	bar := NewProgressBar(10, "No-input test", WithBarWriter(&buf))

	if !bar.Disabled() {
		t.Error("ProgressBar should be disabled when --no-input is set")
	}

	bar.Increment()
	bar.Complete()

	if buf.Len() != 0 {
		t.Errorf("disabled ProgressBar should produce no output, got: %q", buf.String())
	}
}

func TestProgressBar_JSONOutputDisablesBar(t *testing.T) {
	origGetter := noInputGetter
	origConfig := configGetter
	origStderr := isStderrPipedOverride
	defer func() {
		noInputGetter = origGetter
		configGetter = origConfig
		isStderrPipedOverride = origStderr
	}()

	isStderrPipedOverride = func() bool { return false }
	noInputGetter = func() bool { return false }
	configGetter = func() *config.Config {
		return &config.Config{
			Preferences: &config.Preferences{
				OutputFormat: "json",
			},
		}
	}
	t.Setenv("TERM", "xterm-256color")

	var buf bytes.Buffer
	bar := NewProgressBar(10, "JSON test", WithBarWriter(&buf))

	if !bar.Disabled() {
		t.Error("ProgressBar should be disabled for JSON output format")
	}

	bar.Increment()
	bar.Complete()

	if buf.Len() != 0 {
		t.Errorf("disabled ProgressBar should produce no output, got: %q", buf.String())
	}
}

func TestProgressBar_YAMLOutputDisablesBar(t *testing.T) {
	origGetter := noInputGetter
	origConfig := configGetter
	origStderr := isStderrPipedOverride
	defer func() {
		noInputGetter = origGetter
		configGetter = origConfig
		isStderrPipedOverride = origStderr
	}()

	isStderrPipedOverride = func() bool { return false }
	noInputGetter = func() bool { return false }
	configGetter = func() *config.Config {
		return &config.Config{
			Preferences: &config.Preferences{
				OutputFormat: "yaml",
			},
		}
	}
	t.Setenv("TERM", "xterm-256color")

	var buf bytes.Buffer
	bar := NewProgressBar(10, "YAML test", WithBarWriter(&buf))

	if !bar.Disabled() {
		t.Error("ProgressBar should be disabled for YAML output format")
	}

	bar.Increment()
	bar.Complete()

	if buf.Len() != 0 {
		t.Errorf("disabled ProgressBar should produce no output, got: %q", buf.String())
	}
}

func TestProgressBar_StackeyeNoInputEnvDisablesBar(t *testing.T) {
	origGetter := noInputGetter
	origConfig := configGetter
	origStderr := isStderrPipedOverride
	defer func() {
		noInputGetter = origGetter
		configGetter = origConfig
		isStderrPipedOverride = origStderr
	}()

	isStderrPipedOverride = func() bool { return false }
	noInputGetter = nil
	configGetter = nil
	t.Setenv("STACKEYE_NO_INPUT", "1")
	t.Setenv("TERM", "xterm-256color")

	var buf bytes.Buffer
	bar := NewProgressBar(10, "Env test", WithBarWriter(&buf))

	if !bar.Disabled() {
		t.Error("ProgressBar should be disabled when STACKEYE_NO_INPUT=1")
	}

	bar.Increment()
	bar.Complete()

	if buf.Len() != 0 {
		t.Errorf("disabled ProgressBar should produce no output, got: %q", buf.String())
	}
}

func TestProgressBar_RunWithProgressBarDisabledStillExecutes(t *testing.T) {
	// Even when ProgressBar is auto-disabled, RunWithProgressBar must
	// still execute the callback and return its result.
	executed := false
	err := RunWithProgressBar(5, "Disabled test", func(bar *ProgressBar) error {
		executed = true
		for i := 0; i < 5; i++ {
			bar.Increment()
		}
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("RunWithProgressBar should execute callback even when disabled")
	}
}

func TestSetNoInputGetter(t *testing.T) {
	origGetter := noInputGetter
	defer func() { noInputGetter = origGetter }()

	called := false
	SetNoInputGetter(func() bool {
		called = true
		return true
	})

	if noInputGetter == nil {
		t.Fatal("noInputGetter should be set")
	}

	noInputGetter()
	if !called {
		t.Error("noInputGetter was not called after SetNoInputGetter")
	}
}
