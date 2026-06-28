package auth

import (
	"testing"
	"time"
	"net/http"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
	"encoding/hex"
)

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	password := "correct horse battery staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash == password {
		t.Fatal("hash should not equal the plain password")
	}

	ok, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash returned error: %v", err)
	}

	if !ok {
		t.Fatal("expected password to match hash")
	}
}

func TestCheckPasswordHashRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("right-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	ok, err := CheckPasswordHash("wrong-password", hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash returned error: %v", err)
	}

	if ok {
		t.Fatal("expected wrong password to fail")
	}
}

func TestMakeJWTAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "super-secret-test-key"
	expiresIn := time.Hour

	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error making JWT, got: %v", err)
	}

	if tokenString == "" {
		t.Fatal("expected token string, got empty string")
	}

	claims, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("expected no error validating JWT, got: %v", err)
	}

	if claims.Subject != userID.String() {
		t.Errorf("expected subject %q, got %q", userID.String(), claims.Subject)
	}

	if claims.Issuer != "chirpy-access" {
		t.Errorf("expected issuer %q, got %q", "chirpy-access", claims.Issuer)
	}

	if claims.IssuedAt == nil {
		t.Error("expected IssuedAt to be set")
	}

	if claims.ExpiresAt == nil {
		t.Error("expected ExpiresAt to be set")
	}
}

func TestValidateJWTWrongSecret(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "correct-secret"
	wrongSecret := "wrong-secret"

	tokenString, err := MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		t.Fatalf("expected no error making JWT, got: %v", err)
	}

	_, err = ValidateJWT(tokenString, wrongSecret)
	if err == nil {
		t.Fatal("expected error validating JWT with wrong secret, got nil")
	}
}

func TestValidateJWTExpiredToken(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "super-secret-test-key"

	tokenString, err := MakeJWT(userID, tokenSecret, -time.Hour)
	if err != nil {
		t.Fatalf("expected no error making expired JWT, got: %v", err)
	}

	_, err = ValidateJWT(tokenString, tokenSecret)
	if err == nil {
		t.Fatal("expected error validating expired JWT, got nil")
	}
}

func TestValidateJWTMalformedToken(t *testing.T) {
	tokenSecret := "super-secret-test-key"

	_, err := ValidateJWT("not.a.valid.jwt", tokenSecret)
	if err == nil {
		t.Fatal("expected error validating malformed JWT, got nil")
	}
}

func TestValidateJWTWrongIssuer(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "super-secret-test-key"
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "wrong-issuer",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		Subject:   userID.String(),
	})

	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		t.Fatalf("expected no error signing JWT, got: %v", err)
	}

	_, err = ValidateJWT(tokenString, tokenSecret)
	if err == nil {
		t.Fatal("expected error validating JWT with wrong issuer, got nil")
	}
}

func TestValidateJWTRejectsWrongSigningMethod(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	privateKey := []byte("different-secret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		Subject:   userID.String(),
	})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("expected no error signing JWT, got: %v", err)
	}

	_, err = ValidateJWT(tokenString, string(privateKey))
	if err == nil {
		t.Fatal("expected error validating JWT with wrong signing method, got nil")
	}
}


func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		headers     http.Header
		wantToken   string
		wantErr     bool
	}{
		{
			name: "valid bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer abc123"},
			},
			wantToken: "abc123",
			wantErr:   false,
		},
		{
			name: "valid bearer token with extra spaces around token",
			headers: http.Header{
				"Authorization": []string{"Bearer    abc123   "},
			},
			wantToken: "abc123",
			wantErr:   false,
		},
		{
			name:      "missing authorization header",
			headers:   http.Header{},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "authorization header without bearer scheme",
			headers: http.Header{
				"Authorization": []string{"abc123"},
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "wrong scheme",
			headers: http.Header{
				"Authorization": []string{"Basic abc123"},
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "lowercase bearer is invalid",
			headers: http.Header{
				"Authorization": []string{"bearer abc123"},
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "bearer with no token returns empty token",
			headers: http.Header{
				"Authorization": []string{"Bearer "},
			},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(tt.headers)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if gotToken != tt.wantToken {
				t.Errorf("expected token %q, got %q", tt.wantToken, gotToken)
			}
		})
	}
}


func TestMakeRefreshTokenReturnsString(t *testing.T) {
	token := MakeRefreshToken()

	if token == "" {
		t.Fatal("expected refresh token, got empty string")
	}
}

func TestMakeRefreshTokenLength(t *testing.T) {
	token := MakeRefreshToken()

	// 32 random bytes encoded as hex = 64 characters
	if len(token) != 64 {
		t.Fatalf("expected token length 64, got %d", len(token))
	}
}

func TestMakeRefreshTokenIsValidHex(t *testing.T) {
	token := MakeRefreshToken()

	decoded, err := hex.DecodeString(token)
	if err != nil {
		t.Fatalf("expected valid hex string, got error: %v", err)
	}

	if len(decoded) != 32 {
		t.Fatalf("expected decoded token to be 32 bytes, got %d", len(decoded))
	}
}

func TestMakeRefreshTokenGeneratesDifferentTokens(t *testing.T) {
	token1 := MakeRefreshToken()
	token2 := MakeRefreshToken()

	if token1 == token2 {
		t.Fatal("expected two refresh tokens to be different")
	}
}

func TestMakeRefreshTokenGeneratesManyUniqueTokens(t *testing.T) {
	seen := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		token := MakeRefreshToken()

		if seen[token] {
			t.Fatalf("duplicate token generated at iteration %d: %s", i, token)
		}

		seen[token] = true
	}
}

func TestMakeRefreshTokenLengthAndHex(t *testing.T) {
	token := MakeRefreshToken()

	if len(token) != 64 {
		t.Fatalf("expected refresh token length 64, got %d", len(token))
	}

	decoded, err := hex.DecodeString(token)
	if err != nil {
		t.Fatalf("expected valid hex token, got error: %v", err)
	}

	if len(decoded) != 32 {
		t.Fatalf("expected decoded token to be 32 bytes, got %d", len(decoded))
	}
}

func TestMakeRefreshTokenGeneratesUniqueTokens(t *testing.T) {
	seen := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		token := MakeRefreshToken()

		if seen[token] {
			t.Fatalf("duplicate refresh token generated: %s", token)
		}

		seen[token] = true
	}
}

func TestMakeJWTStoresUserIDInSubject(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"

	tokenString, err := MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		t.Fatalf("expected no error creating JWT, got %v", err)
	}

	claims, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("expected valid JWT, got error: %v", err)
	}

	if claims.Subject != userID.String() {
		t.Fatalf("expected subject %s, got %s", userID.String(), claims.Subject)
	}
}

func TestMakeJWTUsesExpirationDuration(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"

	before := time.Now()

	tokenString, err := MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		t.Fatalf("expected no error creating JWT, got %v", err)
	}

	claims, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("expected valid JWT, got error: %v", err)
	}

	if claims.ExpiresAt == nil {
		t.Fatal("expected JWT to have exp claim")
	}

	minExpected := before.Add(time.Hour - 5*time.Second)
	maxExpected := time.Now().Add(time.Hour + 5*time.Second)

	if claims.ExpiresAt.Time.Before(minExpected) || claims.ExpiresAt.Time.After(maxExpected) {
		t.Fatalf(
			"expected expiration about 1 hour from now, got %s",
			claims.ExpiresAt.Time,
		)
	}
}

func TestValidateJWTRejectsExpiredToken(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"

	tokenString, err := MakeJWT(userID, tokenSecret, -time.Hour)
	if err != nil {
		t.Fatalf("expected no error creating expired JWT, got %v", err)
	}

	_, err = ValidateJWT(tokenString, tokenSecret)
	if err == nil {
		t.Fatal("expected expired JWT to be rejected")
	}
}

func TestValidateJWTRejectsWrongSecret(t *testing.T) {
	userID := uuid.New()

	tokenString, err := MakeJWT(userID, "correct-secret", time.Hour)
	if err != nil {
		t.Fatalf("expected no error creating JWT, got %v", err)
	}

	_, err = ValidateJWT(tokenString, "wrong-secret")
	if err == nil {
		t.Fatal("expected JWT with wrong secret to be rejected")
	}
}