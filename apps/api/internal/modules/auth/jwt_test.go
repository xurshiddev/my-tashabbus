package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func TestTokenManagerGenerateAndParse(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Hour)
	user := users.User{ID: uuid.New(), Role: users.RoleSuperAdmin, IsActive: true}

	token, expiresIn, err := manager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if expiresIn != 3600 {
		t.Fatalf("expected expires_in 3600, got %d", expiresIn)
	}

	claims, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.UserID != user.ID {
		t.Fatalf("expected user id %s, got %s", user.ID, claims.UserID)
	}
	if claims.Role != users.RoleSuperAdmin {
		t.Fatalf("expected role %s, got %s", users.RoleSuperAdmin, claims.Role)
	}
}

func TestTokenManagerRejectsInvalidSignature(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Hour)
	other := NewTokenManager("other-secret", time.Hour)
	token, _, err := manager.GenerateAccessToken(users.User{ID: uuid.New(), Role: users.RoleSuperAdmin})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	if _, err := other.ParseAccessToken(token); err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestTokenManagerRejectsExpiredToken(t *testing.T) {
	manager := NewTokenManager("test-secret", -time.Hour)
	token, _, err := manager.GenerateAccessToken(users.User{ID: uuid.New(), Role: users.RoleSuperAdmin})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	if _, err := manager.ParseAccessToken(token); err != ErrInvalidToken && err != ErrExpiredToken {
		t.Fatalf("expected expired or invalid token, got %v", err)
	}
}
