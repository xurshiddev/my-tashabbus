package users

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateUserArgsUseSQLNullForOptionalFields(t *testing.T) {
	args := createUserArgs(CreateUserInput{
		FullName: "Dev Super Admin",
		Role:     RoleSuperAdmin,
	})

	if args[1] != nil {
		t.Fatalf("expected nil phone arg, got %#v", args[1])
	}
	if args[2] != nil {
		t.Fatalf("expected nil telegram id arg, got %#v", args[2])
	}
	if args[3] != nil {
		t.Fatalf("expected nil telegram username arg, got %#v", args[3])
	}
	if args[4] != string(RoleSuperAdmin) {
		t.Fatalf("expected string role arg, got %#v", args[4])
	}
	if args[5] != nil {
		t.Fatalf("expected nil mfy id arg, got %#v", args[5])
	}
}

func TestCreateUserArgsEncodeOptionalValues(t *testing.T) {
	phone := "+998901234567"
	telegramID := int64(123456789)
	username := "devuser"
	mfyID := uuid.New()

	args := createUserArgs(CreateUserInput{
		FullName:         "Dev Super Admin",
		Phone:            &phone,
		TelegramID:       &telegramID,
		TelegramUsername: &username,
		Role:             RoleSuperAdmin,
		MFYID:            &mfyID,
	})

	if args[1] != phone {
		t.Fatalf("expected phone value, got %#v", args[1])
	}
	if args[2] != telegramID {
		t.Fatalf("expected telegram id value, got %#v", args[2])
	}
	if args[3] != username {
		t.Fatalf("expected username value, got %#v", args[3])
	}
	if args[5] != mfyID.String() {
		t.Fatalf("expected mfy id string, got %#v", args[5])
	}
}

func TestSetTelegramIdentityArgsUseSQLNullForOptionalUsername(t *testing.T) {
	id := uuid.New()
	args := setTelegramIdentityArgs(id, SetTelegramIdentityInput{TelegramID: 123456789})

	if args[0] != id {
		t.Fatalf("expected user id arg, got %#v", args[0])
	}
	if args[1] != int64(123456789) {
		t.Fatalf("expected telegram id arg, got %#v", args[1])
	}
	if args[2] != nil {
		t.Fatalf("expected nil telegram username arg, got %#v", args[2])
	}
}
