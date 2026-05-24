package main

import (
	"net/http"
	"log"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"os"
	"database/sql"
	"fmt"
	"github.com/chirpy/src/internal/database"
	// "context"
)


func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err!= nil {
		fmt.Errorf("Error: %w", err)
	}
	dbQueries := database.New(db)
	if dbQueries != nil {
		fmt.Printf("Database Loaded!\n")
	}
	// ctx := context.Background()
	// _ = dbQueries.DeleteAllUsers(ctx)


	cfg := &apiConfig{
		platform: os.Getenv("PLATFORM"),
		dbQueries: dbQueries,

	}
	cfg.fileServerHits.Store(0)
	

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./public_html"))
	appHandler := http.StripPrefix("/app/", fs)

	mux.Handle("/app/", cfg.middlewareMetricsInc(appHandler))

	// mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("POST /api/users", cfg.createUserHandler)
	mux.HandleFunc("POST /api/chirps", cfg.createChirpHandler)
	mux.HandleFunc("GET /api/chirps", cfg.getAllChirpsHandler)

	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)

	server := http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	log.Println("Serving on http://localhost:8080")
	log.Fatal(server.ListenAndServe())

}