package main

import (
	"bufio"
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	fmt.Println("SASAAAL!")

	name := conn.RemoteAddr().String()
	fmt.Printf("%+v connected\n", name)
	conn.Write([]byte("Hello, " + name + "\n\r"))
	// conn impl. IOWriter, so we can write []byte

	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "Exit" {
			conn.Write([]byte("Bye\n\r"))
			fmt.Println(name, "disconnected")
			break
		} else if text != "" {
			fmt.Println(name, "enters", text)
			conn.Write([]byte("You enter " + text + "\n\r"))
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	var conn net.Conn
	for {
		conn, err = listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleConnection(conn)
	}

	fmt.Println("SASAAAL!")
}
