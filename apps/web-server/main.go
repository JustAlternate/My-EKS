package main

import (
	"io"
	"log"
	"net/http"
	"os"
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
		_, err = w.Write([]byte("Failed to fetch counter"))
		if err != nil {
			log.Printf("error when writing to the client: %v", err)
		}
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("API returned non-OK status: %d\n", resp.StatusCode)
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("API error"))
		if err != nil {
			log.Printf("error when writing to the client: %v", err)
		}
		return
	}

	// Read the counter response from the API
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read response body: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("Failed to read counter"))
		if err != nil {
			log.Printf("error when writing to the client: %v", err)
		}
		return
	}

	// Return the counter to the client
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write(body)
	if err != nil {
		log.Printf("error when writing to the client: %v", err)
	}
}

func main() {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		log.Fatal("API_URL environment variable not set")
	}

	apiHandler := &APIHandler{apiUrl: apiURL}

	http.HandleFunc("/send", apiHandler.send)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Starting server on :8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("error when listen and serve: %v", err)
	}
}
