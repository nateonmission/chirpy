package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("got status %d, want %d", res.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(res.Body)
	if string(body) != "200 OK" {
		t.Fatalf("got body %q, want %q", string(body), "200 OK")
	}
}