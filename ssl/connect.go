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

	conn, err := tls.DialWithDialer(d, "tcp", h.Host+":"+h.Port, &tls.Config{
		InsecureSkipVerify: true, ServerName: h.Host})
	if err != nil {
		sc.results <- CertificateStatusResult{Host: h.Host, Port: h.Port, Err: err}
		return
	}
	defer conn.Close()

	pc := conn.ConnectionState().PeerCertificates
	if len(pc) == 0 {
		sc.results <- CertificateStatusResult{Err: fmt.Errorf("no peer certificates received")}
		return
	}

	r := sc.validateCertificate(pc[0])
	r.Host = strings.ToLower(h.Host)
	r.Port = h.Port
	sc.results <- r
}

func (sc *SSLConnect) validateCertificate(cert *x509.Certificate) CertificateStatusResult {

	exp := cert.NotAfter
	days := int(math.Ceil(time.Until(exp).Hours() / 24))
	var status ExpirationStatus

	switch {
	case days < sc.warnDays && days > sc.alertDays:
		status = Warn
	case days < sc.alertDays && days >= 0:
		status = Alert
	case days > sc.warnDays:
		status = Valid
	case days < 0:
		fallthrough
	default:
		status = Expired
	}

	return CertificateStatusResult{
		Subject:  cert.Subject.CommonName,
		NotAfter: cert.NotAfter,
		Ca:       cert.IsCA,
		Err:      nil,
		Status:   status,
		Days:     days,
	}
}
