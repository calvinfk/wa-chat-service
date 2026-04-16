package errs

import "errors"

// Generic
var (
	ErrGenericForbidden        = errors.New("cannot access this resource")
	ErrGenericNotFound         = errors.New("data not found")
	ErrGenericAlreadyExists    = errors.New("data already exists")
	ErrGenericInvalidInput     = errors.New("invalid input")
	ErrGenericValidationError  = errors.New("validation error")
	ErrGenericInvalidQuery     = errors.New("invalid query parameters")
	ErrGenericInvalidBody      = errors.New("invalid request body")
	ErrGenericInvalidFileType  = errors.New("invalid file type")
	ErrGenericFileSizeExceeded = errors.New("file size exceeds the allowed limit")
	ErrGenericEmptyFile        = errors.New("file is empty")
	ErrGenericUnauthorized     = errors.New("unauthorized access")
)
