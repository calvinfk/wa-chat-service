package errs

import "errors"

// Generic
var (
	ErrGenericTxNilDB            = errors.New("transaction requires a non-nil database connection")
	ErrGenericInvalidTransaction = errors.New("invalid transaction type provided")
	ErrGenericForbidden          = errors.New("user cannot access this resource")
	ErrGenericNotFound           = errors.New("requested data not found")
	ErrGenericAlreadyExists      = errors.New("data already exists")
	ErrGenericInvalidInput       = errors.New("invalid input")
	ErrGenericValidationError    = errors.New("validation error")
	ErrGenericInvalidQuery       = errors.New("invalid query parameters")
	ErrGenericInvalidBody        = errors.New("invalid request body")
	ErrGenericInvalidFileType    = errors.New("invalid file type")
	ErrGenericFileSizeExceeded   = errors.New("file size exceeds the allowed limit")
	ErrGenericEmptyFile          = errors.New("file is empty")
)
