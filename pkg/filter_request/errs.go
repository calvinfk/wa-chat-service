package filter_request

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidCountAlias = errors.New("firestore: couldn't get alias for COUNT from results")
	ErrNilPointer        = errors.New("nil pointer error")
)

var (
	ErrInvalidOperator = fmt.Errorf("invalid operator")
)
