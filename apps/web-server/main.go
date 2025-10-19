package main

import (
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
		return
	}
	w.WriteHeader(http.StatusOK)
	err = resp.Body.Close()
	if err != nil {
		log.Fatalf("error occurred when closing body: %v", err)
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
		log.Fatalf("error when listen and server: %v", err)
	}
}
