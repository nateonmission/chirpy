package main

import (
	"sync/atomic"
	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"time"

)



type apiConfig struct {
	fileServerHits atomic.Int32
	platform string
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


type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}