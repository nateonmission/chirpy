package auth

import "testing"

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