package households

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusNew            Status = "NEW"
	StatusVisited        Status = "VISITED"
	StatusExplained      Status = "EXPLAINED"
	StatusPartiallyVoted Status = "PARTIALLY_VOTED"
	StatusFullyVoted     Status = "FULLY_VOTED"
	StatusNotHome        Status = "NOT_HOME"
	StatusCallbackNeeded Status = "CALLBACK_NEEDED"
	StatusRefused        Status = "REFUSED"
	StatusInvalidInfo    Status = "INVALID_INFO"
)

type Household struct {
	ID                        uuid.UUID  `json:"id"`
	MFYID                     uuid.UUID  `json:"mfy_id"`
	StreetID                  uuid.UUID  `json:"street_id"`
	HouseNumber               string     `json:"house_number"`
	TotalNumbers              int32      `json:"total_numbers"`
	ContactedNumbers          int32      `json:"contacted_numbers"`
	VotedNumbers              int32      `json:"voted_numbers"`
	Status                    Status     `json:"status"`
	Notes                     *string    `json:"notes"`
	AssignedResponsibleUserID *uuid.UUID `json:"assigned_responsible_user_id"`
	CreatedByUserID           *uuid.UUID `json:"created_by_user_id"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

type ChangeLog struct {
	ID              uuid.UUID  `json:"id"`
	HouseholdID     uuid.UUID  `json:"household_id"`
	ChangedByUserID *uuid.UUID `json:"changed_by_user_id"`
	FieldName       string     `json:"field_name"`
	OldValue        *string    `json:"old_value"`
	NewValue        *string    `json:"new_value"`
	Note            *string    `json:"note"`
	CreatedAt       time.Time  `json:"created_at"`
}

type CreateHouseholdInput struct {
	MFYID                     uuid.UUID  `json:"-"`
	StreetID                  uuid.UUID  `json:"-"`
	HouseNumber               string     `json:"house_number"`
	TotalNumbers              int32      `json:"total_numbers"`
	ContactedNumbers          int32      `json:"contacted_numbers"`
	VotedNumbers              int32      `json:"voted_numbers"`
	Status                    Status     `json:"status"`
	Notes                     *string    `json:"notes"`
	AssignedResponsibleUserID *uuid.UUID `json:"assigned_responsible_user_id"`
	CreatedByUserID           *uuid.UUID `json:"-"`
}

type UpdateHouseholdInput struct {
	HouseNumber      string  `json:"house_number"`
	TotalNumbers     int32   `json:"total_numbers"`
	ContactedNumbers int32   `json:"contacted_numbers"`
	VotedNumbers     int32   `json:"voted_numbers"`
	Status           Status  `json:"status"`
	Notes            *string `json:"notes"`
}

type CreateChangeLogInput struct {
	HouseholdID     uuid.UUID
	ChangedByUserID *uuid.UUID
	FieldName       string
	OldValue        *string
	NewValue        *string
	Note            *string
}
