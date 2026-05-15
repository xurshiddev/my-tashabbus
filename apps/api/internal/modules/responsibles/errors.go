package responsibles

import "errors"

var (
	ErrForbidden               = errors.New("forbidden")
	ErrAssignmentNotFound      = errors.New("responsible assignment not found")
	ErrResponsibleRoleRequired = errors.New("target user must have RESPONSIBLE_PERSON role")
	ErrResponsibleUserInactive = errors.New("responsible user is inactive")
	ErrResponsibleWrongMFY     = errors.New("responsible user belongs to another MFY")
	ErrHouseRangeRequired      = errors.New("from_house_number and to_house_number are required")
	ErrInvalidHouseRange       = errors.New("from_house_number must be less than or equal to to_house_number")
	ErrInvalidPagination       = errors.New("invalid pagination")
	ErrInternalStoreNotReady   = errors.New("responsible assignment store is not configured")
)
