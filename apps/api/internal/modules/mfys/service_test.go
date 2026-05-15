package mfys

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func TestSuperAdminCanCreateMFY(t *testing.T) {
	service, current := testService(t, users.RoleSuperAdmin, nil)
	mfy, err := service.Create(context.Background(), current, CreateMFYInput{Name: "Bogiston"})
	if err != nil {
		t.Fatalf("create mfy: %v", err)
	}
	if mfy.Name != "Bogiston" {
		t.Fatalf("expected Bogiston, got %s", mfy.Name)
	}
}

func TestCreateMFYValidation(t *testing.T) {
	service, current := testService(t, users.RoleSuperAdmin, nil)
	if _, err := service.Create(context.Background(), current, CreateMFYInput{}); err != ErrNameRequired {
		t.Fatalf("expected ErrNameRequired, got %v", err)
	}
	negative := int32(-1)
	if _, err := service.Create(context.Background(), current, CreateMFYInput{Name: "A", TargetVotes: &negative}); err != ErrTargetVotesNegative {
		t.Fatalf("expected ErrTargetVotesNegative, got %v", err)
	}
}

func TestMFYChairmanCannotCreateMFY(t *testing.T) {
	service, current := testService(t, users.RoleMFYChairman, nil)
	if _, err := service.Create(context.Background(), current, CreateMFYInput{Name: "Bogiston"}); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestMFYChairmanCanAccessOwnMFYOnly(t *testing.T) {
	service, admin := testService(t, users.RoleSuperAdmin, nil)
	ctx := context.Background()
	own, err := service.Create(ctx, admin, CreateMFYInput{Name: "Own"})
	if err != nil {
		t.Fatalf("create own mfy: %v", err)
	}
	other, err := service.Create(ctx, admin, CreateMFYInput{Name: "Other"})
	if err != nil {
		t.Fatalf("create other mfy: %v", err)
	}
	chairman := users.User{Role: users.RoleMFYChairman, MFYID: &own.ID}
	if _, err := service.Get(ctx, chairman, own.ID); err != nil {
		t.Fatalf("expected own mfy access, got %v", err)
	}
	if _, err := service.Get(ctx, chairman, other.ID); err != ErrForbidden {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func testService(t *testing.T, role users.Role, mfyID *uuid.UUID) (*Service, users.User) {
	t.Helper()
	userService := users.NewService(users.NewMemoryStore())
	return NewService(NewMemoryStore(), userService), users.User{Role: role, MFYID: mfyID}
}
