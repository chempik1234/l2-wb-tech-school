package main

import (
	"bufio"
	"io"
	"unicode/utf8"
)

// ScannerToChan transfers lines from reader to chan
func ScannerToChan(reader io.Reader, output chan<- string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		if isCtrlD(text) {
			break
		}
		output <- scanner.Text()
	}
	close(output)
}

func isCtrlD(s string) bool {
	sRune, _ := utf8.DecodeRuneInString(s)
	if sRune == 4 {
		return true
	}
	return s == string([]rune{4, 13}) || s == "^D"
}
