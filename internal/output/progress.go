// Package output provides CLI output helpers.
package output

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// noInputGetter returns true if interactive prompts should be disabled.
// This is set by the cmd package to avoid circular imports.
var noInputGetter func() bool

// SetNoInputGetter sets the function used to check if --no-input is active.
// This should be called once during CLI initialization from the cmd package.
func SetNoInputGetter(getter func() bool) {
	noInputGetter = getter
}

// isAnimationEnabled checks whether animated progress indicators (spinners,
// progress bars) should be shown based on:
// - stderr is a TTY (no animations when piped)
// - terminal supports ANSI (not dumb terminal)
// - interactive mode is active (no --no-input flag, no STACKEYE_NO_INPUT env)
// - output format is human-readable (no animations for JSON/YAML)
func isAnimationEnabled() bool {
	// Spinners write to stderr, so check stderr specifically
	if IsStderrPiped() {
		return false
	}

	// Check interactive mode (TERM=dumb, --no-input, STACKEYE_NO_INPUT)
	if IsDumbTerminal() {
		return false
	}

	if isNoInputRequested() {
		return false
	}

	// Check if output format is machine-readable (JSON/YAML)
	if configGetter != nil {
		if cfg := configGetter(); cfg != nil && cfg.Preferences != nil {
			format := cfg.Preferences.OutputFormat
			if format == "json" || format == "yaml" {
				return false
			}
		}
	}

	return true
}

// Spinner provides a simple text-based spinner for long-running operations.
// It displays a spinning animation with an optional message to indicate progress.
// The spinner auto-disables when stderr is not a TTY, when --no-input is set,
// or when output format is JSON/YAML.
//
// Usage:
//
//	spin := NewSpinner("Loading probes...")
//	spin.Start()
//	defer spin.Stop()
//	// ... perform long operation
//	spin.Stop()
type Spinner struct {
	message  string
	writer   io.Writer
	frames   []string
	interval time.Duration
	active   bool
	disabled bool
	mu       sync.Mutex
	done     chan struct{}
}

// SpinnerOption configures a Spinner.
type SpinnerOption func(*Spinner)

// WithWriter sets the output writer for the spinner.
func WithWriter(w io.Writer) SpinnerOption {
	return func(s *Spinner) {
		s.writer = w
	}
}

// WithFrames sets custom spinner frames.
func WithFrames(frames []string) SpinnerOption {
	return func(s *Spinner) {
		s.frames = frames
	}
}

// WithInterval sets the animation interval.
func WithInterval(d time.Duration) SpinnerOption {
	return func(s *Spinner) {
		s.interval = d
	}
}

// WithDisabled explicitly sets the spinner's disabled state.
// When disabled, Start() is a no-op and the spinner produces no output.
func WithDisabled(disabled bool) SpinnerOption {
	return func(s *Spinner) {
		s.disabled = disabled
	}
}

// DefaultSpinnerFrames provides a simple ASCII spinner animation.
var DefaultSpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ASCIISpinnerFrames provides an ASCII-only spinner for terminals without Unicode.
var ASCIISpinnerFrames = []string{"|", "/", "-", "\\"}

// NewSpinner creates a new spinner with the given message.
// The spinner auto-disables when stderr is not a TTY, --no-input is set,
// STACKEYE_NO_INPUT is set, or output format is JSON/YAML.
// Use WithDisabled(false) to force-enable the spinner regardless of environment.
func NewSpinner(message string, opts ...SpinnerOption) *Spinner {
	s := &Spinner{
		message:  message,
		writer:   os.Stderr,
		frames:   DefaultSpinnerFrames,
		interval: 80 * time.Millisecond,
		disabled: !isAnimationEnabled(),
		done:     make(chan struct{}),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Start begins the spinner animation.
// The spinner runs in a separate goroutine until Stop is called.
// If the spinner is disabled (non-TTY, --no-input, JSON/YAML output),
// this is a no-op.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active || s.disabled {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.done = make(chan struct{})
	s.mu.Unlock()

	go s.spin()
}

// Stop halts the spinner animation and clears the line.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	close(s.done)
	writer := s.writer
	s.mu.Unlock()

	// Clear the spinner line
	fmt.Fprintf(writer, "\r\033[K")
}

// StopWithMessage stops the spinner and prints a final message.
// If the spinner is disabled, no output is produced.
func (s *Spinner) StopWithMessage(message string) {
	s.Stop()
	s.mu.Lock()
	disabled := s.disabled
	writer := s.writer
	s.mu.Unlock()
	if !disabled {
		fmt.Fprintln(writer, message)
	}
}

// StopWithSuccess stops the spinner with a success indicator.
// If the spinner is disabled, no output is produced.
func (s *Spinner) StopWithSuccess(message string) {
	s.Stop()
	s.mu.Lock()
	disabled := s.disabled
	writer := s.writer
	s.mu.Unlock()
	if !disabled {
		fmt.Fprintf(writer, "✓ %s\n", message)
	}
}

// StopWithError stops the spinner with an error indicator.
// If the spinner is disabled, no output is produced.
func (s *Spinner) StopWithError(message string) {
	s.Stop()
	s.mu.Lock()
	disabled := s.disabled
	writer := s.writer
	s.mu.Unlock()
	if !disabled {
		fmt.Fprintf(writer, "✗ %s\n", message)
	}
}

// UpdateMessage changes the spinner message while running.
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// Disabled returns true if the spinner is disabled and will not display output.
func (s *Spinner) Disabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.disabled
}

// spin runs the animation loop.
func (s *Spinner) spin() {
	// Copy immutable fields at start to avoid races
	s.mu.Lock()
	interval := s.interval
	frames := s.frames
	writer := s.writer
	s.mu.Unlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	frame := 0
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.mu.Lock()
			message := s.message
			s.mu.Unlock()

			// Clear line and print current frame
			fmt.Fprintf(writer, "\r\033[K%s %s", frames[frame], message)

			frame = (frame + 1) % len(frames)
		}
	}
}

// RunWithSpinner executes a function while displaying a spinner.
// The spinner is automatically stopped when the function completes.
//
// Example:
//
//	result, err := output.RunWithSpinner("Loading probes...", func() (any, error) {
//	    return client.ListProbes(ctx)
//	})
func RunWithSpinner[T any](message string, fn func() (T, error)) (T, error) {
	spin := NewSpinner(message)
	spin.Start()
	defer spin.Stop()

	return fn()
}

// RunWithSpinnerCtx executes a function with context while displaying a spinner.
// The spinner is stopped when the function completes or the context is canceled.
func RunWithSpinnerCtx[T any](ctx context.Context, message string, fn func(context.Context) (T, error)) (T, error) {
	spin := NewSpinner(message)
	spin.Start()

	// Stop spinner on context cancellation
	go func() {
		<-ctx.Done()
		spin.Stop()
	}()

	result, err := fn(ctx)
	spin.Stop()
	return result, err
}

// ProgressBar provides a text-based progress bar for bulk operations.
// It auto-disables when stderr is not a TTY, when --no-input is set,
// or when output format is JSON/YAML. Use for operations with known
// item counts (import, export, bulk pause/resume).
//
// Usage:
//
//	bar := NewProgressBar(100, "Importing probes...", WithBarDisabled(false))
//	for i := 0; i < 100; i++ {
//	    bar.Increment()
//	}
//	bar.Complete()
type ProgressBar struct {
	total     int
	current   int
	message   string
	width     int
	writer    io.Writer
	disabled  bool
	startTime time.Time
	mu        sync.Mutex
	complete  bool
}

// ProgressBarOption configures a ProgressBar.
type ProgressBarOption func(*ProgressBar)

// WithBarWriter sets the output writer for the progress bar.
func WithBarWriter(w io.Writer) ProgressBarOption {
	return func(p *ProgressBar) {
		p.writer = w
	}
}

// WithBarWidth sets the visual width of the progress bar in characters.
func WithBarWidth(width int) ProgressBarOption {
	return func(p *ProgressBar) {
		if width > 0 {
			p.width = width
		}
	}
}

// WithBarDisabled explicitly sets the progress bar's disabled state.
// When disabled, Increment/Set/Complete produce no output.
func WithBarDisabled(disabled bool) ProgressBarOption {
	return func(p *ProgressBar) {
		p.disabled = disabled
	}
}

// NewProgressBar creates a new progress bar with the given total and message.
// The bar auto-disables when stderr is not a TTY, --no-input is set,
// STACKEYE_NO_INPUT is set, or output format is JSON/YAML.
// Use WithBarDisabled(false) to force-enable regardless of environment.
func NewProgressBar(total int, message string, opts ...ProgressBarOption) *ProgressBar {
	p := &ProgressBar{
		total:    total,
		message:  message,
		width:    30,
		writer:   os.Stderr,
		disabled: !isAnimationEnabled(),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// SetWriter sets the output writer for the progress bar.
// Deprecated: Use WithBarWriter option in NewProgressBar instead.
func (p *ProgressBar) SetWriter(w io.Writer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.writer = w
}

// Disabled returns true if the progress bar is disabled and will not display output.
func (p *ProgressBar) Disabled() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.disabled
}

// Increment advances the progress bar by one step.
// If the bar is disabled, this is a no-op.
func (p *ProgressBar) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.complete || p.disabled {
		return
	}

	if p.startTime.IsZero() {
		p.startTime = time.Now()
	}

	p.current++
	p.render()
}

// Set sets the current progress value.
// If the bar is disabled, this is a no-op.
func (p *ProgressBar) Set(current int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.complete || p.disabled {
		return
	}

	if p.startTime.IsZero() {
		p.startTime = time.Now()
	}

	p.current = current
	p.render()
}

// Complete marks the progress bar as complete and prints a newline.
// If the bar is disabled or total is zero, this is a no-op.
func (p *ProgressBar) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.disabled || p.total <= 0 {
		return
	}

	p.complete = true
	p.current = p.total
	p.render()
	fmt.Fprintln(p.writer)
}

// CompleteWithSuccess marks the bar as complete and prints a success message.
// If the bar is disabled, no output is produced.
func (p *ProgressBar) CompleteWithSuccess(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.disabled {
		return
	}

	p.complete = true
	p.current = p.total
	p.render()
	fmt.Fprintf(p.writer, "\n✓ %s\n", message)
}

// CompleteWithError marks the bar as complete and prints an error message.
// If the bar is disabled, no output is produced.
func (p *ProgressBar) CompleteWithError(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.disabled {
		return
	}

	p.complete = true
	p.render()
	fmt.Fprintf(p.writer, "\n✗ %s\n", message)
}

// render draws the progress bar. Must be called with p.mu held.
func (p *ProgressBar) render() {
	if p.total <= 0 {
		return
	}

	percent := float64(p.current) / float64(p.total)
	filled := int(percent * float64(p.width))
	if filled > p.width {
		filled = p.width
	}

	bar := ""
	for i := 0; i < p.width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	eta := p.estimateRemaining()
	if eta != "" {
		fmt.Fprintf(p.writer, "\r\033[K%s [%s] %d/%d (%.0f%%) %s",
			p.message, bar, p.current, p.total, percent*100, eta)
	} else {
		fmt.Fprintf(p.writer, "\r\033[K%s [%s] %d/%d (%.0f%%)",
			p.message, bar, p.current, p.total, percent*100)
	}
}

// estimateRemaining calculates the ETA based on elapsed time and progress.
// Must be called with p.mu held.
func (p *ProgressBar) estimateRemaining() string {
	if p.current <= 0 || p.startTime.IsZero() || p.complete {
		return ""
	}

	elapsed := time.Since(p.startTime)
	if elapsed < 500*time.Millisecond {
		return ""
	}

	avgPerItem := elapsed / time.Duration(p.current)
	remaining := time.Duration(p.total-p.current) * avgPerItem

	if remaining < time.Second {
		return "ETA: <1s"
	}
	if remaining < time.Minute {
		return fmt.Sprintf("ETA: %ds", int(remaining.Seconds()))
	}
	return fmt.Sprintf("ETA: %dm%ds", int(remaining.Minutes()), int(remaining.Seconds())%60)
}

// RunWithProgressBar executes a function that processes items, displaying a
// progress bar. The callback receives the bar to call Increment() on each item.
//
// Example:
//
//	err := output.RunWithProgressBar(len(probes), "Pausing probes...", func(bar *ProgressBar) error {
//	    for _, probe := range probes {
//	        if err := client.PauseProbe(ctx, probe.ID); err != nil {
//	            return err
//	        }
//	        bar.Increment()
//	    }
//	    return nil
//	})
func RunWithProgressBar(total int, message string, fn func(bar *ProgressBar) error) error {
	bar := NewProgressBar(total, message)
	err := fn(bar)
	if err != nil {
		bar.CompleteWithError(err.Error())
		return err
	}
	bar.Complete()
	return nil
}
