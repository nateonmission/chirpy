package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
	"time"
	"errors"
	"database/sql/driver"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chirpy/src/internal/database"
	"github.com/chirpy/src/internal/auth"
	"github.com/google/uuid"
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

func TestDeleteChirpHandlerMissingToken(t *testing.T) {
	cfg := apiConfig{
		tokenSecret: "test-secret",
	}

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/chirps/"+uuid.NewString(),
		nil,
	)

	req.SetPathValue("chirpID", uuid.NewString())

	rec := httptest.NewRecorder()

	cfg.deleteChirpHandler(rec, req)

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

func TestDeleteChirpHandlerInvalidToken(t *testing.T) {
	cfg := apiConfig{
		tokenSecret: "test-secret",
	}

	chirpID := uuid.NewString()

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/chirps/"+chirpID,
		nil,
	)

	req.SetPathValue("chirpID", chirpID)
	req.Header.Set(
		"Authorization",
		"Bearer this-is-not-a-valid-token",
	)

	rec := httptest.NewRecorder()

	cfg.deleteChirpHandler(rec, req)

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

func TestDeleteChirpHandlerTokenSignedWithWrongSecret(t *testing.T) {
	userID := uuid.New()

	token, err := auth.MakeJWT(
		userID,
		"wrong-secret",
		time.Hour,
	)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	cfg := apiConfig{
		tokenSecret: "correct-secret",
	}

	chirpID := uuid.NewString()

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/chirps/"+chirpID,
		nil,
	)

	req.SetPathValue("chirpID", chirpID)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()

	cfg.deleteChirpHandler(rec, req)

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


func TestDeleteChirpHandlerInvalidChirpID(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := auth.MakeJWT(
		userID,
		secret,
		time.Hour,
	)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	cfg := apiConfig{
		tokenSecret: secret,
	}

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/chirps/not-a-uuid",
		nil,
	)

	req.SetPathValue("chirpID", "not-a-uuid")
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()

	cfg.deleteChirpHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(res.Body)

		t.Fatalf(
			"got status %d, want %d; body: %s",
			res.StatusCode,
			http.StatusNotFound,
			string(body),
		)
	}
}


func TestDeleteChirpHandlerUnauthorized(t *testing.T) {
	tests := []struct {
		name          string
		authorization string
	}{
		{
			name:          "missing token",
			authorization: "",
		},
		{
			name:          "invalid token",
			authorization: "Bearer invalid-token",
		},
		{
			name:          "missing bearer prefix",
			authorization: "invalid-token",
		},
		{
			name:          "empty bearer token",
			authorization: "Bearer ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := apiConfig{
				tokenSecret: "test-secret",
			}

			chirpID := uuid.NewString()

			req := httptest.NewRequest(
				http.MethodDelete,
				"/api/chirps/"+chirpID,
				nil,
			)

			req.SetPathValue("chirpID", chirpID)

			if test.authorization != "" {
				req.Header.Set(
					"Authorization",
					test.authorization,
				)
			}

			rec := httptest.NewRecorder()

			cfg.deleteChirpHandler(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf(
					"got status %d, want %d",
					rec.Code,
					http.StatusUnauthorized,
				)
			}
		})
	}
}


func setupPolkaTest(t *testing.T) (*apiConfig, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}

	cfg := &apiConfig{
		dbQueries: database.New(db),
		polkaKey:  "test-polka-key",
	}

	return cfg, mock
}

func TestPolkaWebhookIgnoresUnknownEvent(t *testing.T) {
	cfg, mock := setupPolkaTest(t)

	body := `{
		"event": "payment.failed",
		"data": {
			"user_id": "3311741c-680c-4546-99f3-fc9efac2036c"
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/polka/webhooks",
		strings.NewReader(body),
	)
	req.Header.Set("Authorization", "ApiKey test-polka-key")

	rec := httptest.NewRecorder()

	cfg.polkaWebhookHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusNoContent,
			rec.Code,
		)
	}

	if rec.Body.Len() != 0 {
		t.Errorf("expected empty response body, got %q", rec.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected database operation: %v", err)
	}
}

func TestPolkaWebhookUpgradesUser(t *testing.T) {
	cfg, mock := setupPolkaTest(t)

	userID := uuid.MustParse("3311741c-680c-4546-99f3-fc9efac2036c")
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id",
		"email",
		"created_at",
		"updated_at",
		"hashed_password",
		"is_chirpy_red",
	}).AddRow(
		userID,
		"test@example.com",
		now,
		now,
		"hashed-password",
		true,
	)

	mock.ExpectQuery(`UPDATE users`).
		WithArgs(userID).
		WillReturnRows(rows)

	body := `{
		"event": "user.upgraded",
		"data": {
			"user_id": "3311741c-680c-4546-99f3-fc9efac2036c"
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/polka/webhooks",
		strings.NewReader(body),
	)

	req.Header.Set("Authorization", "ApiKey test-polka-key")

	rec := httptest.NewRecorder()

	cfg.polkaWebhookHandler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf(
			"expected status %d, got %d; body: %s",
			http.StatusNoContent,
			rec.Code,
			rec.Body.String(),
		)
	}

	if rec.Body.Len() != 0 {
		t.Errorf("expected empty response body, got %q", rec.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("database expectations were not met: %v", err)
	}
}

func TestPolkaWebhookReturnsNotFound(t *testing.T) {
	cfg, mock := setupPolkaTest(t)

	userID := uuid.MustParse("3311741c-680c-4546-99f3-fc9efac2036c")

	mock.ExpectQuery(`UPDATE users`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	body := `{
		"event": "user.upgraded",
		"data": {
			"user_id": "3311741c-680c-4546-99f3-fc9efac2036c"
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/polka/webhooks",
		strings.NewReader(body),
	)

	req.Header.Set("Authorization", "ApiKey test-polka-key")

	rec := httptest.NewRecorder()

	cfg.polkaWebhookHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf(
			"expected status %d, got %d; body: %s",
			http.StatusNotFound,
			rec.Code,
			rec.Body.String(),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("database expectations were not met: %v", err)
	}
}

func TestPolkaWebhookRejectsInvalidUserID(t *testing.T) {
	cfg, mock := setupPolkaTest(t)

	body := `{
		"event": "user.upgraded",
		"data": {
			"user_id": "not-a-valid-uuid"
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/polka/webhooks",
		strings.NewReader(body),
	)

	req.Header.Set("Authorization", "ApiKey test-polka-key")

	rec := httptest.NewRecorder()

	cfg.polkaWebhookHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d; body: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected database operation: %v", err)
	}
}

func TestPolkaWebhookRejectsMalformedJSON(t *testing.T) {
	cfg, mock := setupPolkaTest(t)

	body := `{
		"event": "user.upgraded",
		"data":
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/polka/webhooks",
		strings.NewReader(body),
	)

	req.Header.Set("Authorization", "ApiKey test-polka-key")
	
	rec := httptest.NewRecorder()

	cfg.polkaWebhookHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status %d, got %d; body: %s",
			http.StatusBadRequest,
			rec.Code,
			rec.Body.String(),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected database operation: %v", err)
	}
}

func TestPolkaWebhookReturnsInternalServerError(t *testing.T) {
	cfg, mock := setupPolkaTest(t)

	userID := uuid.MustParse("3311741c-680c-4546-99f3-fc9efac2036c")

	mock.ExpectQuery(`UPDATE users`).
		WithArgs(userID).
		WillReturnError(errors.New("database connection failed"))

	body := `{
		"event": "user.upgraded",
		"data": {
			"user_id": "3311741c-680c-4546-99f3-fc9efac2036c"
		}
	}`



	req := httptest.NewRequest(
		http.MethodPost,
		"/api/polka/webhooks",
		strings.NewReader(body),
	)

	req.Header.Set("Authorization", "ApiKey test-polka-key")

	rec := httptest.NewRecorder()

	cfg.polkaWebhookHandler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf(
			"expected status %d, got %d; body: %s",
			http.StatusInternalServerError,
			rec.Code,
			rec.Body.String(),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("database expectations were not met: %v", err)
	}
}

// Prevents an unused-import error in case a version of sqlmock requires
// database/sql/driver values for custom matchers.
var _ driver.Value
