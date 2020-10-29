package main

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/pterm/pterm"
)

// ConsoleWriter ...
type ConsoleWriter struct {
	wg  *sync.WaitGroup
	twc chan CertificateStatusResult
}

// NewConsoleWriter ... Creates a new console writer
func NewConsoleWriter(wg *sync.WaitGroup, twc chan CertificateStatusResult) *ConsoleWriter {
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

func (t *ConsoleWriter) colorWriter(r CertificateStatusResult) {
	defer t.wg.Done()
	if r.err != nil {
		s := fmt.Sprintf("%s:%s (%s)", r.host, r.port, r.err.Error())
		pfx := pterm.Error.Prefix
		pfx.Text = "  ERROR  "
		pterm.Error.WithPrefix(pfx).Println(s)
		return

	}

	s := fmt.Sprintf("%s:%s (%s | %s days | %s)", r.host, r.port, r.subject, strconv.Itoa(r.days), r.notAfter.String())

	switch r.status {
	case valid:
		t.printValid(s)
	case expired:
		t.printExpired(s)
	case warn:
		t.printWarn(s)
	case alert:
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
	pfx.Text = "EXPRIRED "
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
