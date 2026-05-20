package main

import (
	"net/http"
	"log"
	"os"
)

func main() {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./public_html"))
	mux.Handle("/", fs)

	server := http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	log.Println("Serving on http://localhost:8080")
	log.Fatal(server.ListenAndServe())

}