package auth

import (
	"testing"
	"time"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
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