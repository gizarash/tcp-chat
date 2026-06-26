package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("unable to start listener: %v", err)
	}
	defer listener.Close()
	fmt.Println("server is listening at port 8080...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("unable to accept connection: %v\n", err)
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	log.Printf("[+] New connection: %s\n", conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("[-] Disconnected: %s", conn.RemoteAddr().String())
				break
			} else {
				log.Printf("unable to read client input: %v", err)
				break
			}
		}
		fmt.Fprintf(conn, "echo: %s", input)
	}
}