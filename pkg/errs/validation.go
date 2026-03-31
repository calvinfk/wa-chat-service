package errs

var (
	ErrValidateEmptyField       = "cannot be empty"
	ErrValidateInvalidUUID      = "must be a valid UUID"
	ErrValidatePasswordMismatch = "password and confirm password do not match"
	ErrValidateInvalidEmail     = "must be a valid email"
	ErrValidatePasswordTooShort = "password must be at least 8 characters long"
	ErrValidateInvalidStatus    = "status must be either 1 (open), 2 (in progress), or 3 (closed)"
	ErrValidateInvalidIsActive  = "must be either 0 or 1"
)
