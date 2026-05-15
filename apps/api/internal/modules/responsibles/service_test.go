package responsibles

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/households"
	"github.com/my-tashabbus/api/internal/modules/mfys"
	"github.com/my-tashabbus/api/internal/modules/streets"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func TestResponsibleAssignmentRules(t *testing.T) {
	ctx := context.Background()
	service, householdService, streetService, _, userService, street := testResponsibleService(t)
	admin := users.User{ID: uuid.New(), Role: users.RoleSuperAdmin}
	responsible, _ := userService.Create(ctx, users.CreateUserInput{FullName: "Responsible", Role: users.RoleResponsiblePerson})
	leader, _ := userService.Create(ctx, users.CreateUserInput{FullName: "Leader", Role: users.RoleStreetLeader, MFYID: &street.MFYID})
	inactive, _ := userService.Create(ctx, users.CreateUserInput{FullName: "Inactive", Role: users.RoleResponsiblePerson, MFYID: &street.MFYID})
	inactive, _ = userService.Deactivate(ctx, inactive.ID)
	otherMFY := uuid.New()
	otherResponsible, _ := userService.Create(ctx, users.CreateUserInput{FullName: "Other", Role: users.RoleResponsiblePerson, MFYID: &otherMFY})

	if _, err := service.Create(ctx, admin, street.ID, CreateAssignmentInput{ResponsibleUserID: leader.ID, FromHouseNumber: "1", ToHouseNumber: "10"}); err != ErrResponsibleRoleRequired {
		t.Fatalf("expected role error, got %v", err)
	}
	if _, err := service.Create(ctx, admin, street.ID, CreateAssignmentInput{ResponsibleUserID: inactive.ID, FromHouseNumber: "1", ToHouseNumber: "10"}); err != ErrResponsibleUserInactive {
		t.Fatalf("expected inactive error, got %v", err)
	}
	if _, err := service.Create(ctx, admin, street.ID, CreateAssignmentInput{ResponsibleUserID: otherResponsible.ID, FromHouseNumber: "1", ToHouseNumber: "10"}); err != ErrResponsibleWrongMFY {
		t.Fatalf("expected wrong mfy error, got %v", err)
	}
	if _, err := service.Create(ctx, admin, street.ID, CreateAssignmentInput{ResponsibleUserID: responsible.ID, FromHouseNumber: "20", ToHouseNumber: "10"}); err != ErrInvalidHouseRange {
		t.Fatalf("expected invalid range error, got %v", err)
	}

	household, _ := householdService.Create(ctx, admin, street.ID, households.CreateHouseholdInput{HouseNumber: "5", TotalNumbers: 2, Status: households.StatusNew})
	assignment, err := service.Create(ctx, admin, street.ID, CreateAssignmentInput{ResponsibleUserID: responsible.ID, FromHouseNumber: "1", ToHouseNumber: "10"})
	if err != nil {
		t.Fatalf("assign responsible: %v", err)
	}
	if assignment.ResponsibleUserID != responsible.ID {
		t.Fatalf("expected responsible id")
	}
	updated, err := householdService.Get(ctx, admin, household.ID)
	if err != nil {
		t.Fatalf("get household: %v", err)
	}
	if updated.AssignedResponsibleUserID == nil || *updated.AssignedResponsibleUserID != responsible.ID {
		t.Fatalf("expected household assigned by numeric range")
	}
	assignedUser, _ := userService.GetByID(ctx, responsible.ID)
	if assignedUser.MFYID == nil || *assignedUser.MFYID != street.MFYID {
		t.Fatalf("expected null mfy responsible user to be assigned to street mfy")
	}
	_, _ = streetService, assignment
}

func TestStreetLeaderResponsibleAssignmentScope(t *testing.T) {
	ctx := context.Background()
	service, _, streetService, _, userService, street := testResponsibleService(t)
	admin := users.User{ID: uuid.New(), Role: users.RoleSuperAdmin}
	leader, _ := userService.Create(ctx, users.CreateUserInput{FullName: "Leader", Role: users.RoleStreetLeader, MFYID: &street.MFYID})
	if _, err := streetService.AssignLeader(ctx, admin, street.ID, leader.ID); err != nil {
		t.Fatalf("assign leader: %v", err)
	}
	responsible, _ := userService.Create(ctx, users.CreateUserInput{FullName: "Responsible", Role: users.RoleResponsiblePerson, MFYID: &street.MFYID})

	if _, err := service.Create(ctx, leader, street.ID, CreateAssignmentInput{ResponsibleUserID: responsible.ID, FromHouseNumber: "1", ToHouseNumber: "10"}); err != nil {
		t.Fatalf("assigned street leader should assign responsible: %v", err)
	}
	if _, err := service.Create(ctx, users.User{ID: uuid.New(), Role: users.RoleStreetLeader}, street.ID, CreateAssignmentInput{ResponsibleUserID: responsible.ID, FromHouseNumber: "1", ToHouseNumber: "10"}); err != ErrForbidden {
		t.Fatalf("unassigned street leader should be forbidden, got %v", err)
	}
}

func testResponsibleService(t *testing.T) (*Service, *households.Service, *streets.Service, *mfys.Service, *users.Service, streets.Street) {
	t.Helper()
	ctx := context.Background()
	userService := users.NewService(users.NewMemoryStore())
	mfyService := mfys.NewService(mfys.NewMemoryStore(), userService)
	streetService := streets.NewService(streets.NewMemoryStore(), mfyService, userService)
	admin := users.User{ID: uuid.New(), Role: users.RoleSuperAdmin}
	mfy, err := mfyService.Create(ctx, admin, mfys.CreateMFYInput{Name: "Bogiston"})
	if err != nil {
		t.Fatalf("create mfy: %v", err)
	}
	street, err := streetService.Create(ctx, admin, streets.CreateStreetInput{MFYID: mfy.ID, Name: "Mustaqillik"})
	if err != nil {
		t.Fatalf("create street: %v", err)
	}
	householdService := households.NewService(households.NewMemoryStore(), streetService)
	service := NewService(NewMemoryStore(), streetService, userService, householdService)
	return service, householdService, streetService, mfyService, userService, street
}
