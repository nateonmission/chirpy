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
		polkaKey: os.Getenv("POLKA_KEY"),

	}
	cfg.fileServerHits.Store(0)
	cfg.tokenSecret = os.Getenv("JWT_SECRET")
	

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./public_html"))
	appHandler := http.StripPrefix("/app/", fs)

	mux.Handle("/app/", cfg.middlewareMetricsInc(appHandler))

	// mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)
	mux.HandleFunc("GET /api/healthz", healthHandler)

	mux.HandleFunc("POST /api/users", cfg.createUserHandler)
	mux.HandleFunc("PUT /api/users", cfg.updateUserHandler)

	mux.HandleFunc("POST /api/login", cfg.loginUserHandler)
	mux.HandleFunc("POST /api/refresh", cfg.refreshTokenHandler)
	mux.HandleFunc("POST /api/revoke", cfg.revokeTokenHandler)

	mux.HandleFunc("POST /api/chirps", cfg.createChirpHandler)
	mux.HandleFunc("GET /api/chirps", cfg.getAllChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirpByIDHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.deleteChirpHandler)

	mux.HandleFunc("POST /api/polka/webhooks", cfg.polkaWebhookHandler)

	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)


	server := http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	log.Println("Serving on http://localhost:8080")
	log.Fatal(server.ListenAndServe())

}