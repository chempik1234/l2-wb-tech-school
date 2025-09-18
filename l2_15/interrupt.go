package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// createInterruptContext creates a context that tracks CTRL + C
//
// it launches a goroutine that waits for syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL, os.Interrupt
func createInterruptContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL)
		<-sigChan
		cancel()
	}()

	return ctx, cancel
}
