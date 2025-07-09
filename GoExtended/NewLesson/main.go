package main

import (
	"context" // For graceful shutdown with timeout
	"fmt"
	"log" // For logging server messages
	"net/http"
	"os"        // To listen for OS signals
	"os/signal" // To handle OS signals
	"syscall"   // Specific OS signals like SIGINT, SIGTERM
	"time"      // For setting a shutdown timeout
)

func home_page(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "My first page")
}

func contacts_page(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Second Page")
}

func main() {
	// Create a new ServeMux for routing, rather than using the default global mux
	// This is generally good practice for more complex applications
	mux := http.NewServeMux()
	mux.HandleFunc("/", home_page)
	mux.HandleFunc("/contacts/", contacts_page)

	// Create an http.Server instance. This gives us more control than ListenAndServe
	server := &http.Server{
		Addr:    ":3030",
		Handler: mux, // Assign our custom mux
	}

	// Create a channel to listen for OS signals.
	// We're interested in SIGINT (Ctrl+C) and SIGTERM (graceful termination signal).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a separate goroutine.
	// This allows the main goroutine to continue and listen for signals.
	go func() {
		log.Printf("Server starting on %s", server.Addr)
		// ListenAndServe returns an error when it stops.
		// http.ErrServerClosed is expected during graceful shutdown.
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Block the main goroutine until a signal is received from the 'quit' channel.
	<-quit
	log.Println("Received shutdown signal. Shutting down server...")

	// Create a context with a timeout for graceful shutdown.
	// This ensures the server doesn't wait indefinitely for active connections to close.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Release resources associated with the context once shutdown is complete

	// Attempt a graceful shutdown. This stops new connections and waits for existing ones.
	if err := server.Shutdown(ctx); err != nil {
		// Log a fatal error if shutdown itself fails (e.g., due to timeout)
		log.Fatalf("Server graceful shutdown failed: %v", err)
	}

	log.Println("Server stopped successfully.")
}
