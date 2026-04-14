package utils

import (
	"mime/multipart"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

type StructValidator struct {
	validate *validator.Validate
}

func NewValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	v.RegisterValidation("ext", func(fl validator.FieldLevel) bool {
		if fl.Field().Kind() != reflect.String {
			return false
		}
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
	v.RegisterValidation("min_files", func(fl validator.FieldLevel) bool {
		files, ok := fl.Field().Interface().([]*multipart.FileHeader)
		if !ok {
			return false
		}
		min, _ := strconv.Atoi(fl.Param())
		return len(files) >= min
	})
	v.RegisterValidation("max_files", func(fl validator.FieldLevel) bool {
		files, ok := fl.Field().Interface().([]*multipart.FileHeader)
		if !ok {
			return false
		}
		max, _ := strconv.Atoi(fl.Param())
		return len(files) <= max
	})
	v.RegisterValidation("filter_options", func(fl validator.FieldLevel) bool {
		if fl.Field().Kind() != reflect.String {
			return false
		}
		value := fl.Field().String()
		valueParts := strings.SplitN(value, ":", 2)
		if len(valueParts) == 2 {
			value = valueParts[1]
		}
		param := fl.Param()
		allowedOptions := strings.Split(param, " ")
		if valueParts[0] == "in" {
			for option := range strings.SplitSeq(valueParts[1], ",") {
				if !slices.Contains(allowedOptions, option) {
					return false
				}
			}
			return true
		} else {
			return slices.Contains(allowedOptions, value)
		}
	})
	return v
}

func NewStructValidator() *StructValidator {
	return &StructValidator{
		validate: NewValidator(),
	}
}

// Validator needs to implement the Validate method
func (v *StructValidator) Validate(out any) error {
	if out == nil {
		return nil // Or return a specific "missing body" error
	}
	return v.validate.Struct(out)
}

func GetValidatorErrorMessages(err error) map[string]string {
	if err == nil {
		return nil
	}
	if _, ok := err.(validator.ValidationErrors); !ok {
		return map[string]string{
			"error": err.Error(),
		}
	}
	return FormatErrors(err, nil)
}

func FormatErrors(err error, rootEntity any) map[string]string {
	result := make(map[string]string)

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			tag := fe.Tag()
			param := fe.Param()

			if param != "" {
				// Translate "FirstName" -> "first_name"
				// We pass fe.Value()'s parent context if possible,
				// but usually, your Document struct is the context.
				jsonParam := getJsonName(fe.Type(), param)
				tag = tag + "=" + jsonParam
			}

			result[fe.Namespace()] = tag
		}
	}
	return result
}

// ValidateEmail checks if the provided email string matches a regular expression pattern for valid email addresses. It returns true if the email is valid according to the regex pattern, and false otherwise. The regex pattern used is a common one for basic email validation, checking for the presence of characters before and after the "@" symbol and a valid domain format.
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
