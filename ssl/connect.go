package ssl

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

// NewSSLConnect creates a new SSL connector with configurable timeout
func NewSSLConnect(warnDays, alertDays int, timeout time.Duration, wg *sync.WaitGroup, wp *workerpool.WorkerPool, r chan<- CertificateStatusResult) *SSLConnect {
	return &SSLConnect{
		warnDays: warnDays,
		alertDays: alertDays,
		timeout: timeout,
		wg: wg,
		wp: wp,
		results: r,
		dialerPool: &sync.Pool{
			New: func() interface{} {
				return &net.Dialer{
					Timeout:   timeout,
					KeepAlive: 30 * time.Second,
				}
			},
		},
	}
}

// Connect establishes SSL connection and validates certificate
func (sc *SSLConnect) Connect(h SSLHost) {
	defer sc.wg.Done()

	// Get dialer from pool for better resource management
	d := sc.dialerPool.Get().(*net.Dialer)
	defer sc.dialerPool.Put(d)

	// Create TLS config with better security settings
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // We're only checking certificate expiration
		ServerName:         h.Host,
		MinVersion:         tls.VersionTLS10, // Support older servers for monitoring
	}

	address := net.JoinHostPort(h.Host, h.Port)
	conn, err := tls.DialWithDialer(d, "tcp", address, tlsConfig)
	if err != nil {
		sc.results <- CertificateStatusResult{
			Host: strings.ToLower(h.Host),
			Port: h.Port,
			Err:  fmt.Errorf("connection failed: %w", err),
		}
		return
	}
	defer conn.Close()

	// Get connection state
	state := conn.ConnectionState()
	pc := state.PeerCertificates
	if len(pc) == 0 {
		sc.results <- CertificateStatusResult{
			Host: strings.ToLower(h.Host),
			Port: h.Port,
			Err:  fmt.Errorf("no peer certificates received"),
		}
		return
	}

	// Validate the leaf certificate
	r := sc.validateCertificate(pc[0])
	r.Host = strings.ToLower(h.Host)
	r.Port = h.Port
	r.TLSVersion = sc.getTLSVersion(state.Version)
	sc.results <- r
}

// getTLSVersion converts TLS version constant to readable string
func (sc *SSLConnect) getTLSVersion(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}

func (sc *SSLConnect) validateCertificate(cert *x509.Certificate) CertificateStatusResult {
	exp := cert.NotAfter
	now := time.Now()
	duration := exp.Sub(now)
	days := int(math.Ceil(duration.Hours() / 24))

	var status ExpirationStatus
	switch {
	case days < 0:
		status = Expired
	case days <= sc.alertDays:
		status = Alert
	case days <= sc.warnDays:
		status = Warn
	default:
		status = Valid
	}

	// Get the best subject name (prefer SAN over CN)
	subject := cert.Subject.CommonName
	if len(cert.DNSNames) > 0 {
		subject = cert.DNSNames[0] // Use first SAN entry
	}

	return CertificateStatusResult{
		Subject:  subject,
		NotAfter: cert.NotAfter,
		Ca:       cert.IsCA,
		Err:      nil,
		Status:   status,
		Days:     days,
	}
}
