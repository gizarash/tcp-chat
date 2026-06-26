package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Server struct {
	clients map[net.Conn]struct{}
	mu sync.RWMutex
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("unable to start listener: %v", err)
	}
	defer listener.Close()
	fmt.Println("server is listening at port 8080...")
	server := &Server{
		clients: make(map[net.Conn]struct{}),
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("unable to accept connection: %v\n", err)
			continue
		}
		server.mu.Lock()
		server.clients[conn] = struct{}{}
		server.mu.Unlock()
		go handleConn(conn, server)
	}
}

func handleConn(conn net.Conn, s *Server) {
	defer conn.Close()
	log.Printf("[+] New connection: %s\n", conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.mu.Lock()
				delete(s.clients, conn)
				s.mu.Unlock()
				log.Printf("[-] Disconnected: %s", conn.RemoteAddr().String())
				break
			} else {
				log.Printf("unable to read client input: %v", err)
				break
			}
		}
		s.mu.RLock()
		for nextConn := range s.clients {
			if nextConn != conn {
				fmt.Fprintf(nextConn, "[%s]: %s", conn.RemoteAddr().String(), input)
			}
		}
		s.mu.RUnlock()
	}
}