package responsibles

import (
	"time"

	"github.com/google/uuid"
)

type Assignment struct {
	ID                uuid.UUID  `json:"id"`
	StreetID          uuid.UUID  `json:"street_id"`
	ResponsibleUserID uuid.UUID  `json:"responsible_user_id"`
	AssignedByUserID  *uuid.UUID `json:"assigned_by_user_id"`
	FromHouseNumber   string     `json:"from_house_number"`
	ToHouseNumber     string     `json:"to_house_number"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CreateAssignmentInput struct {
	StreetID          uuid.UUID  `json:"-"`
	ResponsibleUserID uuid.UUID  `json:"user_id"`
	AssignedByUserID  *uuid.UUID `json:"-"`
	FromHouseNumber   string     `json:"from_house_number"`
	ToHouseNumber     string     `json:"to_house_number"`
}
