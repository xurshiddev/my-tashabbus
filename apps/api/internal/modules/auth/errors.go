package auth

import "errors"

var (
	ErrInvalidToken        = errors.New("invalid token")
	ErrExpiredToken        = errors.New("token is expired")
	ErrWeakJWTSecret       = errors.New("jwt secret is not safe for production")
	ErrDevLoginDisabled    = errors.New("dev login is disabled in production")
	ErrTelegramTokenNeeded = errors.New("telegram bot token is required")
	ErrInvalidInitData     = errors.New("invalid telegram init data")
	ErrOldInitData         = errors.New("telegram init data is too old")
	ErrUserNotRegistered   = errors.New("telegram user is not registered")
)
