package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type TokenResponse struct {
	AccessToken string     `json:"access_token"`
	TokenType   string     `json:"token_type"`
	ExpiresIn   int64      `json:"expires_in"`
	User        users.User `json:"user"`
}

type TokenClaims struct {
	UserID     uuid.UUID
	Role       users.Role
	TelegramID *int64
	IssuedAt   time.Time
	ExpiresAt  time.Time
}

type DevLoginRequest struct {
	FullName         string     `json:"full_name"`
	Role             users.Role `json:"role"`
	TelegramID       *int64     `json:"telegram_id"`
	TelegramUsername *string    `json:"telegram_username"`
}

type TelegramAuthRequest struct {
	InitData      string `json:"init_data"`
	DevTelegramID *int64 `json:"dev_telegram_id"`
	DevFullName   string `json:"dev_full_name"`
	DevUsername   string `json:"dev_username"`
}

type TelegramUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type DeploymentMFY struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}
