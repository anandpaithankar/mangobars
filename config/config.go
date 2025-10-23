package config

import (
	"runtime"
	"time"
)

// Config holds all configuration options for mangobars
type Config struct {
	// Input/Output
	InputFile  string
	ResultFile string

	// Target host (single host mode)
	TargetHost string
	TargetPort string

	// Certificate validation thresholds
	WarnDays  int
	AlertDays int

	// Performance settings
	MaxWorkers int
	Timeout    time.Duration
	BatchSize  int

	// Display options
	ShowProgress bool
	Verbose      bool
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		InputFile:    "host.csv",
		ResultFile:   "result.csv",
		TargetPort:   "443",
		WarnDays:     20,
		AlertDays:    10,
		MaxWorkers:   calculateOptimalWorkers(),
		Timeout:      3 * time.Second,
		BatchSize:    100,
		ShowProgress: true,
		Verbose:      false,
	}
}

// calculateOptimalWorkers determines the optimal number of workers based on system resources
func calculateOptimalWorkers() int {
	numCPU := runtime.NumCPU()
	optimal := numCPU * 2
	if optimal < 5 {
		return 5
	}
	if optimal > 50 {
		return 50
	}
	return optimal
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.MaxWorkers < 1 {
		c.MaxWorkers = 1
	}
	if c.MaxWorkers > 100 {
		c.MaxWorkers = 100
	}
	if c.Timeout < time.Millisecond*100 {
		c.Timeout = time.Millisecond * 100
	}
	if c.BatchSize < 1 {
		c.BatchSize = 1
	}
	return nil
}