package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type APIHandler struct {
	apiUrl string
}

func (a *APIHandler) send(w http.ResponseWriter, r *http.Request) {
	log.Printf("received call from: %v\n", r.RemoteAddr)

	resp, err := http.Get(a.apiUrl)
	if err != nil {
		log.Printf("failed to call API: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Failed to fetch counter"))
		if err != nil {
			log.Printf("error writing response: %v", err)
		}
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API returned non-OK status: %d\n", resp.StatusCode)
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("API error"))
		if err != nil {
			log.Printf("error writing response: %v", err)
		}
		return
	}

	// Read the counter response from the API
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read response body: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Failed to read counter"))
		if err != nil {
			log.Printf("error writing response: %v", err)
		}
		return
	}

	// Return the counter to the client
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		log.Printf("error writing response: %v", err)
	}
}

var (
	ready bool = false
	mu    sync.Mutex
)

func main() {
	server := &http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/liveness", liveness)
	http.HandleFunc("/readiness", readiness)

	go func() {
		log.Println("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		log.Fatal("API_URL environment variable not set")
	}

	apiHandler := &APIHandler{apiUrl: apiURL}
	http.HandleFunc("/send", apiHandler.send)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	mu.Lock()
	ready = true
	mu.Unlock()

	log.Println("Application is READY")

	// Setup graceful shutdown
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint

	log.Println("Shutting down server...")

	mu.Lock()
	ready = false
	mu.Unlock()

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

func liveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintln(w, "OK"); err != nil {
		log.Printf("error writing liveness response: %v", err)
	}
}

func readiness(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if ready {
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintln(w, "OK"); err != nil {
			log.Fatal(err)
		}
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := fmt.Fprintln(w, "Service Unavailable"); err != nil {
			log.Fatal(err)
		}
	}
}
