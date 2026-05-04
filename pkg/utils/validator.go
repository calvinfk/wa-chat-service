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

// NewValidator initializes a new instance of the validator with custom validation functions and tag name extraction for JSON fields.
// It registers custom validation tags such as "ext" for file extensions, "min_files" and "max_files" for validating the number of uploaded files, and "filter_options" for validating filter_request values.
// The tag name function ensures that validation error messages reference the JSON field names instead of the struct field names, improving clarity in error reporting.
func NewValidator() *validator.Validate {
	v := validator.New()
	// Register a custom tag name function to extract JSON field names from struct tags for better error messages.
	// This allows validation errors to reference the JSON field names instead of the Go struct field names, making error messages more user-friendly and easier to understand in the context of API requests and responses.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// The "ext" tag validates that a string field (e.g., a filename) has an allowed file extension.
	// It returns true for empty values so that optional fields do not fail validation unless a value is actually provided.
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
	// The "min_files" and "max_files" tags validate that a slice of *multipart.FileHeader has at least or at most a certain number of files, respectively.
	// These helpers are used by multipart upload DTOs to enforce file count limits before the request reaches business logic.
	v.RegisterValidation("min_files", func(fl validator.FieldLevel) bool {
		files, ok := fl.Field().Interface().([]*multipart.FileHeader)
		if !ok {
			return false
		}
		min, _ := strconv.Atoi(fl.Param())
		return len(files) >= min
	})
	// The "max_files" tag validates that a slice of *multipart.FileHeader does not exceed a certain number of files, ensuring that users do not upload more files than allowed by the application.
	v.RegisterValidation("max_files", func(fl validator.FieldLevel) bool {
		files, ok := fl.Field().Interface().([]*multipart.FileHeader)
		if !ok {
			return false
		}
		max, _ := strconv.Atoi(fl.Param())
		return len(files) <= max
	})
	// The "filter_options" tag validates that a string field is a valid filter_request value
	// It also supports a special "in:" prefix to allow for multiple comma-separated values, ensuring that all specified values are within the allowed options.
	// Example: "in:active,pending" means every value after the prefix must be one of the allowed options in the tag parameter.
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

// Custom StructValidator that implements the Validate method required by Fiber's Validator interface.
// This allows us to use our custom validation logic, including custom validation tags like "ext", "min_files", "max_files", and "filter_options".
type StructValidator struct {
	validate *validator.Validate
}

func NewStructValidator() *StructValidator {
	v := NewValidator()
	return &StructValidator{
		validate: v,
	}
}

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
	return FormatErrors(err)
}

func FormatErrors(err error) map[string]string {
	result := make(map[string]string)

	if ve, ok := err.(validator.ValidationErrors); ok {
		// Each validation error describes a field, the failed rule, and any rule parameter.
		// We convert that into a compact map keyed by the JSON field path for easier API responses.
		for _, fe := range ve {
			tag := fe.Tag()
			param := fe.Param()

			if param != "" {
				// Translate "FirstName" -> "first_name"
				// We pass fe.Value()'s parent context if possible,
				// but usually, your Document struct is the context.
				tag = tag + "=" + param
			}
			name := fe.Namespace()
			result[name] = tag
		}
	}
	return result
}

// ValidateEmail checks if the provided email string matches a regular expression pattern for valid email addresses. It returns true if the email is valid according to the regex pattern, and false otherwise. The regex pattern used is a common one for basic email validation, checking for the presence of characters before and after the "@" symbol and a valid domain format.
// This is a lightweight syntax check only; it does not verify that the mailbox exists or that the domain can receive mail.
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
