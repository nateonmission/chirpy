package main

import (
	"net/http"
	"fmt"
)


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		fmt.Printf("Counter: %d\n", cfg.fileServerHits.Load())

		next.ServeHTTP(w, r)
	})
}
