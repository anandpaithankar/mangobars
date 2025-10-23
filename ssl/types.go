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

// CertificateStatusResult contains the result of certificate validation
type CertificateStatusResult struct {
	Host       string
	Port       string
	Subject    string
	Ca         bool
	Days       int
	NotAfter   time.Time
	Status     ExpirationStatus
	TLSVersion string
	Err        error
}

// SSLHost ...
type SSLHost struct {
	Host string
	Port string
}

// SSLConnect handles SSL certificate validation with connection pooling
type SSLConnect struct {
	warnDays   int
	alertDays  int
	timeout    time.Duration
	wg         *sync.WaitGroup
	wp         *workerpool.WorkerPool
	results    chan<- CertificateStatusResult
	dialerPool *sync.Pool
}
