package users

import "errors"

var (
	ErrInvalidRole        = errors.New("invalid role")
	ErrFullNameRequired   = errors.New("full name is required")
	ErrUserNotFound       = errors.New("user not found")
	ErrTelegramIDConflict = errors.New("telegram id already exists")
	ErrTelegramIDRequired = errors.New("telegram id is required")
	ErrInvalidPagination  = errors.New("invalid pagination")
	ErrUserInactive       = errors.New("user is inactive")
	ErrStoreNotConfigured = errors.New("user store is not configured")
)
