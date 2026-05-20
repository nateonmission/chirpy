package main

import (
	"net/http"
	"log"
	"os"
)

func main() {
	mux := http.NewServeMux()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Working directory:", wd)

	_, err = os.Stat("./public_html/index.html")
	if err != nil {
		log.Fatal("Cannot find ./public_html/index.html: ", err)
	}

	fs := http.FileServer(http.Dir("./public_html"))
	mux.Handle("/", fs)

	server := http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	log.Println("Serving on http://localhost:8080")
	log.Fatal(server.ListenAndServe())

}