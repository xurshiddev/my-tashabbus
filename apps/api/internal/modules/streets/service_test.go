package streets

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/mfys"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func TestStreetPermissionsAndValidation(t *testing.T) {
	ctx := context.Background()
	service, mfyService, userService := testStreetService()
	admin := users.User{ID: uuid.New(), Role: users.RoleSuperAdmin}
	mfy, err := mfyService.Create(ctx, admin, mfys.CreateMFYInput{Name: "Bogiston"})
	if err != nil {
		t.Fatalf("create mfy: %v", err)
	}
	chairman := users.User{ID: uuid.New(), Role: users.RoleMFYChairman, MFYID: &mfy.ID}

	if _, err := service.Create(ctx, admin, CreateStreetInput{MFYID: mfy.ID, Name: "Mustaqillik", PlannedHouseholdsCount: 1}); err != nil {
		t.Fatalf("super admin create street: %v", err)
	}
	if _, err := service.Create(ctx, chairman, CreateStreetInput{MFYID: mfy.ID, Name: "Navruz", PlannedHouseholdsCount: 1}); err != nil {
		t.Fatalf("chairman create own street: %v", err)
	}
	if _, err := service.Create(ctx, chairman, CreateStreetInput{MFYID: uuid.New(), Name: "Other", PlannedHouseholdsCount: 1}); err != ErrForbidden {
		t.Fatalf("expected forbidden for another mfy, got %v", err)
	}
	if _, err := service.Create(ctx, admin, CreateStreetInput{MFYID: mfy.ID, PlannedHouseholdsCount: 1}); err != ErrNameRequired {
		t.Fatalf("expected name required, got %v", err)
	}
	if _, err := service.Create(ctx, admin, CreateStreetInput{MFYID: mfy.ID, Name: "Bad", PlannedHouseholdsCount: -1}); err != ErrPlannedHouseholdsCount {
		t.Fatalf("expected planned count error, got %v", err)
	}
	if _, err := service.Create(ctx, admin, CreateStreetInput{MFYID: mfy.ID, Name: "Mustaqillik", PlannedHouseholdsCount: 1}); err != ErrDuplicateStreetName {
		t.Fatalf("expected duplicate street error, got %v", err)
	}
	_ = userService
}

func TestStreetLeaderAssignmentRules(t *testing.T) {
	ctx := context.Background()
	service, mfyService, userService := testStreetService()
	admin := users.User{ID: uuid.New(), Role: users.RoleSuperAdmin}
	mfy, _ := mfyService.Create(ctx, admin, mfys.CreateMFYInput{Name: "Bogiston"})
	otherMFY, _ := mfyService.Create(ctx, admin, mfys.CreateMFYInput{Name: "Other"})
	street, _ := service.Create(ctx, admin, CreateStreetInput{MFYID: mfy.ID, Name: "Mustaqillik"})
	leader, _ := userService.Create(ctx, users.CreateUserInput{FullName: "Leader", Role: users.RoleStreetLeader, MFYID: &mfy.ID})
	responsible, _ := userService.Create(ctx, users.CreateUserInput{FullName: "Responsible", Role: users.RoleResponsiblePerson, MFYID: &mfy.ID})
	otherLeader, _ := userService.Create(ctx, users.CreateUserInput{FullName: "Other", Role: users.RoleStreetLeader, MFYID: &otherMFY.ID})

	if _, err := service.AssignLeader(ctx, admin, street.ID, responsible.ID); err != ErrStreetLeaderRoleRequired {
		t.Fatalf("expected role error, got %v", err)
	}
	if _, err := service.AssignLeader(ctx, admin, street.ID, otherLeader.ID); err != ErrStreetLeaderWrongMFY {
		t.Fatalf("expected wrong mfy error, got %v", err)
	}
	first, err := service.AssignLeader(ctx, admin, street.ID, leader.ID)
	if err != nil {
		t.Fatalf("assign leader: %v", err)
	}
	replacement, _ := userService.Create(ctx, users.CreateUserInput{FullName: "Replacement", Role: users.RoleStreetLeader, MFYID: &mfy.ID})
	second, err := service.AssignLeader(ctx, admin, street.ID, replacement.ID)
	if err != nil {
		t.Fatalf("reassign leader: %v", err)
	}
	if first.ID == second.ID {
		t.Fatalf("expected new assignment on reassign")
	}
	if _, err := service.Get(ctx, replacement, street.ID); err != nil {
		t.Fatalf("assigned leader should view street: %v", err)
	}
	if _, err := service.Get(ctx, leader, street.ID); err != ErrForbidden {
		t.Fatalf("old leader should be forbidden, got %v", err)
	}
}

func testStreetService() (*Service, *mfys.Service, *users.Service) {
	userService := users.NewService(users.NewMemoryStore())
	mfyService := mfys.NewService(mfys.NewMemoryStore(), userService)
	return NewService(NewMemoryStore(), mfyService, userService), mfyService, userService
}
