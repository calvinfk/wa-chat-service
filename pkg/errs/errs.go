package errs

import "errors"

// Auth
var (
	ErrAuthEmailExists         = errors.New("email already exists")
	ErrAuthInvalidCredentials  = errors.New("invalid credentials")
	ErrAuthInvalidRefreshToken = errors.New("invalid refresh token")
	ErrAuthExpiredRefreshToken = errors.New("expired refresh token")
	ErrAuthExpiredAccessToken  = errors.New("token is expired")
	ErrAuthInvalidAccessToken  = errors.New("invalid access token")
	ErrAuthMissingRefreshToken = errors.New("refresh token is missing")
)

// User
var (
	ErrUserRoleUserNotFound = errors.New("user role not found")
	ErrUserNotFound         = errors.New("user not found")
)

// Role
var (
	ErrRoleNotFound = errors.New("role not found")
)

// Report
var (
	ErrReportStatusInvalid = errors.New("report status invalid")
	ErrReportStatusOpen    = errors.New("report status is open, cannot perform this action")
	ErrReportNotInactive   = errors.New("report is active, cannot perform this action")
)
