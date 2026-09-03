package main

import (
	"fmt"
	"log"
	"net/http"

	"go-udp-server/api"
	"go-udp-server/api/handlers"
)

// enableCORS middleware allows the Web UI (port 3000) to securely talk to the API (port 8000)
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

func main() {
	fmt.Println("Starting Auth API Server...")

	// Initialize the PostgreSQL Database connection
	api.InitDB()

	// Register the endpoints with CORS enabled
	http.HandleFunc("/register", enableCORS(handlers.Register))
	http.HandleFunc("/login", enableCORS(handlers.Login))

	// Start the HTTP server on port 8000
	port := ":8000"
	fmt.Printf("API Server is running and listening on http://localhost%s\n", port)
	
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Failed to start API server:", err)
	}
}
