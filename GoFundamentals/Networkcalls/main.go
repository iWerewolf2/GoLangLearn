package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	// Connect to a server on the specified port
	message := "Hello from the client!"

	// Create a TCP connection
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1) // Exit with an error code
	} else {
		// Send the message
		fmt.Println("Sending:", message)
		err = net.Send(conn, []byte(message))

		// Check for errors
		if err != nil {
			fmt.Println("Error sending message:", err)
			return // Return from the function if an error occurs
		}
		// Receive the response (example assumes a simple "echo" server)
		buffer := make([]byte, 1024)
		message, err = net.Read(buffer)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return // Exit from the function if an error occurs

		} else {
			fmt.Println("Received:", message)
		}
	}
	// Close the connection when finished
	time.Sleep(time.Second * 5)
	err = net.Close()
	if err != nil {
		fmt.Println("Error closing server:", err)
	}

	// ... (rest of the code to handle the network connection)
}
