package main

import (
	"sync/atomic"
	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"time"
	"github.com/chirpy/src/internal/database"

)



type apiConfig struct {
	fileServerHits atomic.Int32
	platform string
	dbQueries *database.Queries
}

type chirpToValidate struct {
	Body string `json:"body"`
}

type chirpError struct {
	Error string `json:"error"`
}

type chirpValid struct {
	Valid bool `json:"valid"`
	CleanedBody string `json:"cleaned_body"`
}


type createUserParam struct {
	Email	string `json:"email"`
}


type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

// type receivedChirp struct {
// 	Body	string	`json:"body"`
// 	User_ID	uusid.UUID	`json:"user_id"`
// }

type CreateChirpParams struct {
	Body   string	`json:"body"`
	UserID uuid.UUID	`json:"user_id"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body	string    `json:"body"`
	User_ID	uuid.UUID	`json:"user_id"`
}