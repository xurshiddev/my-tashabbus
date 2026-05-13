package users

import (
	"context"
	"testing"
)

func TestIsValidRole(t *testing.T) {
	if !IsValidRole(RoleSuperAdmin) {
		t.Fatalf("expected SUPER_ADMIN to be valid")
	}
	if IsValidRole(Role("ADMIN")) {
		t.Fatalf("expected ADMIN to be invalid")
	}
}

func TestServiceRejectsInvalidRole(t *testing.T) {
	service := NewService(NewMemoryStore())
	_, err := service.Create(context.Background(), CreateUserInput{
		FullName: "Ali",
		Role:     Role("ADMIN"),
	})
	if err != ErrInvalidRole {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}
