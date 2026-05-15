package mfys

import (
	"time"

	"github.com/google/uuid"
)

type MFY struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	Region          *string    `json:"region"`
	District        *string    `json:"district"`
	TargetVotes     *int32     `json:"target_votes"`
	Season          *string    `json:"season"`
	IsActive        bool       `json:"is_active"`
	CreatedByUserID *uuid.UUID `json:"created_by_user_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateMFYInput struct {
	Name            string     `json:"name"`
	Region          *string    `json:"region"`
	District        *string    `json:"district"`
	TargetVotes     *int32     `json:"target_votes"`
	Season          *string    `json:"season"`
	CreatedByUserID *uuid.UUID `json:"-"`
}

type UpdateMFYInput struct {
	Name        string  `json:"name"`
	Region      *string `json:"region"`
	District    *string `json:"district"`
	TargetVotes *int32  `json:"target_votes"`
	Season      *string `json:"season"`
	IsActive    bool    `json:"is_active"`
}

type AssignChairmanInput struct {
	UserID uuid.UUID `json:"user_id"`
}
