package households

import "errors"

var (
	ErrForbidden             = errors.New("forbidden")
	ErrHouseholdNotFound     = errors.New("household not found")
	ErrChangeLogNotFound     = errors.New("household change log not found")
	ErrHouseNumberRequired   = errors.New("house number is required")
	ErrInvalidCounts         = errors.New("household counts must be non-negative and not exceed total numbers")
	ErrInvalidStatus         = errors.New("invalid household status")
	ErrInvalidPagination     = errors.New("invalid pagination")
	ErrDuplicateHousehold    = errors.New("household already exists in this street")
	ErrStreetRequired        = errors.New("street is required")
	ErrInternalStoreNotReady = errors.New("household store is not configured")
)
