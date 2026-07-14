package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"

	_ "github.com/lib/pq"
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

func TestUpdateUserHandlerMissingToken(t *testing.T) {
	cfg := apiConfig{
		tokenSecret: "test-secret",
	}

	body := `{
		"email": "new@example.com",
		"password": "new-password"
	}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/users",
		strings.NewReader(body),
	)
	rec := httptest.NewRecorder()

	cfg.updateUserHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"got status %d, want %d",
			res.StatusCode,
			http.StatusUnauthorized,
		)
	}
}

func TestUpdateUserHandlerInvalidToken(t *testing.T) {
	cfg := apiConfig{
		tokenSecret: "test-secret",
	}

	body := `{
		"email": "new@example.com",
		"password": "new-password"
	}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/users",
		strings.NewReader(body),
	)
	req.Header.Set(
		"Authorization",
		"Bearer this-is-not-a-valid-token",
	)

	rec := httptest.NewRecorder()

	cfg.updateUserHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"got status %d, want %d",
			res.StatusCode,
			http.StatusUnauthorized,
		)
	}
}

func TestUpdateUserHandlerMissingTokenWithInvalidBody(t *testing.T) {
	cfg := apiConfig{
		tokenSecret: "test-secret",
	}

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/users",
		strings.NewReader(`{"email":`),
	)
	rec := httptest.NewRecorder()

	cfg.updateUserHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf(
			"got status %d, want %d",
			res.StatusCode,
			http.StatusUnauthorized,
		)
	}
}
