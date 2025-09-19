package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Println("EXAMPLE")

	// better than scanln
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		fmt.Printf("you scanned: %s\n", scanner.Text())
	}
}
