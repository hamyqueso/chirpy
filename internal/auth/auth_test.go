package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT_Succeeds(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	gotID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT error: %v", err)
	}
	if gotID != userID {
		t.Fatalf("want %s, got %s", userID, gotID)
	}
}

func TestValidateJWT_ExpiredTokenFails(t *testing.T) {
	userID := uuid.New()
	secret := "test_secret"
	expiresIn := 0

	token, err := MakeJWT(userID, secret, time.Duration(expiresIn))
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	gotID, err := ValidateJWT(token, secret)
	if err == nil {
		t.Fatalf("expected error for expired token, got none (id = %s)", gotID)
	}
	if gotID != uuid.Nil {
		t.Fatalf("want zero UUID, got %s", gotID)
	}
}

func TestValidateJWT_WrongSecretFails(t *testing.T) {
	userID := uuid.New()
	signSecret := "right-secret"
	checkSecret := "wrong-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, signSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	gotID, err := ValidateJWT(token, checkSecret)
	if err == nil {
		t.Fatalf("expected error for wrong secret, got none (id=%s)", gotID)
	}
	if gotID != uuid.Nil {
		t.Fatalf("want zero UUID, got %s", gotID)
	}
}
