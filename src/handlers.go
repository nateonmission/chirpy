package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chirpy/src/internal/auth"
	"github.com/chirpy/src/internal/database"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK"))

}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	body := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileServerHits.Load())
	w.Write([]byte(body))

}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
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

	hashedPassword, err := auth.HashPassword(userBase.Password)
	if err != nil {
		log.Printf("Error hashing password: %s\n", err)
		code := 500
		msg := fmt.Sprintf("Error hashing password: %s\n", err)
		respondWithError(w, code, msg)
		return
	}

	params := database.CreateUserParams{
		Email:          userBase.Email,
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

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer token: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error getting bearer token: %s\n", err))
		return
	}

	claims, err := auth.ValidateJWT(tokenString, cfg.tokenSecret)
	if err != nil {
		log.Printf("Error validating JWT: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error validating JWT: %s\n", err))
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		log.Printf("Invalid user ID in JWT subject: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid user ID in JWT subject")
		return
	}

	decoder := json.NewDecoder(r.Body)
	userBase := CreateUserStruct{}
	err = decoder.Decode(&userBase)
	if err != nil {
		log.Printf("Error decoding parameters: %s\n", err)
		code := http.StatusBadRequest
		msg := fmt.Sprintf("Error decoding parameters: %s\n", err)
		respondWithError(w, code, msg)
		return
	}

	hashedPassword, err := auth.HashPassword(userBase.Password)
	if err != nil {
		log.Printf("Error hashing password: %s\n", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error hashing password: %s\n", err))
		return
	}

	params := database.UpdateUserParams{
		ID:             userID,
		Email:          userBase.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.dbQueries.UpdateUser(r.Context(), params)
	if err != nil {
		log.Printf("Failed to update user: %s\n", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update user. user: %s may not exist\n", params.Email))
		return
	}

	returnUser := User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	log.Printf("Updated user: %s\n", user.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(returnUser)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	type chirpRequest struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s\n", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error decoding parameters: %s\n", err))
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer token: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error getting bearer token: %s\n", err))
		return
	}

	claims, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		log.Printf("Error validating JWT: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error validating JWT: %s\n", err))
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		log.Printf("Invalid user ID in JWT subject: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid user ID in JWT subject")
		return
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Failed to create chirp: %s\n", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp.\n")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chirp)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chirp)
	return

}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer token: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error getting bearer token: %s\n", err))
		return
	}

	claims, err := auth.ValidateJWT(tokenString, cfg.tokenSecret)
	if err != nil {
		log.Printf("Error validating JWT: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error validating JWT: %s\n", err))
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		log.Printf("Invalid user ID in token: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	chirpID_string := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpID_string)
	if err != nil {
		log.Printf("Failed to parse chirp UUID: %s\n", err)
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("Failed to get chirp: %s\n", err)
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Forbidden")
		return
	}

	err = cfg.dbQueries.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		log.Printf("Failed to delete chirp: %s\n", err)
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Failed to delete chirp",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return

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

		token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, time.Hour)
		if err != nil {
			log.Printf("Error creating JWT: %s\n", err)
			code := 500
			msg := fmt.Sprintf("Error creating JWT: %s\n", err)
			respondWithError(w, code, msg)
			return
		}

		refreshToken := auth.MakeRefreshToken()
		_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
			Token:     refreshToken,
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(time.Hour),
		})
		if err != nil {
			log.Printf("Error creating refresh token: %s\n", err)
			code := 500
			msg := fmt.Sprintf("Error creating refresh token: %s\n", err)
			respondWithError(w, code, msg)
			return
		}

		respUser := User{
			ID:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			Token:        token,
			RefreshToken: refreshToken,
			IsChirpyRed:  user.IsChirpyRed,
		}

		log.Printf("User %s logged in successfully\n", user.Email)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
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

func (cfg *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting refresh token: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid refresh token")
		return
	}

	user, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("Invalid refresh token: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.tokenSecret, time.Hour)
	if err != nil {
		log.Printf("Error creating access token: %s\n", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating access token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(refreshResponse{
		Token: accessToken,
	})
}

func (cfg *apiConfig) revokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting refresh token: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid refresh token")
		return
	}

	_, err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("Error revoking refresh token: %s\n", err)
		respondWithError(w, http.StatusInternalServerError, "Error revoking refresh token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) polkaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("Error getting API key: %s\n", err)
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid API key")
		return
	}

	if apiKey != cfg.polkaKey {
		log.Printf("Invalid API key: %s\n", apiKey)
		respondWithError(w, http.StatusUnauthorized, "Invalid API key")
		return
	}

	decoder := json.NewDecoder(r.Body)
	webhookObj := WebhookObject{}
	err = decoder.Decode(&webhookObj)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid webhook payload")
		return
	}

	if webhookObj.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(webhookObj.Data.UserId)
	if err != nil {
		log.Printf("Error parsing user ID: %s\n", err)
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error parsing user ID: %s\n", err))
		return
	}

	_, err = cfg.dbQueries.UpgradeToRed(r.Context(), userID)

	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	if err != nil {
		log.Printf("Error upgrading user to red: %s\n", err)
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Error upgrading user to red",
		)
		return
	}

	log.Printf("User %s upgraded to Chirpy Red successfully\n", userID)
	w.WriteHeader(http.StatusNoContent)
}
