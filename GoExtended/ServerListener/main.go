package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// Listen for incoming connections
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 8080")

	for {
		// Accept a new connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		// Handle the connection in a new goroutine
		go handleConnection(conn)
	}
}

// handleConnection reads a message and echoes it back
func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("New connection from:", conn.RemoteAddr())

	// Copy data from the connection's reader to its writer
	if _, err := io.Copy(conn, conn); err != nil {
		fmt.Println("Error handling connection:", err)
	}
	fmt.Println("Connection closed for:", conn.RemoteAddr())
}
