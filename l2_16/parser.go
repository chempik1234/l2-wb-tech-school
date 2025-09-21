package main

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// IParser is a universal interface for parsers
type IParser interface {
	// ParseResponse (if html) parses response for links and replaces them with local filenames, returning the links and error
	//
	// if HTML: return processed file as []byte, []string of recursive links, .fileFormat, err
	//
	// else: return nil, nil, file format, err
	ParseResponse(urlFrom *url.URL, resp *http.Response) ([]byte, []string, string, error)
}

// ParserToLocalFiles returns a file name to save locally instead of file format
type ParserToLocalFiles struct {
	localFileDirectory string
}

// NewParserToLocalFiles creates a new NewParserToLocalFiles that creates file names for a given directory
func NewParserToLocalFiles(localFileDirectory string) *ParserToLocalFiles {
	return &ParserToLocalFiles{localFileDirectory: localFileDirectory}
}

// ParseResponse in ParserToLocalFiles returns full filePath instead of just .format
func (p *ParserToLocalFiles) ParseResponse(urlFrom *url.URL, resp *http.Response) ([]byte, []string, string, error) {
	// file format
	// fileFormat := GetFileFormat(resp.Header.Get("Content-Type"), "txt")

	// raw file data (step 1)

	//region open body
	fileData, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, nil, "", fmt.Errorf("error reading body: %w", err)
	}

	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			fmt.Println("error closing body")
		}
	}(resp.Body)
	//endregion

	// we have to read again, so we create new reader based on already read data
	bodyReader := io.NopCloser(bytes.NewReader(fileData))

	var subLinksStrings []string
	subLinksStrings, err = GetLinks(bodyReader)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to get links: %w", err)
	}

	//region to absolute and local links
	subLinks := make([]string, 0, len(subLinksStrings))

	// predict how pages will be named as files and make links to them
	subLinksToLocalMap := make(map[string]string, len(subLinksStrings))

	// get only absolute
	basePath := fmt.Sprintf("%s://%s", urlFrom.Scheme, urlFrom.Host)

	var newLink string
	for _, link := range subLinksStrings {
		if strings.HasPrefix(link, "/") {
			newLink = basePath + link

			subLinks = append(subLinks, newLink)
			subLinksToLocalMap[link] = URLToFilePath(p.localFileDirectory, newLink)

		} else if !strings.HasPrefix(link, "#") {

			// check if we're not copying 3rd party links
			if sameHost, hostErr := IsSameDomain(urlFrom.Host, link); !sameHost || hostErr != nil {
				continue
			}

			subLinks = append(subLinks, link)
			subLinksToLocalMap[link] = URLToFilePath(p.localFileDirectory, link)
		}
	}
	//endregion

	//region replace links
	fileData = replaceLinks(fileData, subLinksToLocalMap)
	//endregion

	fileName := URLToFilePath(p.localFileDirectory, urlFrom.String()) //, fileFormat)

	return fileData, subLinks, fileName, nil
}

// replaceLinks in html according to linksMap: url -> some text
func replaceLinks(fileData []byte, linksMap map[string]string) []byte {
	// Convert byte slice to string for processing
	content := string(fileData)

	// Replace each URL with its local file path
	for rawURL, localPath := range linksMap {
		textToReplace := fmt.Sprintf("\"%s\"", rawURL)
		textToReplaceWith := fmt.Sprintf("\"%s\"", localPath)

		// Replace all occurrences of the URL with the local path
		content = strings.ReplaceAll(content, textToReplace, textToReplaceWith)

		// Also replace HTML-encoded versions of the URL
		encodedURL := html.EscapeString(rawURL)
		if encodedURL != rawURL {
			content = strings.ReplaceAll(content, fmt.Sprintf("\"%s\"", encodedURL), textToReplaceWith)
		}
	}

	return []byte(content)
}
