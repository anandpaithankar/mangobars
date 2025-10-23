package writer

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/anandpaithankar/mangobars/ssl"
)

// FileWriter ...
type FileWriter struct {
	name string
	f    *os.File
	wg   *sync.WaitGroup
	fwc  chan ssl.CertificateStatusResult
}

// NewFileWriter ... Creates a new file writer
func NewFileWriter(wg *sync.WaitGroup, fwc chan ssl.CertificateStatusResult, name string) (*FileWriter, func()) {
	f, err := os.Create(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating a %s file", name)
		return nil, func() {}
	}

	fw := &FileWriter{name: name, f: f, fwc: fwc, wg: wg}
	go fw.run()
	return fw, fw.release
}

// Close ... Close the file
func (fw *FileWriter) release() {
	if fw.f != nil {
		fw.f.Close()
	}
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

func (fw *FileWriter) write(r ssl.CertificateStatusResult) {
	defer fw.wg.Done()
	var s string
	if r.Err != nil {
		s = fmt.Sprintf("%s,%s,ERROR,%s\n", r.Host, r.Port, r.Err.Error())
	} else {
		tlsVersion := r.TLSVersion
		if tlsVersion == "" {
			tlsVersion = "Unknown"
		}
		s = fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s\n",
			r.Host, r.Port, r.Subject, string(r.Status),
			strconv.Itoa(r.Days), tlsVersion, r.NotAfter.Format("2006-01-02 15:04:05 UTC"))
	}
	_, err := fw.f.WriteString(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
	}
}
