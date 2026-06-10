package auth

import (
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"fmt"
	"time"
	"strings"
	"net/http"
)

func HashPassword(password string) (string, error) {
	hashed_password, err := argon2id.CreateHash(password, argon2id.DefaultParams);
	if err != nil {
		return fmt.Sprintf("Error creating password hash: %s", err), err
	}
	return hashed_password, nil;
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:  "chirpy-access",
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject: userID.String(),
	})

	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("Error signing JWT: %w", err)
	}

	return tokenString, nil
}

func ValidateJWT(tokenString string, tokenSecret string) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer("chirpy-access"),
	)

	if err != nil {
		return nil, fmt.Errorf("Error parsing JWT: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("Invalid JWT")
	}

	return claims, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header missing")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("Invalid Authorization header format")
	}
	trimmed := strings.TrimSpace(parts[1])
	if trimmed == "" {
		return "", fmt.Errorf("Bearer token is empty")
	}
	return trimmed, nil
}
