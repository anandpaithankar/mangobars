package ssl

import (
	"sync"
	"time"

	"github.com/gammazero/workerpool"
)

// ExpirationStatus ... Enum type
type ExpirationStatus string

const (
	Expired ExpirationStatus = "Expired"
	Warn                     = "Warn"
	Alert                    = "Alert"
	Valid                    = "Valid"
)

// CertificateStatusResult ...
type CertificateStatusResult struct {
	Host     string
	Port     string
	Subject  string
	Ca       bool
	Days     int
	NotAfter time.Time
	Status   ExpirationStatus
	Err      error
}

// SSLHost ...
type SSLHost struct {
	Host string
	Port string
}

// SSLConnect ...
type SSLConnect struct {
	warnDays  int
	alertDays int
	wg        *sync.WaitGroup
	wp        *workerpool.WorkerPool
	results   chan<- CertificateStatusResult
}
