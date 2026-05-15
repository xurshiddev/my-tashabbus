package streets

import (
	"time"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type Street struct {
	ID                     uuid.UUID  `json:"id"`
	MFYID                  uuid.UUID  `json:"mfy_id"`
	Name                   string     `json:"name"`
	PlannedHouseholdsCount int32      `json:"planned_households_count"`
	Notes                  *string    `json:"notes"`
	IsActive               bool       `json:"is_active"`
	CreatedByUserID        *uuid.UUID `json:"created_by_user_id"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type StreetLeaderAssignment struct {
	ID               uuid.UUID  `json:"id"`
	StreetID         uuid.UUID  `json:"street_id"`
	UserID           uuid.UUID  `json:"user_id"`
	AssignedByUserID *uuid.UUID `json:"assigned_by_user_id"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type AssignmentWithUser struct {
	Assignment *StreetLeaderAssignment `json:"assignment"`
	User       *users.User             `json:"user"`
}

type CreateStreetInput struct {
	MFYID                  uuid.UUID  `json:"-"`
	Name                   string     `json:"name"`
	PlannedHouseholdsCount int32      `json:"planned_households_count"`
	Notes                  *string    `json:"notes"`
	CreatedByUserID        *uuid.UUID `json:"-"`
}

type UpdateStreetInput struct {
	Name                   string  `json:"name"`
	PlannedHouseholdsCount int32   `json:"planned_households_count"`
	Notes                  *string `json:"notes"`
	IsActive               bool    `json:"is_active"`
}
