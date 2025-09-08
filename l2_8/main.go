package main

import (
	"flag"
	"fmt"
	"github.com/beevik/ntp"
	"log"
)

func main() {
	addressFlag := flag.String("address", "time.google.com", "NTP server address")
	flag.Parse()

	address := *addressFlag
	fmt.Println("begin reading NTP:", address)
	result, err := ntp.Time(address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
