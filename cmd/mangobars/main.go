package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

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

const maxSSLWorkers = 10

var (
	inputFile  string
	resultFile string
	targetHost string
	targetPort string
	warnDays   int
	alertDays  int
	cw         *ConsoleWriter
	fw         *FileWriter
)

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
	cwc := make(chan CertificateStatusResult)
	fwc := make(chan CertificateStatusResult)
	var releaseFile func()

	cw = NewConsoleWriter(&wg, cwc)
	if len(targetHost) == 0 {
		fw, releaseFile = NewFileWriter(&wg, fwc, resultFile)
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
		fmt.Println(usageString)
		flag.PrintDefaults()
	}

	flag.StringVar(&targetHost, "h", "", "Hostname with or without port")
	flag.StringVar(&targetPort, "p", "443", "Port")
	flag.IntVar(&warnDays, "w", 20, "Warn if the expiration falls within specified days.")
	flag.IntVar(&alertDays, "a", 10, "Alert if the expiration falls within specified days.")
	flag.StringVar(&inputFile, "i", "host.csv", "CSV file containing host information.")
	flag.StringVar(&resultFile, "o", "result.csv", "Result from the scan.")
	flag.Parse()
}

func process() (r chan CertificateStatusResult, e error) {
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

func dispatch(reader io.Reader) (chan CertificateStatusResult, error) {
	results := make(chan CertificateStatusResult)
	wp := workerpool.New(maxSSLWorkers)
	var wg sync.WaitGroup
	sc := NewSSLConnect(warnDays, alertDays, &wg, wp, results)

	cleanup := func() {
		wg.Wait()
		wp.Stop()
		close(results)
	}

	r := csv.NewReader(reader)
	r.Comma = ','
	r.Comment = '#'
	for {
		entry, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			cleanup()
			return nil, err
		}
		task := SSLHost{
			host: entry[0],
			port: entry[1],
		}
		wg.Add(1)
		wp.Submit(func() {
			sc.Connect(task)
		})
	}

	go cleanup()
	return results, nil
}
