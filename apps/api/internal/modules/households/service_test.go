package households

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/mfys"
	"github.com/my-tashabbus/api/internal/modules/streets"
	"github.com/my-tashabbus/api/internal/modules/users"
)

func TestHouseholdPermissionsAndValidation(t *testing.T) {
	ctx := context.Background()
	service, streetService, _, userService, street := testHouseholdService(t)
	admin := users.User{ID: uuid.New(), Role: users.RoleSuperAdmin}
	chairman := users.User{ID: uuid.New(), Role: users.RoleMFYChairman, MFYID: &street.MFYID}
	leader, err := userService.Create(ctx, users.CreateUserInput{FullName: "Leader", Role: users.RoleStreetLeader, MFYID: &street.MFYID})
	if err != nil {
		t.Fatalf("create leader: %v", err)
	}
	if _, err := streetService.AssignLeader(ctx, admin, street.ID, leader.ID); err != nil {
		t.Fatalf("assign leader: %v", err)
	}
	responsible := users.User{ID: uuid.New(), Role: users.RoleResponsiblePerson, MFYID: &street.MFYID}

	if _, err := service.Create(ctx, admin, street.ID, CreateHouseholdInput{HouseNumber: "12", TotalNumbers: 5, Status: StatusNew}); err != nil {
		t.Fatalf("super admin create household: %v", err)
	}
	if _, err := service.Create(ctx, chairman, street.ID, CreateHouseholdInput{HouseNumber: "13", TotalNumbers: 5, Status: StatusNew}); err != nil {
		t.Fatalf("chairman create household: %v", err)
	}
	if _, err := service.Create(ctx, leader, street.ID, CreateHouseholdInput{HouseNumber: "14", TotalNumbers: 5, Status: StatusNew}); err != nil {
		t.Fatalf("street leader create household: %v", err)
	}
	if _, err := service.Create(ctx, responsible, street.ID, CreateHouseholdInput{HouseNumber: "15", TotalNumbers: 5, Status: StatusNew}); err != ErrForbidden {
		t.Fatalf("expected responsible create forbidden, got %v", err)
	}
	if _, err := service.Create(ctx, admin, street.ID, CreateHouseholdInput{TotalNumbers: 5, Status: StatusNew}); err != ErrHouseNumberRequired {
		t.Fatalf("expected house number required, got %v", err)
	}
	if _, err := service.Create(ctx, admin, street.ID, CreateHouseholdInput{HouseNumber: "16", TotalNumbers: -1, Status: StatusNew}); err != ErrInvalidCounts {
		t.Fatalf("expected count validation error, got %v", err)
	}
	if _, err := service.Create(ctx, admin, street.ID, CreateHouseholdInput{HouseNumber: "16", TotalNumbers: 1, VotedNumbers: 2, Status: StatusNew}); err != ErrInvalidCounts {
		t.Fatalf("expected voted count validation error, got %v", err)
	}
	if _, err := service.Create(ctx, admin, street.ID, CreateHouseholdInput{HouseNumber: "16", TotalNumbers: 1, Status: Status("BAD")}); err != ErrInvalidStatus {
		t.Fatalf("expected invalid status, got %v", err)
	}
	if _, err := service.Create(ctx, admin, street.ID, CreateHouseholdInput{HouseNumber: "12", TotalNumbers: 5, Status: StatusNew}); err != ErrDuplicateHousehold {
		t.Fatalf("expected duplicate household, got %v", err)
	}
}

func TestResponsiblePersonAccessAndUpdateLogs(t *testing.T) {
	ctx := context.Background()
	service, _, _, _, street := testHouseholdService(t)
	admin := users.User{ID: uuid.New(), Role: users.RoleSuperAdmin}
	responsibleID := uuid.New()
	responsible := users.User{ID: responsibleID, Role: users.RoleResponsiblePerson, MFYID: &street.MFYID}
	otherResponsible := users.User{ID: uuid.New(), Role: users.RoleResponsiblePerson, MFYID: &street.MFYID}

	household, err := service.Create(ctx, admin, street.ID, CreateHouseholdInput{HouseNumber: "20", TotalNumbers: 5, Status: StatusNew, AssignedResponsibleUserID: &responsibleID})
	if err != nil {
		t.Fatalf("create household: %v", err)
	}
	if _, err := service.Get(ctx, responsible, household.ID); err != nil {
		t.Fatalf("responsible should access assigned household: %v", err)
	}
	if _, err := service.Get(ctx, otherResponsible, household.ID); err != ErrForbidden {
		t.Fatalf("expected other responsible forbidden, got %v", err)
	}
	updated, err := service.Update(ctx, responsible, household.ID, UpdateHouseholdInput{HouseNumber: "20", TotalNumbers: 5, ContactedNumbers: 4, VotedNumbers: 3, Status: StatusPartiallyVoted})
	if err != nil {
		t.Fatalf("responsible update assigned household: %v", err)
	}
	if updated.VotedNumbers != 3 {
		t.Fatalf("expected updated voted count")
	}
	logs, err := service.Logs(ctx, admin, household.ID, 20, 0)
	if err != nil {
		t.Fatalf("list logs: %v", err)
	}
	if len(logs) == 0 {
		t.Fatalf("expected change logs")
	}
}

func testHouseholdService(t *testing.T) (*Service, *streets.Service, *mfys.Service, *users.Service, streets.Street) {
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
	return NewService(NewMemoryStore(), streetService), streetService, mfyService, userService, street
}
