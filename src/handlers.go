package main


import(
	"net/http"
	"fmt"
	"encoding/json"
	"log"
	"github.com/chirpy/src/internal/database"
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
	cfg.fileServerHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	body := fmt.Sprintf("Hits: %d\n", cfg.fileServerHits.Load())
	w.Write([]byte(body))
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


func handlerCreateUser(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
    params := database.CreateUserParams{}
    err := decoder.Decode(&params)
    if err != nil {
		log.Printf("Error decoding parameters: %s\n", err)
		code := 500
		msg := fmt.Sprintf("Error decoding parameters: %s\n", err)
		respondWithError(w, code, msg) 
    }

	fmt.Printf("%s\n", params.Email)
	user, err := dbQueries.CreateUser(r.Context, params)
	if err != nil {
		fmt.Errorf("Failed to create user!!!")
	}
	fmt.Printf("%s\n", user.Email)
}


