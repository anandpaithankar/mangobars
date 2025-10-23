package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/anandpaithankar/mangobars/ssl"
	"github.com/anandpaithankar/mangobars/writer"
	"github.com/gammazero/workerpool"
)

const usageString = `Usage: mangobars [OPTION] [FILEPATH]
Checks the expiration status for Server certificates.
Example:
	mangobars -w 20 -a 10 -i host.csv -o result.csv
	mangobars -h amazon.com -p 443
	mangobars -h amazon.com
	mangobars -h amazon.com:443
`

var (
	inputFile     string
	resultFile    string
	targetHost    string
	targetPort    string
	warnDays      int
	alertDays     int
	maxWorkers    int
	timeout       int
	batchSize     int
	cw            *writer.ConsoleWriter
	fw            *writer.FileWriter
)

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

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
}

func start() error {
	usage()

	res, err := process()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	cwc := make(chan ssl.CertificateStatusResult)
	fwc := make(chan ssl.CertificateStatusResult)
	var releaseFile func()

	cw = writer.NewConsoleWriter(&wg, cwc)
	if len(targetHost) == 0 {
		fw, releaseFile = writer.NewFileWriter(&wg, fwc, resultFile)
		defer releaseFile()
	}

	for r := range res {
		cwc <- r
		if len(targetHost) == 0 {
			fwc <- r
		}
	}

	close(cwc)
	close(fwc)
	wg.Wait()
	return nil
}

func usage() {
	flag.Usage = func() {
		fmt.Printf(usageString)
		flag.PrintDefaults()
	}

	flag.StringVar(&targetHost, "h", "", "Hostname with or without port. Input file specified with `-i` will be ignored.")
	flag.StringVar(&targetPort, "p", "443", "Port")
	flag.IntVar(&warnDays, "w", 20, "Warn if the certificate expiration is less than specified days but has enough time not to be alerted.")
	flag.IntVar(&alertDays, "a", 10, "Alert if the certificate expiration is less than specified days.")
	flag.StringVar(&inputFile, "i", "host.csv", "CSV file containing host information.")
	flag.StringVar(&resultFile, "o", "result.csv", "Output file name.")
	flag.IntVar(&maxWorkers, "workers", calculateOptimalWorkers(), "Maximum number of concurrent workers")
	flag.IntVar(&timeout, "timeout", 3000, "Connection timeout in milliseconds")
	flag.IntVar(&batchSize, "batch", 100, "Batch size for processing large files")
	flag.Parse()
}

func process() (r chan ssl.CertificateStatusResult, e error) {
	var reader io.Reader
	var f *os.File
	if len(targetHost) != 0 {
		var entry string
		if strings.Contains(targetHost, ":") {
			values := strings.Split(targetHost, ":")
			if len(values) > 2 {
				return nil, fmt.Errorf("invalid hostname: %s", targetHost)
			}
			entry = fmt.Sprintf("%s,%s", values[0], values[1])
		} else {
			entry = fmt.Sprintf("%s,%s", targetHost, targetPort)
		}
		reader = strings.NewReader(entry)
	} else {
		var err error
		f, err = os.Open(inputFile)
		if err != nil {
			return nil, err
		}
		reader = f
		defer f.Close()
	}
	return dispatch(reader)
}

func dispatch(reader io.Reader) (chan ssl.CertificateStatusResult, error) {
	results := make(chan ssl.CertificateStatusResult, batchSize)
	wp := workerpool.New(maxWorkers)
	var wg sync.WaitGroup

	// Create SSL connector with timeout configuration
	timeoutDuration := time.Duration(timeout) * time.Millisecond
	sc := ssl.NewSSLConnect(warnDays, alertDays, timeoutDuration, &wg, wp, results)

	cleanup := func() {
		wg.Wait()
		wp.Stop()
		close(results)
	}

	r := csv.NewReader(reader)
	r.Comma = ','
	r.Comment = '#'
	r.FieldsPerRecord = -1 // Allow variable number of fields per record

	for {
		entry, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			cleanup()
			return nil, fmt.Errorf("CSV parsing error: %w", err)
		}

		if len(entry) == 0 {
			continue // Skip empty lines
		}

		// Derive port with better error handling
		derivePort := func(record []string) string {
			if len(record) >= 2 && len(strings.TrimSpace(record[1])) > 0 {
				return strings.TrimSpace(record[1])
			}
			return "443"
		}

		task := ssl.SSLHost{
			Host: strings.TrimSpace(entry[0]),
			Port: derivePort(entry),
		}

		// Skip empty hosts
		if task.Host == "" {
			continue
		}

		wg.Add(1)
		wp.Submit(func() {
			sc.Connect(task)
		})
	}

	go cleanup()
	return results, nil
}

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
}
