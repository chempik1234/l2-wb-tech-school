package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
)

func main() {
	// step 1. flags
	hFlag := flag.String("h", "", "host (default empty)")
	pFlag := flag.Int("p", 8080, "port (default 8080)")
	flag.Parse()

	host := *hFlag
	port := *pFlag

	address := fmt.Sprintf("%s:%d", host, port)

	// step 2. create base connection
	fmt.Println("connecting to", address)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(fmt.Errorf("error connecting to %s: %w", address, err))
	}
	defer conn.Close()

	// step 3. prepare root ctx
	ctx, cancelRoot := signal.NotifyContext(context.Background(), os.Interrupt)

	// step 4. start read/write
	var errorsChanWrite chan error
	ctx, _, errorsChanWrite = StartWritingConn(ctx, conn)

	_, _ = StartReadingConn(ctx, conn)

	// step 5. wait for completion
	select {
	case err = <-errorsChanWrite:
		break
	case <-ctx.Done():
		break
	}

	// step 6. shutdown
	if err != nil {
		log.Println(fmt.Errorf("error during writing: %w", err))
	}
	cancelRoot()
}
