package streets

import "errors"

var (
	ErrNameRequired             = errors.New("street name is required")
	ErrPlannedHouseholdsCount   = errors.New("planned households count cannot be negative")
	ErrStreetNotFound           = errors.New("street not found")
	ErrAssignmentNotFound       = errors.New("street leader assignment not found")
	ErrDuplicateStreetName      = errors.New("active street name already exists in this mfy")
	ErrForbidden                = errors.New("forbidden")
	ErrInvalidPagination        = errors.New("invalid pagination")
	ErrStreetLeaderRoleRequired = errors.New("target user must be STREET_LEADER")
	ErrStreetLeaderWrongMFY     = errors.New("street leader belongs to another mfy")
	ErrStoreNotConfigured       = errors.New("street store is not configured")
)
