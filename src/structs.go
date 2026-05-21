package main

import (
	"sync/atomic"
)



type apiConfig struct {
	fileServerHits atomic.Int32
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