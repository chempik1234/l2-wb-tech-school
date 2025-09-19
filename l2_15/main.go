package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
)

type inputErr struct {
	input string
	err   error
}

func printFromChan(wg *sync.WaitGroup, input <-chan string) {
	defer wg.Done()
	for s := range input {
		fmt.Println(s)
	}
}

func executeCommand(ctx context.Context, pwd *string, input <-chan string, executeString string) <-chan string {
	resultChan := make(chan string)

	go func() {
		defer close(resultChan)

		fields := parseFields(executeString)

		if len(fields) == 0 {
			return
		}
		commandName := fields[0]
		args := fields[1:]

		for i, a := range args {
			if strings.HasPrefix(a, "$") {
				args[i] = os.Getenv(strings.TrimPrefix(a, "$"))
			}
		}

		if comFunc, ok := commands[commandName]; ok {
			comFunc(ctx, pwd, commandName, args, input, resultChan)
		} else {
			execCommand(ctx, pwd, commandName, args, input, resultChan)
		}
	}()
	return resultChan
}

func main() {
	wFlag := flag.Bool("w", false, "don't print pwd like '(GO CMD) C:\\Users\\User\\l2_15'. If testing, set TRUE!")
	flag.Parse()

	skipPrintPWD := *wFlag

	pwdStr, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Errorf("could not get pwd: %w", err))
	}

	pwd := &pwdStr

	// we'll use channel to be able to perform select{} with OS.KILL context vs input
	//region setup reader channel
	reader := bufio.NewReader(os.Stdin)
	readerChan := make(chan inputErr)

	go func(channel chan inputErr) {
		var input string
		var readErr error
		for {
			input, readErr = reader.ReadString('\n')
			channel <- inputErr{strings.TrimSuffix(input, "\n"), readErr}
		}
	}(readerChan)
	//

	var inputData inputErr
	var input string

	var tokens []string

	var toContinue bool

	var ok bool

main:
	for {
		if !skipPrintPWD {
			fmt.Print("(GO CMD) ", *pwd, ">")
		}

		toContinue = false

		ctx, stop := createInterruptContext()
		select {
		case <-ctx.Done():
			toContinue = true
		case inputData, ok = <-readerChan:
			if !ok {
				break main
			}
			break
		}

		err = inputData.err
		input = inputData.input

		if len(input) == 0 {
			toContinue = true
		}

		if isCtrlD(input) {
			err = io.EOF
		}

		if errors.Is(err, io.EOF) {
			stop()
			break
		} else if err != nil {
			log.Fatal(err)
		}

		if toContinue {
			fmt.Print("\n")
			continue
		}

		tokens = strings.Split(input, "|")

		if len(tokens) == 0 {
			continue
		}

		// clean all commands
		for i, v := range tokens {
			tokens[i] = cleanString(v)
		}

		if slices.Contains(tokens, "exit") {
			return
		}

		ctx, stop = signal.NotifyContext(context.Background(), os.Interrupt)

		onlyInputChan := make(chan string)
		go func(ctx context.Context, input chan inputErr, output chan<- string) {
			var value inputErr
		out2:
			for {
				select {
				case <-ctx.Done():
					break out2
				case value = <-input:
					if value.err != nil {
						break out2
					}
					if len(value.input) == 0 || isCtrlD(value.input) {
						break out2
					}
					select {
					case <-ctx.Done():
						input <- value
						break
					case onlyInputChan <- value.input:
						break
					}
				}
			}
			close(output)
		}(ctx, readerChan, onlyInputChan)

		var currentReadChan <-chan string
		currentReadChan = onlyInputChan
		for _, token := range tokens {
			currentReadChan = executeCommand(ctx, pwd, currentReadChan, token)
		}

		wg := &sync.WaitGroup{}
		wg.Add(1)

		go printFromChan(wg, currentReadChan)

		wg.Wait()

		stop()
	}
}
