package benchmark

import (
	"fmt"
	"time"
)

// Benchmark tracks performance metrics
type Benchmark struct {
	startTime    time.Time
	endTime      time.Time
	totalHosts   int
	successCount int
	errorCount   int
}

// NewBenchmark creates a new benchmark tracker
func NewBenchmark(totalHosts int) *Benchmark {
	return &Benchmark{
		startTime:  time.Now(),
		totalHosts: totalHosts,
	}
}

// RecordSuccess increments the success counter
func (b *Benchmark) RecordSuccess() {
	b.successCount++
}

// RecordError increments the error counter
func (b *Benchmark) RecordError() {
	b.errorCount++
}

// Finish marks the benchmark as complete and returns performance stats
func (b *Benchmark) Finish() string {
	b.endTime = time.Now()
	duration := b.endTime.Sub(b.startTime)

	hostsPerSecond := float64(b.totalHosts) / duration.Seconds()

	return fmt.Sprintf(
		"Performance Summary:\n"+
		"  Total hosts: %d\n"+
		"  Successful: %d\n"+
		"  Errors: %d\n"+
		"  Duration: %v\n"+
		"  Rate: %.1f hosts/second",
		b.totalHosts, b.successCount, b.errorCount,
		duration.Round(time.Millisecond), hostsPerSecond)
}