package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"
	"time"
)

func main() {
	dFlag := flag.Int("d", 0, "depth (default 0 for only root page)")
	tFlag := flag.Int("t", 10, "timeout seconds per page (default 10)")
	rFlag := flag.Int("r", 1, "max tries per page (default 1)")
	flag.Parse()

	depth := *dFlag
	timeoutSeconds := *tFlag
	retries := *rFlag

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

	pageNumber := &atomic.Int32{}

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
	)

	err = downloaderObj.Start(urlParsed, pageNumber, fileDir, depth)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to download page: %w", err))
	}
}
