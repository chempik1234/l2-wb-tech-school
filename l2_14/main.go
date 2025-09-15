package main

import (
	"fmt"
	"time"
)

// or function returns a channels that's closed as soon as any of given channel sends
//
// for loop through channels list until one of them is ready
func or(channels ...<-chan any) <-chan any {
	c := make(chan any)
	// launch checker in background
	go func() {
		for {
			for _, channel := range channels {
				select {
				case <-channel:
					close(c)
					return
				default:
					break
				}
			}
		}
	}()
	return c
}

// main is taken directly from task
func main() {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v", time.Since(start))

	/*
		done after 1.0004893s
	*/
}
