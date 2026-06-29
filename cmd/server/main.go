package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type Server struct {
	clients map[net.Conn]string
	mu      sync.RWMutex
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("unable to start listener: %v", err)
	}
	defer listener.Close()
	fmt.Println("server is listening at port 8080...")
	s := &Server{
		clients: make(map[net.Conn]string),
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("unable to accept connection: %v\n", err)
			continue
		}
		s.mu.Lock()
		s.clients[conn] = ""
		s.mu.Unlock()
		go handleConn(conn, s)
	}
}

func handleConn(conn net.Conn, s *Server) {
	defer conn.Close()
	log.Printf("[+] New connection: %s\n", conn.RemoteAddr().String())
	fmt.Fprint(conn, "Enter your name: ")
	reader := bufio.NewReader(conn)
	userName := ""
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("unable to read client input: %v", err)
			}
			disconnect(conn, s)
			break
		}
		if userName == "" {
			if validateName(input) {
				s.mu.Lock()
				userName = strings.TrimSpace(input)
				s.clients[conn] = userName
				for nextConn := range s.clients {
					if nextConn != conn && s.clients[nextConn] != "" {
						fmt.Fprintf(nextConn, "* %s joined the chat\n", s.clients[conn])
					}
				}
				s.mu.Unlock()
			} else {
				fmt.Fprintln(conn, "Name cannot be empty, longer than 20 characters, or contain any space characters.")
				fmt.Fprint(conn, "Enter your name: ")
				continue
			}
		} else {
			if strings.HasPrefix(input, "/") {
				trimmedInput := strings.TrimSpace(input)
				if trimmedInput == "/list" {
					s.mu.RLock()
					for nextConn := range s.clients {
						if nextConn != conn && s.clients[nextConn] != "" {
							fmt.Fprintln(conn, s.clients[nextConn])
						}
					}
					s.mu.RUnlock()
				} else if trimmedInput == "/quit" {
					disconnect(conn, s)
					break
				} else if strings.HasPrefix(trimmedInput, "/msg ") {
					cmdSlice := strings.SplitN(trimmedInput, " ", 3)
					if len(cmdSlice) >= 3 {
						reciever := cmdSlice[1]
						message := cmdSlice[2]
						if strings.TrimSpace(message) == "" {
							fmt.Fprintln(conn, "message cannot be empty")
						} else {
							s.mu.RLock()
							found := false
							for nextConn := range s.clients {
								if nextConn != conn && s.clients[nextConn] == reciever {
									fmt.Fprintf(nextConn, "[private] %s: %s\n", s.clients[conn], message)
									fmt.Fprintf(conn, "[private -> %s]: %s\n", s.clients[nextConn], message)
									found = true
									break
								}
							}
							if !found {
								fmt.Fprintf(conn, "user \"%s\" not found\n", reciever)
							}
							s.mu.RUnlock()
						}
					} else {
						fmt.Fprintln(conn, "correct msg command is \"/msg username message\"")
					}
				} else {
					fmt.Fprintf(conn, "unknown command: %s\n", trimmedInput)
				}
			} else {
				s.mu.RLock()
				for nextConn := range s.clients {
					if nextConn != conn && s.clients[nextConn] != "" && strings.TrimSpace(input) != "" {
						fmt.Fprintf(nextConn, "[%s]: %s", s.clients[conn], input)
					}
				}
				s.mu.RUnlock()
			}
		}
	}
}

func validateName(name string) bool {
	trimmedName := strings.TrimSpace(name)
	return trimmedName != "" && len(trimmedName) <= 20 && !strings.Contains(trimmedName, " ")
}

func disconnect(conn net.Conn, s *Server) {
	s.mu.Lock()
	name := s.clients[conn]
	delete(s.clients, conn)
	if name != "" {
		for nextConn := range s.clients {
			if s.clients[nextConn] != "" {
				fmt.Fprintf(nextConn, "* %s left the chat\n", name)
			}
		}
	}
	s.mu.Unlock()
}
