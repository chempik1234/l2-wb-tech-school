package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
)

// Downloader does all the job in this module
type Downloader struct {
	// savedLinks store all saved urls so same urls aren't downloaded twice
	savedLinks map[string]struct{}
	// mu is the mutex for savedLinks safety
	mu *sync.Mutex
	// client is for making requests
	client *http.Client
	// retries is the number that describes tries count per any operation that can be retried
	retries int

	// parser is the Parser used for processing input html data, can be replaced
	parser IParser
}

// NewDownloader creates a new *Downloader with given client and retries and empty savedLinks cache
func NewDownloader(client *http.Client, retries int, parser *Parser) *Downloader {
	return &Downloader{
		client:     client,
		savedLinks: make(map[string]struct{}, 0),
		retries:    retries,
		mu:         &sync.Mutex{},
		parser:     parser,
	}
}

// download downloads page by url, saves as {pageNum}.html
//
// depthLeft goes like 10, 9, 8, .., 0
//
// where 10 is input depth (-d 10), with every level it goes down
//
// if depthLeft - 1 == 0 then return, else continue
func (d *Downloader) download(urlToDownload *url.URL, saveDirectory string, depthLeft int) error {
	urlString := urlToDownload.String()
	// we should not lock it for too long
	d.mu.Lock()
	if _, ok := d.savedLinks[urlString]; ok {
		// early unlock
		d.mu.Unlock()
		return nil
	}
	d.savedLinks[urlString] = struct{}{}

	// early unlock
	d.mu.Unlock()

	var resp *http.Response
	var err error
	for try := 0; try < d.retries; try++ {
		resp, err = d.client.Get(urlString)
		if errors.Is(err, nil) {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("failed to download %s (%d tries): %s", urlToDownload, d.retries, err)
	}

	//region parse
	var parsedFileData []byte
	var parsedSubLinks []string
	var fileName string
	parsedFileData, parsedSubLinks, fileName, err = d.parser.ParseResponse(urlToDownload, resp)

	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fileName = fmt.Sprintf("%s%c%s", saveDirectory, os.PathSeparator, fileName)
	//endregion

	err = d.save(parsedFileData, fileName)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	//region recursive

	// early return if max recursive reached
	if depthLeft <= 0 {
		return nil
	}

	wg2 := new(sync.WaitGroup)
	wg2.Add(len(parsedSubLinks))

	for _, subLink := range parsedSubLinks {
		go func(wg2 *sync.WaitGroup, subLink string, saveDirectory string, depthLeft int) {
			defer wg2.Done()

			var subURL *url.URL

			subURL, err = url.Parse(subLink)

			if err != nil {
				fmt.Println(fmt.Errorf("failed to parse sub link %s: %w", subLink, err))
			}

			err = d.download(subURL, saveDirectory, depthLeft)

			if err != nil {
				fmt.Println(fmt.Errorf("failed to fetch sub link %s: %w", subLink, err))
			}
		}(wg2, subLink, saveDirectory, depthLeft-1)
	}

	wg2.Wait()
	//endregion

	return nil
}

// save creates and saves file with given []byte and filepath
func (d *Downloader) save(fileData []byte, filePath string) error {
	var err error

	var file *os.File

	for try := 0; try < d.retries; try++ {
		file, err = os.Create(filePath)
		if errors.Is(err, nil) {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("failed to create %s (%d tries): %s", filePath, d.retries, err)
	}

	for try := 0; try < d.retries; try++ {
		_, err = file.Write(fileData)
		if errors.Is(err, nil) {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("failed to write to %s (%d tries): %s", filePath, d.retries, err)
	}

	return nil
}

// Start calls a root Downloader.download, returns error if root failed, waits for all recursive operations to complete
func (d *Downloader) Start(urlToDownload *url.URL, saveDirectory string, depthLeft int) error {
	err := d.download(urlToDownload, saveDirectory, depthLeft)
	if err != nil {
		return fmt.Errorf("failed to download page: %w", err)
	}
	return nil
}
