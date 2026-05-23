package main


import(
	"net/http"
	"fmt"
	"encoding/json"
	"log"
	"context"
)

func healthHandler(w http.ResponseWriter, r *http.Request){

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))

}


func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	body := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileServerHits.Load())
	w.Write([]byte(body))

}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request){
	if cfg.platform != "dev" {
		code := 403
		msg := fmt.Sprintf("403 Unauthorized")
		respondWithError(w, code, msg) 
	} else {
		ctx := context.Background()
		_ = cfg.dbQueries.DeleteAllUsers(ctx)
		cfg.fileServerHits.Store(0)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		body := fmt.Sprintf("Hits: %d\n", cfg.fileServerHits.Load())
		w.Write([]byte(body))
	}
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
    params := chirpToValidate{}
    err := decoder.Decode(&params)
    if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		code := 500
		msg := fmt.Sprintf("Error decoding parameters: %s", err)
		respondWithError(w, code, msg) 
    }

	if len(params.Body) <= 140 {
		code := 200
		respondWithJSON(w, code, params)
	} else if len(params.Body) > 140 {
		code := 400
		msg := fmt.Sprintf("Chirp is too long")
		respondWithError(w, code, msg) 
	} else {
		code := 500
		msg := fmt.Sprintf("Unknow Error")
		respondWithError(w, code , msg) 
	}
}


func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
    params := createUserParam{}
    err := decoder.Decode(&params)
    if err != nil {
		log.Printf("Error decoding parameters: %s\n", err)
		code := 500
		msg := fmt.Sprintf("Error decoding parameters: %s\n", err)
		respondWithError(w, code, msg) 
		return
    }

	// fmt.Printf("%s\n", params.Email)

	user, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("Failed to create user: %s", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create user. user: %s already exists\n", params.Email))
		return
	}

	fmt.Printf("Created user: %s\n", user.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}


