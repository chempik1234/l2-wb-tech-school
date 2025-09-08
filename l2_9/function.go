package l2_9

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalidUnpackString is an error that describes a case when string unpacking failure is the user's fault
var ErrInvalidUnpackString = errors.New("invalid unpack string")

// ErrInternalUnpackingError is an error that describes a case when encountered an unexpected error during unpacking
var ErrInternalUnpackingError = errors.New("internal unpacking error")

// Unpack converts string from format (.[0-9]+)+
//
// Example:
//
// "a4bc2d5e" => "aaaabccddddde"
//
// "45" => ErrInvalidUnpackString
//
// "" => ""
func Unpack(input string) (string, error) {
	/*
		The algorithm:
		1) iterate through every character
		2) if encountered a number then use it as repeat count
		3) finished the number or there was no number - print the previously read character
		4) after iterating print if any

		escaping:
		1) encountered \ - let the program know we're now "escaping"
		2) escaping - just set current char as the printed one and set escaping=false

		states:
		1) saw \, moving on
		2) saw something after \, set as current char
		3) see numbers, interpret as repeat counter digits
		4) see any else char, print current char * current counter times
	*/

	result := strings.Builder{}
	inputRunes := []rune(input)

	currentSymbol := '-'
	currentSymbolSet := false
	escaping := false

	var currentRepeatStringBuilder = strings.Builder{}

	for i := 0; i <= len(inputRunes); i++ {
		if i < len(inputRunes) {
			// abc\45 => current symbol is 5, no more actions: no repeat counting, no prev char print
			//     ^
			if escaping {
				currentSymbol = inputRunes[i]
				currentSymbolSet = true
				escaping = false
				continue
			}

			// begin escaping
			if inputRunes[i] == '\\' {
				escaping = true
			}

			// begin/continue writing repeat counter as string
			//
			// a123dv  repeat='1' + '2'='12'
			//   ^
			// a123dv  repeat='12' + '3' = '123'
			//    ^
			if inputRunes[i] >= '0' && inputRunes[i] <= '9' && !escaping {
				if !currentSymbolSet {
					return "", fmt.Errorf("%w: repeat number must follow the repeated rune, index %d",
						ErrInvalidUnpackString, i)
				}
				currentRepeatStringBuilder.WriteRune(inputRunes[i])
				continue
			}
		}

		// else we just write rune

		// abc123d4  write 'c' 123x
		//       ^
		if currentRepeatStringBuilder.Len() > 0 {
			if currentSymbolSet {
				currentRepeatStringValue := currentRepeatStringBuilder.String()

				repeatNumber, err := strconv.Atoi(currentRepeatStringValue)
				if err != nil {
					return "", fmt.Errorf("%w: repeat number '%s' couldn't be converted, index %d",
						ErrInternalUnpackingError, currentRepeatStringValue, i)
				}
				for j := 0; j < repeatNumber; j++ {
					result.WriteRune(currentSymbol)
				}
				currentSymbolSet = false

				currentRepeatStringBuilder.Reset()
			} else {
				return "", fmt.Errorf("%w: no symbol set to repeat, index %d",
					ErrInvalidUnpackString, i)
			}
		}

		if currentSymbolSet {
			result.WriteRune(currentSymbol)
		}

		if i < len(inputRunes) && !escaping {
			currentSymbol = inputRunes[i]
			currentSymbolSet = true
		}
	}

	if escaping {
		return "", fmt.Errorf("%w: mustn't end with escaping", ErrInvalidUnpackString)
	}

	return result.String(), nil
}
