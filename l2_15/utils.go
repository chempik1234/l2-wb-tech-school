package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

func cleanString(s string) string {
	return strings.TrimSpace(strings.TrimSuffix(s, "\n"))
}

func isCtrlD(s string) bool {
	sRune, _ := utf8.DecodeRuneInString(s)
	if sRune == 4 {
		return true
	}
	return s == string([]rune{4, 13}) || s == "^D"
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {

	}
}

func bufferFromFile(filename string) (*os.File, io.Reader, error) {
	file, err := os.Open(cleanString(filename))
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't open input file: %w", err)
	}
	return file, bufio.NewReader(file), nil
}

func chanToFile(file *os.File) chan<- string {
	resultChan := make(chan string)

	go func() {
		var err error
		for v := range resultChan {
			_, err = file.WriteString(v + "\n")
			if err != nil {
				fmt.Println("error writing to file:", file.Name(), err)
			}
		}
	}()

	return resultChan
}

func createFile(filepath string) (*os.File, error) {
	// try to open, maybe it exists
	file, err := os.OpenFile(filepath, os.O_WRONLY, os.ModePerm)
	if err != nil {
		// step 1. create directory
		var directory string
		if strings.Contains(filepath, "/") || strings.Contains(filepath, "\\") {
			filepath = strings.Replace(filepath, "/", "\\", -1)
			directory = filepath[:strings.LastIndex(filepath, "\\")]
			err = os.MkdirAll(directory, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("couldn't create directory %s: %v", filepath, err)
			}
		}

		// step 2. create file
		_, err = os.Create(filepath)
		if err != nil {
			return nil, fmt.Errorf("couldn't create file after creating directory %s: %v", filepath, err)
		}

		// step 3. open again
		file, err = os.OpenFile(filepath, os.O_WRONLY, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("couldn't open file after creating path & file %s: %v", filepath, err)
		}
	}
	return file, nil
}

// parseArgs parses args and redirects from raw args list
//
// returns: args, input, output, error
func parseArgs(args []string) ([]string, string, string, error) {
	result := make([]string, 0, len(args))

	inputRedirect, outputRedirect := "", ""
	settingInputRedirect, settingOutputRedirect := false, false
	for _, arg := range args {

		if settingInputRedirect { //  < "sora.txt"
			inputRedirect = cleanString(arg)
			settingInputRedirect = false

		} else if settingOutputRedirect { //  > "sora.txt"
			outputRedirect = cleanString(arg)
			settingOutputRedirect = false

		} else if arg == "<" { //  "<" sora.txt
			if len(inputRedirect) > 0 {
				return nil, "", "", errors.New("can't set multiple input redirects")
			}

			settingInputRedirect = true

		} else if arg == ">" { //  ">" sora.txt
			if len(outputRedirect) > 0 {
				return nil, "", "", errors.New("can't set multiple output redirects")
			}

			settingOutputRedirect = true

		} else { //  random_arg
			settingOutputRedirect = false
			settingInputRedirect = false
			result = append(result, arg)
		}
	}
	return result, inputRedirect, outputRedirect, nil
}

// parseFields converts a string
//
// 'echo sss"home1 ohome2" "home3\n"'
//
// into
//
// [`echo`, `ssshome1 ohome2` `home3\n`]
func parseFields(executeString string) []string {
	// read fields, but not with strings.Fields because of quotes!
	//
	// no escaping with \
	//
	// "aaaaaaaaaaaaaaaaaa         aaaaaaaa"
	// true true true true true true true  false
	//
	readingQuote := false

	// if we found any reason to finish word, we append it whole into slice
	foundSeparator := false

	// currently building field
	currentFieldBuffer := strings.Builder{}

	fields := make([]string, 0)
	for _, char := range executeString + " " {
		foundSeparator = false
		if char == '"' {
			if readingQuote {
				foundSeparator = true
			}
			readingQuote = !readingQuote
		} else if unicode.IsSpace(char) {
			if !readingQuote {
				foundSeparator = true
			} else {
				currentFieldBuffer.WriteRune(char)
			}
		} else {
			currentFieldBuffer.WriteRune(char)
		}

		if foundSeparator {
			if currentFieldBuffer.Len() == 0 {
				continue
			}
			fields = append(fields, currentFieldBuffer.String())
			currentFieldBuffer.Reset()
		}
	}

	return fields
}

func emptyChannel[T any](ch <-chan T) {
out3:
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				break out3
			}
			continue
		default:
			break out3
		}
	}
}
