package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := 24 * time.Hour

	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("MakeJWT() error = %v, want nil", err)
	}
	if token == "" {
		t.Error("MakeJWT() token is empty, want non-empty")
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := 24 * time.Hour

	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("MakeJWT() error = %v, want nil", err)
	}
	if token == "" {
		t.Error("MakeJWT() token is empty, want non-empty")
	}

	validatedUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Errorf("ValidateJWT() error = %v, want nil", err)
	}
	if validatedUserID != userID {
		t.Errorf("ValidateJWT() userID = %v, want %v", validatedUserID, userID)
	}
}
