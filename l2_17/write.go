package main

import (
	"context"
	"fmt"
	"net"
	"os"
)

// StartWritingConn starts a goroutine that writes text through TCP from stdin
//
// returns a context with cancel on EOF and errors chan
func StartWritingConn(ctx context.Context, conn net.Conn) (context.Context, context.CancelFunc, chan error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)

	errChan := make(chan error)

	go runWriting(ctx, cancel, conn, errChan)

	return ctx, cancel, errChan
}

// runWriting is a blocking function that runs wiring session from stdin to conn
//
// meant to be launched in background
func runWriting(ctx context.Context, cancel context.CancelFunc, conn net.Conn, errChan chan error) {
	// step 1. setup input channel
	// step 2. input chan VS ctx.Done(), write to conn
	// step 3. cancel()

	// step 1.
	inputChan := make(chan string)
	go ScannerToChan(os.Stdin, inputChan)
	fmt.Println("-- ready to write")

	// step 2.
	var line string
	var ok bool
out:
	for {
		// check ctx
		select {
		case <-ctx.Done():
			break out
		case line, ok = <-inputChan:
			if !ok {
				break out
			}
			break
		}

		_, err := conn.Write([]byte(line + "\n"))
		if err != nil {
			errChan <- fmt.Errorf("error writing to %s: %w", conn.RemoteAddr().String(), err)
			break
		}
	}

	// step 3.
	fmt.Println("-- end writing")
	cancel()
}
