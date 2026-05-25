package auth

import (
	"github.com/alexedwards/argon2id"
	"fmt"
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

