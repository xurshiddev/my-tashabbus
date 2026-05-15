package mfys

import "errors"

var (
	ErrNameRequired         = errors.New("mfy name is required")
	ErrTargetVotesNegative  = errors.New("target votes cannot be negative")
	ErrMFYNotFound          = errors.New("mfy not found")
	ErrForbidden            = errors.New("forbidden")
	ErrInvalidPagination    = errors.New("invalid pagination")
	ErrChairmanRoleRequired = errors.New("target user must be MFY_CHAIRMAN")
	ErrStoreNotConfigured   = errors.New("mfy store is not configured")
)
