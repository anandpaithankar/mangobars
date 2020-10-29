package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
)

// ExpirationStatus ... Enum type
type ExpirationStatus string

const (
	expired ExpirationStatus = "Expired"
	warn                     = "Warn"
	alert                    = "Alert"
	valid                    = "Valid"
)

// CertificateStatusResult ...
type CertificateStatusResult struct {
	host     string
	port     string
	subject  string
	ca       bool
	days     int
	notAfter time.Time
	status   ExpirationStatus
	err      error
}

// SSLHost ...
type SSLHost struct {
	host string
	port string
}

// SSLConnect ...
type SSLConnect struct {
	warnDays  int
	alertDays int
	wg        *sync.WaitGroup
	wp        *workerpool.WorkerPool
	results   chan<- CertificateStatusResult
}

// NewSSLConnect ...
func NewSSLConnect(warnDays, alertDays int, wg *sync.WaitGroup, wp *workerpool.WorkerPool, r chan<- CertificateStatusResult) *SSLConnect {
	return &SSLConnect{warnDays: warnDays, alertDays: alertDays, wg: wg, wp: wp, results: r}
}

// Connect ... Connects to the SSL host
func (sc *SSLConnect) Connect(h SSLHost) {
	defer sc.wg.Done()
	d := &net.Dialer{
		Timeout: time.Millisecond * time.Duration(2000),
	}

	conn, err := tls.DialWithDialer(d, "tcp", h.host+":"+h.port, &tls.Config{
		InsecureSkipVerify: true, ServerName: h.host})
	if err != nil {
		sc.results <- CertificateStatusResult{host: h.host, port: h.port, err: err}
		return
	}
	defer conn.Close()

	pc := conn.ConnectionState().PeerCertificates
	if len(pc) == 0 {
		sc.results <- CertificateStatusResult{err: fmt.Errorf("no peer certificates received")}
		return
	}

	r := sc.validateCertificate(pc[0])
	r.host = strings.ToLower(h.host)
	r.port = h.port
	sc.results <- r
}

func (sc *SSLConnect) validateCertificate(cert *x509.Certificate) CertificateStatusResult {

	exp := cert.NotAfter
	days := int(math.Ceil(time.Until(exp).Hours() / 24))
	var status ExpirationStatus

	switch {
	case days < warnDays && days > alertDays:
		status = warn
	case days < alertDays && days >= 0:
		status = alert
	case days > warnDays:
		status = valid
	case days < 0:
		fallthrough
	default:
		status = expired
	}

	return CertificateStatusResult{
		subject:  cert.Subject.CommonName,
		notAfter: cert.NotAfter,
		ca:       cert.IsCA,
		err:      nil,
		status:   status,
		days:     days,
	}
}
