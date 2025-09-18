package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("EXAMPLE")

	timeChan := time.After(time.Second * 10)
out:
	for {
		select {
		case <-timeChan:
			break out
		default:
			fmt.Println("some print")
			time.Sleep(time.Millisecond * 500)
		}
	}
	fmt.Println("EXAMPLE")
}
