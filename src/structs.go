package main

import (
	"github.com/chirpy/src/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"sync/atomic"
	"time"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	platform       string
	dbQueries      *database.Queries
	loggedinUser   uuid.UUID
	tokenSecret    string
	polkaKey       string
}

type chirpToValidate struct {
	Body string `json:"body"`
}

type chirpError struct {
	Error string `json:"error"`
}

type chirpValid struct {
	Valid       bool   `json:"valid"`
	CleanedBody string `json:"cleaned_body"`
}

type CreateUserStruct struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

// type receivedChirp struct {
// 	Body	string	`json:"body"`
// 	User_ID	uusid.UUID	`json:"user_id"`
// }

type CreateChirpParams struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	User_ID   uuid.UUID `json:"user_id"`
}

type refreshResponse struct {
	Token string `json:"token"`
}

type updateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type WebhookObject struct {
	Event string      `json:"event,omitempty"`
	Data  WebhookData `json:"data,omitempty"`
}

type WebhookData struct {
	UserId string `json:"user_id,omitempty"`
}
