package main

import (
	"net/http"
	"log"
)


func main() {
	cfg := &apiConfig{}
	cfg.fileServerHits.Store(0)
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./public_html"))
	appHandler := http.StripPrefix("/app/", fs)

	mux.Handle("/app/", cfg.middlewareMetricsInc(appHandler))

	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)

	server := http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	log.Println("Serving on http://localhost:8080")
	log.Fatal(server.ListenAndServe())

}