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
	defer resp.Body.Close()
	w.WriteHeader(http.StatusOK)
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
	http.ListenAndServe(":8080", nil)
}
