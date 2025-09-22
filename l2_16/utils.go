package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/url"
	"strings"
)

// forbiddenSymbols is a map with all forbidden filename symbols that are replaced with _
var forbiddenSymbols = map[rune]struct{}{
	'\\': {}, '/': {}, '>': {}, '<': {}, ':': {}, '"': {}, '|': {}, '?': {}, '*': {},
}

// GetFileFormat gets format from Content-Type
func GetFileFormat(contentType string, defaultFormat string) string {
	var fileFormat string

	// -1 -> len - 1, N -> N
	fileFormatEndsAt := strings.Index(contentType, ";")
	if fileFormatEndsAt == -1 {
		fileFormatEndsAt = len(contentType)
	}

	fileFormatStartsAt := strings.Index(contentType, "/")
	if fileFormatStartsAt == -1 || fileFormatStartsAt >= fileFormatEndsAt {
		fileFormatStartsAt = 0
	} else {
		fileFormatStartsAt++
	}

	fileFormat = contentType[fileFormatStartsAt:fileFormatEndsAt]

	if len(fileFormat) == 0 {
		return defaultFormat
	}

	return fileFormat
}

// GetLinks parses html from response and returns list of links that are fetched
//
// body is already fully-read response
func GetLinks(bodyReader io.ReadCloser) ([]string, error) {
	doc, err := html.Parse(bodyReader)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "href" || attr.Key == "src" {
					links = append(links, attr.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return links, nil
}

// IsSameDomain returns if someURL has the given host (url.Host) as domain
func IsSameDomain(host string, someURL string) (bool, error) {
	urlParsed, err := url.Parse(someURL)
	if err != nil {
		return false, fmt.Errorf("invalid URL: %v", err)
	}
	return urlParsed.Host == host, nil
}

// URLToFileName is used to convert URL into proper filename
func URLToFileName(someURL string) string {
	urlRunes := []rune(someURL)
	name := make([]rune, len(urlRunes))
	for i, c := range urlRunes {
		if _, ok := forbiddenSymbols[c]; ok {
			name[i] = '_'
		} else {
			name[i] = c
		}
	}

	fileName := string(name)

	indexDot := strings.LastIndex(someURL, ".")
	indexSlash := strings.LastIndex(someURL, "/")
	if indexDot == -1 || indexSlash > indexDot {
		fileName = fileName + ".html"
	}

	return fileName //fmt.Sprintf("%s.%s", string(name), fileFormat))
}
