package main

/*
Usage: go run main.go -f path_to_csv_file -t 200 -m 100000000
*/

import (
	"flag"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/Rakanixu/bulk_status/stat"
)

const (
	COLUMN_HTTP_RESOURCE   = "Resource URL"
	COLUMN_IS_ERROR        = "Error Code"
	COLUMN_ERR_DESCRIPTION = "Error Description"
	NO_ERROR_VALUE         = "OK"
)

var stats *stat.Stat

func main() {
	f := flag.String("f", "data.csv", "CSV File")
	n := flag.Int("t", 100, "Number of go routines will be spawn")
	bs := flag.Int64("m", 1000000000, "CSV file size (bytes) default to 1000MB")

	flag.Parse()

	if *f == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	s, err := os.Open(*f)
	if err != nil {
		log.Fatal(err)
	}

	// Increase size if CSV file is > 1000MB
	b := make([]byte, *bs)
	count, err := s.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	records := strings.Split(string(b[:count]), "\n")
	columms := strings.Split(records[0], ",")

	// Get useful columns positions in CSV
	var resourceIndex, errIndex, errDescIndex int
	for k, v := range columms {
		switch v {
		case COLUMN_HTTP_RESOURCE:
			resourceIndex = k
		case COLUMN_IS_ERROR:
			errIndex = k
		case COLUMN_ERR_DESCRIPTION:
			errDescIndex = k
		}
	}

	// Spawn goroutines to handle requests in parallel
	var wg sync.WaitGroup
	c := make(chan stat.CsvData)
	stats := stat.NewStat()

	wg.Add(1)
	for i := 0; i < *n; i++ {
		go stats.QueryResource(c)
	}

	// Range over CSV values
	for _, v := range records[1:] {
		r := strings.Split(v, ",")

		if r[errIndex] != NO_ERROR_VALUE {
			// Avoid acccess to unloccated memory
			if len(r) > errIndex && len(r) > resourceIndex && len(r) > errDescIndex {
				// Push to channel
				c <- stat.CsvData{
					Url:            r[resourceIndex],
					ErrCode:        r[errIndex],
					ErrDescription: r[errDescIndex],
				}
			}
		}
	}
	wg.Done()

	stats.Info()
}
