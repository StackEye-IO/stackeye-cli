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

// Spinner provides a simple text-based spinner for long-running operations.
// It displays a spinning animation with an optional message to indicate progress.
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

// DefaultSpinnerFrames provides a simple ASCII spinner animation.
var DefaultSpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ASCIISpinnerFrames provides an ASCII-only spinner for terminals without Unicode.
var ASCIISpinnerFrames = []string{"|", "/", "-", "\\"}

// NewSpinner creates a new spinner with the given message.
func NewSpinner(message string, opts ...SpinnerOption) *Spinner {
	s := &Spinner{
		message:  message,
		writer:   os.Stderr,
		frames:   DefaultSpinnerFrames,
		interval: 80 * time.Millisecond,
		done:     make(chan struct{}),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Start begins the spinner animation.
// The spinner runs in a separate goroutine until Stop is called.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
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
func (s *Spinner) StopWithMessage(message string) {
	s.Stop()
	s.mu.Lock()
	writer := s.writer
	s.mu.Unlock()
	fmt.Fprintln(writer, message)
}

// StopWithSuccess stops the spinner with a success indicator.
func (s *Spinner) StopWithSuccess(message string) {
	s.Stop()
	s.mu.Lock()
	writer := s.writer
	s.mu.Unlock()
	fmt.Fprintf(writer, "✓ %s\n", message)
}

// StopWithError stops the spinner with an error indicator.
func (s *Spinner) StopWithError(message string) {
	s.Stop()
	s.mu.Lock()
	writer := s.writer
	s.mu.Unlock()
	fmt.Fprintf(writer, "✗ %s\n", message)
}

// UpdateMessage changes the spinner message while running.
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
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

// ProgressBar provides a simple text-based progress bar.
// Use this for operations with known progress (like batch operations).
type ProgressBar struct {
	total    int
	current  int
	message  string
	width    int
	writer   io.Writer
	mu       sync.Mutex
	complete bool
}

// NewProgressBar creates a new progress bar.
func NewProgressBar(total int, message string) *ProgressBar {
	return &ProgressBar{
		total:   total,
		message: message,
		width:   30,
		writer:  os.Stderr,
	}
}

// SetWriter sets the output writer for the progress bar.
func (p *ProgressBar) SetWriter(w io.Writer) {
	p.writer = w
}

// Increment advances the progress bar by one step.
func (p *ProgressBar) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.complete {
		return
	}

	p.current++
	p.render()
}

// Set sets the current progress value.
func (p *ProgressBar) Set(current int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.complete {
		return
	}

	p.current = current
	p.render()
}

// Complete marks the progress bar as complete.
func (p *ProgressBar) Complete() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.complete = true
	p.current = p.total
	p.render()
	fmt.Fprintln(p.writer)
}

// render draws the progress bar.
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

	fmt.Fprintf(p.writer, "\r%s [%s] %d/%d (%.0f%%)",
		p.message, bar, p.current, p.total, percent*100)
}
