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

// Database
var (
	ErrDBTxNil     = errors.New("transaction requires a non-nil database connection")
	ErrDBTxInvalid = errors.New("invalid transaction type provided")
)

func IsUnauthenticatedError(err error) bool {
	unauthenticatedErrors := map[error]bool{
		ErrAuthInvalidCredentials:  true,
		ErrAuthInvalidAccessToken:  true,
		ErrAuthExpiredAccessToken:  true,
		ErrAuthMissingRefreshToken: true,
	}
	return unauthenticatedErrors[err]
}
