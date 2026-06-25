package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("unable to start listener: %v\n", err)
		return
	}
	defer listener.Close()
	fmt.Println("server is listening at port 8080...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("unable to accept connection: %v\n", err)
			return
		}
		handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("[+] New connection: %s\n", conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("[-] Disconnected: %s\n", conn.RemoteAddr().String())
				break
			} else {
				fmt.Printf("unable to read client input: %v\n", err)
				break
			}
		}
		fmt.Fprintf(conn, "echo: %s", input)
	}
}