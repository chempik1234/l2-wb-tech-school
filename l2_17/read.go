package main

import (
	"context"
	"fmt"
	"net"
)

// StartReadingConn starts a goroutine that reads from TCP and prints in fmt.Println
//
// returns a context with cancel on EOF
func StartReadingConn(ctx context.Context, conn net.Conn) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)

	go runReading(ctx, cancel, conn)

	return ctx, cancel
}

// runReading is a blocking function that runs reading session from conn to fmt.Println
//
// meant to be launched in background
func runReading(ctx context.Context, cancel context.CancelFunc, conn net.Conn) {
	// step 1. create input channel
	// step 2. read bufio chan VS ctx.Done()
	// step 3. cancel()

	// step 1.
	scannerLines := make(chan string)
	go ScannerToChan(conn, scannerLines)

	fmt.Println("-- ready to read")

	// step 2.
	var text string
	var ok bool
out:
	for {
		select {
		case <-ctx.Done():
			break out
		case text, ok = <-scannerLines:
			if !ok {
				break out
			}
			break
		}
		fmt.Println("received:", text)
	}

	// step 3.
	fmt.Println("-- end reading")
	cancel()
}
