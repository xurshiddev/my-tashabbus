package users

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleSuperAdmin        Role = "SUPER_ADMIN"
	RoleMFYChairman       Role = "MFY_CHAIRMAN"
	RoleStreetLeader      Role = "STREET_LEADER"
	RoleResponsiblePerson Role = "RESPONSIBLE_PERSON"
)

type User struct {
	ID               uuid.UUID  `json:"id"`
	FullName         string     `json:"full_name"`
	Phone            *string    `json:"phone"`
	TelegramID       *int64     `json:"telegram_id"`
	TelegramUsername *string    `json:"telegram_username"`
	Role             Role       `json:"role"`
	MFYID            *uuid.UUID `json:"mfy_id"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CreateUserInput struct {
	FullName         string     `json:"full_name"`
	Phone            *string    `json:"phone"`
	TelegramID       *int64     `json:"telegram_id"`
	TelegramUsername *string    `json:"telegram_username"`
	Role             Role       `json:"role"`
	MFYID            *uuid.UUID `json:"mfy_id"`
}

type UpdateUserInput struct {
	FullName string     `json:"full_name"`
	Phone    *string    `json:"phone"`
	Role     Role       `json:"role"`
	MFYID    *uuid.UUID `json:"mfy_id"`
	IsActive bool       `json:"is_active"`
}

type SetTelegramIdentityInput struct {
	TelegramID       int64   `json:"telegram_id"`
	TelegramUsername *string `json:"telegram_username"`
}

func IsValidRole(role Role) bool {
	switch role {
	case RoleSuperAdmin, RoleMFYChairman, RoleStreetLeader, RoleResponsiblePerson:
		return true
	default:
		return false
	}
}
