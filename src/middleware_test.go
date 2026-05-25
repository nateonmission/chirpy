package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
)


func TestMiddlewareMetricsInc(t *testing.T) {
	cfg := &apiConfig{}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := cfg.middlewareMetricsInc(next)

	req := httptest.NewRequest(http.MethodGet, "/app/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if cfg.fileServerHits.Load() != 1 {
		t.Fatalf("got hits %d, want 1", cfg.fileServerHits.Load())
	}
}