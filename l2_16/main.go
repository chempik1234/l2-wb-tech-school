package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	dFlag := flag.Int("d", 0, "depth (default 0 for only root page)")
	tFlag := flag.Int("t", 10, "timeout seconds per page (default 10)")
	rFlag := flag.Int("r", 1, "max tries per page (default 1)")
	wFlag := flag.Uint("w", 10, "max worker pool (default 10)")
	flag.Parse()

	depth := *dFlag
	timeoutSeconds := *tFlag
	retries := *rFlag
	maxWorkers := uint32(*wFlag)

	if timeoutSeconds <= 0 {
		log.Fatal("timeout seconds must be greater than zero")
	}

	timeout := time.Duration(timeoutSeconds) * time.Second

	args := flag.Args()
	if len(args) != 1 {
		log.Fatalln("there must only be 1 argument: url, got", len(args))
	}

	urlString := args[0]
	urlParsed, err := url.Parse(urlString)
	if err != nil {
		log.Fatalln(fmt.Errorf("invalid url: %s: %w", urlString, err))
	}

	var fileDir string
	fileDir, err = os.Getwd()
	if err != nil {
		log.Fatalln(fmt.Errorf("failedd to get working directory: %w", err))
	}

	downloaderObj := NewDownloader(
		&http.Client{
			Timeout: timeout,
		},
		retries,
		NewParserToLocalFiles(fileDir),
		NewPoolWorkerLocker(maxWorkers),
	)

	err = downloaderObj.Start(urlParsed, fileDir, depth)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to download page: %w", err))
	}
}
