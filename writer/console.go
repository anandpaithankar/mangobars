package writer

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/anandpaithankar/mangobars/ssl"
	"github.com/pterm/pterm"
)

// ConsoleWriter ...
type ConsoleWriter struct {
	wg  *sync.WaitGroup
	twc chan ssl.CertificateStatusResult
}

// NewConsoleWriter ... Creates a new console writer
func NewConsoleWriter(wg *sync.WaitGroup, twc chan ssl.CertificateStatusResult) *ConsoleWriter {
	this := &ConsoleWriter{twc: twc, wg: wg}
	go this.run()
	return this
}

func (t *ConsoleWriter) run() {
	for {
		select {
		case r, ok := <-t.twc:
			if !ok {
				return
			}
			t.wg.Add(1)
			t.colorWriter(r)
		default:
			// TODO: Implement cancellation,context
		}
	}
}

func (t *ConsoleWriter) colorWriter(r ssl.CertificateStatusResult) {
	defer t.wg.Done()
	if r.Err != nil {
		s := fmt.Sprintf("%s:%s (%s)", r.Host, r.Port, r.Err.Error())
		pfx := pterm.Error.Prefix
		pfx.Text = "  ERROR  "
		pterm.Error.WithPrefix(pfx).Println(s)
		return
	}

	// Enhanced output with TLS version
	tlsInfo := ""
	if r.TLSVersion != "" {
		tlsInfo = fmt.Sprintf(" | %s", r.TLSVersion)
	}

	s := fmt.Sprintf("%s:%s (%s | %s days%s | %s)",
		r.Host, r.Port, r.Subject, strconv.Itoa(r.Days), tlsInfo, r.NotAfter.Format("2006-01-02 15:04:05 UTC"))

	switch r.Status {
	case ssl.Valid:
		t.printValid(s)
	case ssl.Expired:
		t.printExpired(s)
	case ssl.Warn:
		t.printWarn(s)
	case ssl.Alert:
		t.printAlert(s)
	}
}

func (t *ConsoleWriter) printValid(s string) {
	pfx := pterm.Prefix{
		Text:  "  VALID  ",
		Style: pterm.Success.Prefix.Style,
	}

	pterm.Success.WithPrefix(pfx).Println(s)
}

func (t *ConsoleWriter) printExpired(s string) {
	pfx := pterm.Error.Prefix
	pfx.Text = " EXPIRED "
	pterm.Error.WithPrefix(pfx).Println(s)
}

func (t *ConsoleWriter) printWarn(s string) {
	pfx := pterm.Warning.Prefix
	pfx.Text = "  WARN   "
	pterm.Warning.WithPrefix(pfx).Println(s)
}

func (t *ConsoleWriter) printAlert(s string) {
	pfx := pterm.Prefix{
		Text:  "  ALERT  ",
		Style: pterm.NewStyle(pterm.FgLightRed, pterm.BgDarkGray, pterm.FastBlink),
	}
	pterm.Error.WithPrefix(pfx).Println(s)
}
