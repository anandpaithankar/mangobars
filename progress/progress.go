package progress

import (
	"fmt"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

// ProgressTracker tracks the progress of SSL certificate checks
type ProgressTracker struct {
	total     int
	completed int
	mu        sync.Mutex
	spinner   *pterm.SpinnerPrinter
	startTime time.Time
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(total int) *ProgressTracker {
	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Checking %d certificates...", total))
	return &ProgressTracker{
		total:     total,
		completed: 0,
		spinner:   spinner,
		startTime: time.Now(),
	}
}

// Increment increments the completed count and updates the display
func (pt *ProgressTracker) Increment() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.completed++
	elapsed := time.Since(pt.startTime)

	if pt.completed == pt.total {
		pt.spinner.Success(fmt.Sprintf("Completed %d certificate checks in %v", pt.total, elapsed.Round(time.Millisecond*10)))
	} else {
		percentage := float64(pt.completed) / float64(pt.total) * 100
		pt.spinner.UpdateText(fmt.Sprintf("Checking certificates... %d/%d (%.1f%%) - %v elapsed",
			pt.completed, pt.total, percentage, elapsed.Round(time.Second)))
	}
}

// Stop stops the progress tracker
func (pt *ProgressTracker) Stop() {
	if pt.spinner != nil {
		pt.spinner.Stop()
	}
}