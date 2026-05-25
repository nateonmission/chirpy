package main


import(
	"net/http"
	"fmt"
	"encoding/json"
	"log"
	"context"
	"github.com/chirpy/src/internal/database"
	"github.com/chirpy/src/internal/auth"
	"github.com/google/uuid"
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
		_ = cfg.dbQueries.DeleteAllChirps(ctx)
		cfg.fileServerHits.Store(0)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		body := fmt.Sprintf("Hits: %d\n", cfg.fileServerHits.Load())
		w.Write([]byte(body))
	}
}




func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
    userBase := CreateUserStruct{}
    err := decoder.Decode(&userBase)
    if err != nil {
		log.Printf("Error decoding parameters: %s\n", err)
		code := 500
		msg := fmt.Sprintf("Error decoding parameters: %s\n", err)
		respondWithError(w, code, msg) 
		return
    }

	hashedPassword, err := auth.HashPassword(userBase.Password);
	if err != nil {
		log.Printf("Error hashing password: %s\n", err)
		code := 500
		msg := fmt.Sprintf("Error hashing password: %s\n", err)
		respondWithError(w, code, msg)
		return
	}
	
	params := database.CreateUserParams{
		Email: userBase.Email,
		HashedPassword: hashedPassword,
	}

	// fmt.Printf("%s\n", params.Email)

	user, err := cfg.dbQueries.CreateUser(r.Context(), params)
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

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
    params := database.CreateChirpParams{}
    err := decoder.Decode(&params)
    if err != nil {
		log.Printf("Error decoding parameters: %s\n", err)
		code := 500
		msg := fmt.Sprintf("Error decoding parameters: %s\n", err)
		respondWithError(w, code, msg) 
		return
    }

	if len(params.Body) <= 140 {
		chirp, err := cfg.dbQueries.CreateChirp(r.Context(), params)
		if err != nil {
			log.Printf("Failed to create chirp: %s\n", err)
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create chirp.\n"))
			return
		}

		fmt.Printf("%s\n", chirp.ID)

		user, err := cfg.dbQueries.GetUserByID(r.Context(), chirp.UserID)
		if err != nil {
			log.Printf("Failed to create user: %s\n", err)
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to load user\n"))
			return
		}

		fmt.Printf("User, %s, chiped: '%s'\n", user.Email, chirp.Body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(chirp)
		return
	} else if len(params.Body) > 140 {
		code := 400
		msg := fmt.Sprintf("Chirp is too long")
		respondWithError(w, code, msg) 
		return
	} else {
		code := 500
		msg := fmt.Sprintf("Unknow Error")
		respondWithError(w, code , msg) 
		return
	}
}

func (cfg *apiConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.ListAllChirps(r.Context())
	if err != nil {
		log.Printf("Failed to get chirps: %s\n", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to load chirps\n"))
		return
	}

	// for chirp := range chirps {
	// 	user, err := cfg.dbQueries.GetUserByID(r.Context(), chirp.UserID)
	// 	if err != nil {
	// 		log.Printf("Failed to create user: %s\n", err)
	// 		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to load user\n"))
	// 		return
	// 	}

	// 	fmt.Printf()
	// }
	// fmt.Printf(chirps)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(chirps)
	return

}

func (cfg *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	chirpID_string := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpID_string)
	if err != nil {
		log.Fatalf("failed to parse UUID: %v", err)
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("Failed to get chirp: %s\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to load chirps\n"))
		return
	}

	w.Header().Set("Content-Type", "application/json");
	w.WriteHeader(http.StatusOK);
	json.NewEncoder(w).Encode(chirp);
	return;



}

func (cfg *apiConfig) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
    userBase := CreateUserStruct{}
    err := decoder.Decode(&userBase)
    if err != nil {
		log.Printf("Error decoding parameters: %s\n", err)
		code := 500
		msg := fmt.Sprintf("Error decoding parameters: %s\n", err)
		respondWithError(w, code, msg) 
		return
    }

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), userBase.Email)
	if err != nil {
		log.Printf("Failed to get user: %s\n", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to load user\n"))
		return
	}

	if validate, err := auth.CheckPasswordHash(userBase.Password, user.HashedPassword); validate && err == nil {
		cfg.loggedinUser = user.ID
		log.Printf("User %s logged in successfully\n", user.Email)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		respUser := User{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		}
		json.NewEncoder(w).Encode(respUser)
		return
	} else {
		log.Printf("User %s failed to log in\n", user.Email)
		code := 401
		msg := fmt.Sprintf("Invalid credentials")
		respondWithError(w, code, msg) 
		return
	}
	
}








