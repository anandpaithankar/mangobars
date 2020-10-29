package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
)

// FileWriter ...
type FileWriter struct {
	name string
	f    *os.File
	wg   *sync.WaitGroup
	fwc  chan CertificateStatusResult
}

// NewFileWriter ... Creates a new file writer
func NewFileWriter(wg *sync.WaitGroup, fwc chan CertificateStatusResult, name string) *FileWriter {
	f, err := os.Create(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating a %s file", name)
		return nil
	}
	fw := &FileWriter{name: name, f: f, fwc: fwc, wg: wg}
	go fw.run()
	return fw
}

func (fw *FileWriter) run() {
	for {
		select {
		case r, ok := <-fw.fwc:
			if !ok {
				fw.f.Close()
				return
			}
			fw.wg.Add(1)
			fw.write(r)
		default:
			// TODO: Implement cancellation,context
		}
	}
}

func (fw *FileWriter) write(r CertificateStatusResult) {
	defer fw.wg.Done()
	var s string
	if r.err != nil {
		s = fmt.Sprintf("%s,%s,%s\n", r.host, r.port, r.err.Error())

	} else {
		s = fmt.Sprintf("%s,%s,%s,%s,%s,%s\n", r.host, r.port, r.subject, string(r.status), strconv.Itoa(r.days), r.notAfter.String())
	}
	_, err := fw.f.WriteString(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %v", err)
	}
}
