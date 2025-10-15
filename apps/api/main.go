package main

import (
	"log"
	"net/http"
)

func root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	log.Printf("received call from: %v\n", r.RemoteAddr)
}

func main() {
	http.HandleFunc("/", root)
	log.Println("Starting server on :3030")
	http.ListenAndServe(":3030", nil)
}
