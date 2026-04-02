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
	v.RegisterValidation("ext", func(fl validator.FieldLevel) bool {
		field := fl.Field().String()
		if field == "" {
			return true
		}

		// Get the parameters from the tag (e.g., "jpg png jpeg")
		param := fl.Param()
		allowedExts := strings.Split(param, " ")

		loweredField := strings.ToLower(field)
		for _, ext := range allowedExts {
			// Check if filename ends with .ext
			if strings.HasSuffix(loweredField, "."+strings.ToLower(ext)) {
				return true
			}
		}
		return false
	})
	// TODO: Add validator if link is expired or not valid anymore (e.g., for media links)
	// TODO: check if from google storage, check the extension is allowed
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
