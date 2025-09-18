package main

import (
	"bufio"
	"context"
	"fmt"
	gops "github.com/mitchellh/go-ps"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

func transferOutput(reader io.Reader, prefix string, resultChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		resultChan <- line
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("\nError reading %s: %v\n", prefix, err)
	}
}

type commandFunc func(ctx context.Context, pwd *string, commandName string, args []string, inputChan <-chan string, resultChan chan<- string)

var echoCommand commandFunc = func(ctx context.Context, pwd *string, commandName string, args []string, inputChan <-chan string, resultChan chan<- string) {
	resultChan <- strings.Join(args, " ")
}

var pwdCommand commandFunc = func(ctx context.Context, pwd *string, commandName string, args []string, inputChan <-chan string, resultChan chan<- string) {
	resultChan <- *pwd
}

var cdCommand commandFunc = func(ctx context.Context, pwd *string, commandName string, args []string, inputChan <-chan string, resultChan chan<- string) {
	if err := os.Chdir(args[0]); err != nil {
		resultChan <- fmt.Sprintf("directory %s doesn't exist", args[0])
	} else {
		*pwd, err = os.Getwd()
		if err != nil {
			resultChan <- fmt.Sprintf("couldn't get wd: %v", err)
		}
	}
}

var killCommand commandFunc = func(ctx context.Context, pwd *string, commandName string, args []string, inputChan <-chan string, resultChan chan<- string) {
	pid, err := strconv.Atoi(args[0])
	if err != nil {
		resultChan <- fmt.Sprintf("couldn't convert pid %v to int", args[0])
	}

	var process *os.Process
	process, err = os.FindProcess(pid)
	if err != nil {
		resultChan <- fmt.Sprintf("couldn't find process %d", pid)
	}
	err = process.Kill()
	if err != nil {
		resultChan <- fmt.Sprintf("couldn't kill pid %v: %v", pid, err)
	}
}

var psCommand commandFunc = func(ctx context.Context, pwd *string, commandName string, args []string, inputChan <-chan string, resultChan chan<- string) {
	psList, err := gops.Processes()
	if err != nil {
		resultChan <- fmt.Sprintf("couldn't get process list: %v", err)
		return
	}
	resultChan <- fmt.Sprintf("      PID    PPID\tCMD")
	for _, proc := range psList {
		resultChan <- fmt.Sprintf("%9.0d\t%5.0d\t%s", proc.Pid(), proc.PPid(), proc.Executable())
	}
}

func execCommand(ctx context.Context, pwd *string, commandName string, commandArgs []string, inputChan <-chan string, resultChan chan<- string) {
	// firstly remove <, > from args

	//region redirects
	args, inputRedirect, outputRedirect, err := parseArgs(commandArgs)
	if err != nil {
		resultChan <- fmt.Errorf("ERROR: %w", err).Error()
	}
	//endregion

	//region init cmd
	cmd := exec.CommandContext(ctx, commandName, args...)
	if pwd != nil {
		cmd.Dir = *pwd
	}
	//endregion

	//region setup input
	if len(inputRedirect) > 0 {
		var file *os.File
		file, cmd.Stdin, err = bufferFromFile(inputRedirect)
		if err != nil {
			resultChan <- fmt.Sprintf("couldn't open input file: %v", err)
		}

		// close file to read
		defer closeFile(file)
	} else {
		cmd.Stdin = NewChannelReader(inputChan)
	}
	//endregion

	// Get pipes to output them into resultChannel

	//region setup output pipe
	var stdoutPipe, stderrPipe io.ReadCloser

	stdoutPipe, err = cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("\nstdout pipe error: %v\n", err)
		return
	}

	stderrPipe, err = cmd.StderrPipe()
	if err != nil {
		fmt.Printf("\nstderr pipe error: %v\n", err)
		return
	}
	//endregion

	err = cmd.Start()
	if err != nil {
		resultChan <- fmt.Sprintf("cmd start error: %v", err)
		return
	}

	// "Process started with PID: %d", cmd.Process.Pid

	//region direct pipes to output
	wg := &sync.WaitGroup{}
	wg.Add(2)

	var outputChannel chan<- string
	if len(outputRedirect) > 0 {
		var outputFile *os.File
		outputFile, err = createFile(outputRedirect)
		if err != nil {
			fmt.Printf("\noutput file error: %v\n", err)
			return
		}
		defer closeFile(outputFile)

		outputChannel = chanToFile(outputFile)
	} else {
		outputChannel = resultChan
	}

	go transferOutput(stdoutPipe, "stdout", outputChannel, wg)
	go transferOutput(stderrPipe, "stderr", outputChannel, wg)
	//endregion

	// now just wait
	err = cmd.Wait()
	wg.Wait()

	if err != nil {
		fmt.Printf("\nProcess %d exited with error: %v\n", cmd.Process.Pid, err)
	}

	// empty the rest values in input channel
	emptyChannel(inputChan)
}

var commands = map[string]commandFunc{
	"kill": killCommand,
	"pwd":  pwdCommand,
	"cd":   cdCommand,
	"echo": echoCommand,
	"ps":   psCommand,
}
