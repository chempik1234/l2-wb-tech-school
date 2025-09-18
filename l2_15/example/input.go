package main

import "fmt"

func main() {
	fmt.Println("EXAMPLE")

	var s string

	_, _ = fmt.Scanln(&s)

	fmt.Println("you scanned:", s)
}
