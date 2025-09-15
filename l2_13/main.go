package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func columnNeeded(cols map[uint8]struct{}, col uint8) bool {
	_, ok := cols[col]
	return ok
}

// parseFieldIndices parses string like A-B,N,M-P ...
func parseFieldIndices(s string) (map[uint8]struct{}, error) {
	// parse fieldIndices
	// 132,567-134
	//  ^       ^
	//  currentNumFrom = 13
	//          ^
	//          ^
	//        currentNumTo = 13
	fieldIndices := make(map[uint8]struct{})
	var currentNumFrom uint8 = 0
	var currentNumTo uint8 = 0

	isNumFromFinished := false // if true then append digits to currentNumTo, else currentNumFrom

	/*
		isNumFromFinished && isNumToFinished	= 1-3
		isNumFromFinished && !isNumToFinished	= entering currentNumTo
		!isNumFromFinished 						= entering currentNumFrom
	*/

	// index for error traceback
	i := 0
	for _, r := range s + "," {
		if '0' <= r && r <= '9' {
			if isNumFromFinished {
				currentNumTo = currentNumTo*10 + uint8(r-'0') // + 0-9
			} else {
				currentNumFrom = currentNumFrom*10 + uint8(r-'0') // + 0-9
			}
		} else if r == '-' {
			if currentNumFrom <= 0 {
				return nil, fmt.Errorf("invalid syntax at position %d", i)
			}
			currentNumTo = 0
			isNumFromFinished = true
			// isNumToFinished = false
		} else if r == ',' {
			if isNumFromFinished {
				if currentNumTo <= 0 {
					return nil, fmt.Errorf("invalid syntax at position %d", i)
				}
				for f := currentNumFrom; f <= currentNumTo; f++ {
					fieldIndices[f] = struct{}{}
				}
			} else {
				if currentNumFrom <= 0 {
					return nil, fmt.Errorf("invalid syntax at position %d", i)
				}
				fieldIndices[currentNumFrom] = struct{}{}
			}
			isNumFromFinished = false
			// isNumToFinished = false
			currentNumTo = 0
			currentNumFrom = 0
		} else {
			return nil, errors.New("only digits, '-' and ',' are allowed")
		}

		i++
	}
	return fieldIndices, nil
}

func main() {
	var err error

	fFlag := flag.String("f", "", "fields to print, count from 1 example: 1,3-5")
	dFlag := flag.String("d", "\t", "custom delimiter, only 1 symbol")
	sFlag := flag.Bool("s", false, "print only strings that contain delimiter")

	flag.Parse()

	fieldsFlagString := *fFlag       // get fields string, then parse to map like 1,3-5,1,3 -> 1: struct{}, 3: struct{}, 4: struct{}, 5: struct{}
	delimiterRunes := []rune(*dFlag) // get the "delimiting rune" from that, example: Ю -> [Ю] -> 'Ю'
	printOnlySeparated := *sFlag     // if -s then we cache the first column and print it when encounter 1st delimiter

	if len(delimiterRunes) != 1 {
		log.Fatal("you must specify 1 char for delimiter, example: \"\\t\"")
	}

	delimiter := delimiterRunes[0]

	if len(fieldsFlagString) == 0 {
		log.Fatal("-f flag is required")
	}

	var fieldIndices map[uint8]struct{}
	fieldIndices, err = parseFieldIndices(fieldsFlagString)
	if err != nil {
		log.Fatal(fmt.Errorf("invalid fields interval: %w", err))
	}

	reader := bufio.NewReader(os.Stdin)

	var input string
	var currentCol uint8

	// if -s then we cache the first column and print it when encounter 1st delimiter
	var firstColBuffer strings.Builder
	if printOnlySeparated {
		firstColBuffer = strings.Builder{}
	}

	// defined is we should print \n at the end
	var printedAnything bool

	// defines if something was printed in the current col
	//      AAAAAAA    BBBB
	//      ^    ^     ^
	//    false  |   false again
	//         true
	var printedAnythingInColumn bool

	for {
		currentCol = 1
		printedAnything = false
		printedAnythingInColumn = false

		input, err = reader.ReadString('\n')

		// При конце файла выходим
		errors.Is(err, nil)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		input = strings.TrimSuffix(input, "\n")

		// let's say string has MMM len, has MM columns and there are M fields to print, MMM>MM>M
		//
		// these are 2 options:
		// 1. split string and output only needed columns
		//    loop through string (MMM), then print M columns
		// 2. loop through string and output columns mid-loop
		//    loop through string MMM and print (print mid-air)
		//
		// we choose option 2
		firstColBuffer.Reset()
		for _, r := range input {
			// we either encounter a delimiter or not.
			// if delimiter then:
			//  1. col++
			//  2. if we have cached 1st column then also print it
			//
			// else: print if needed (if we know there are more columns or if it is one of those 'more' column)
			//       also print delimiter before a new col if it isn't 1st to be printed (use printedAnything)
			//
			//  AAAAA=AAA
			//       ^
			//     col=2
			//   but we don't      because that's a
			//   print it yet      delimiter symbol

			// step 1. if delimiter then...
			if r == delimiter {
				if !printedAnything && columnNeeded(fieldIndices, currentCol) && printOnlySeparated && firstColBuffer.Len() > 0 {
					fmt.Print(firstColBuffer.String())
					printedAnything = true
				}
				printedAnythingInColumn = false
				currentCol++
			}

			// step 2 if it's the first col to be printed in this line, save whatever new char into firstCol
			if !printedAnything && r != delimiter {
				firstColBuffer.WriteRune(r)
			}

			// if current col is to be printed, we print it.
			// BUT:
			//  1. not if we try to ignore no-delimiters rows (that's what if delimiter == true condition does)
			//  2. we print delimiter if it's not first column to be actually printed
			if columnNeeded(fieldIndices, currentCol) && r != delimiter &&
				(printOnlySeparated && currentCol > 1 || !printOnlySeparated) {
				if !printedAnythingInColumn && printedAnything {
					fmt.Print(string(delimiter))
				}

				fmt.Print(string(r))
				printedAnything = true
				printedAnythingInColumn = true
			}
		}

		if printedAnything || !printOnlySeparated {
			fmt.Print("\n")
		}
	}
}
