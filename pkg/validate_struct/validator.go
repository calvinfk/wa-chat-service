package validate_struct

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type structValidator struct {
	validate *validator.Validate
}

func New() *structValidator {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return &structValidator{
		validate: v,
	}
}

// Validator needs to implement the Validate method
func (v *structValidator) Validate(out any) error {
	if out == nil {
		return nil // Or return a specific "missing body" error
	}
	return v.validate.Struct(out)
}
