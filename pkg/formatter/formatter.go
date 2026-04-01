package formatter

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// AnyToPointer is a generic function that takes a value of any type and returns a pointer to that value. This can be useful in situations where you need to pass a pointer to a value, such as when working with optional fields in structs.
//
//go:fix inline
func AnyToPointer[T any](value T) *T {
	return new(value)
}

// HasCommonValueMap checks if there is at least one common value between two slices of any comparable type. It uses a map to store the values of the first slice for efficient lookups, and then iterates through the second slice to check for common values. If a common value is found, it returns true; otherwise, it returns false after checking all values.
func HasCommonValueMap[T comparable](arr1 []T, arr2 []T) bool {
	if len(arr1) == 0 || len(arr2) == 0 {
		return false
	}
	valueMap := make(map[T]bool, len(arr1))
	for _, value := range arr1 {
		valueMap[value] = true
	}
	for _, value := range arr2 {
		if valueMap[value] {
			return true
		}
	}
	return false
}

// NullToEmptyString converts a pointer to a string to an empty string if the pointer is nil. If the pointer is not nil, it returns the value pointed to by the pointer. This function is useful for handling optional string fields in structs, allowing you to easily convert nil pointers to empty strings when needed.
func NullToEmptyString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// CapitalizeFirstLetter capitalizes the first letter of a given string. If the input string is empty, it returns the empty string. Otherwise, it converts the first character to uppercase and concatenates it with the rest of the string unchanged.
func CapitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
}

// PrintJson takes any data structure, marshals it into a pretty-printed JSON format, and logs the resulting JSON string. If there is an error during the marshaling process, it returns the error instead of logging.
func PrintJson(data any) error {
	if data == nil {
		log.Println("nil")
		return nil
	}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	log.Println(string(jsonData))
	return nil
}

// RandString generates a random string of the specified length using a combination of uppercase and lowercase letters. It uses a seeded random number generator to ensure that the generated string is different each time the function is called.
func RandString(length int) string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	charset := "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// HashString takes a string input and returns its SHA-256 hash as a hexadecimal string. It uses the crypto/sha256 package to compute the hash and fmt.Sprintf to format the output as a hexadecimal string.
func HashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", hash)
}

// ValidateEmail checks if the provided email string matches a regular expression pattern for valid email addresses. It returns true if the email is valid according to the regex pattern, and false otherwise. The regex pattern used is a common one for basic email validation, checking for the presence of characters before and after the "@" symbol and a valid domain format.
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// StructToMap converts a struct to a map[string]any by first marshaling the struct to JSON and then unmarshaling it back into a map.
func StructToMap(v any, omitNull bool) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	err = json.Unmarshal(b, &result)
	// if the value is null
	if omitNull {
		omitNullValues(result)
	}
	return result, err
}

func omitNullValues(m map[string]any) map[string]any {
	// recursively omit null values from the map
	for k, v := range m {
		switch val := v.(type) {
		case map[string]any:
			m[k] = omitNullValues(val)
			if len(m[k].(map[string]any)) == 0 {
				delete(m, k)
			}
		case []any:
			for i, item := range val {
				if itemMap, ok := item.(map[string]any); ok {
					val[i] = omitNullValues(itemMap)
					if len(val[i].(map[string]any)) == 0 {
						val = append(val[:i], val[i+1:]...)
					}
				} else if item == nil {
					val = append(val[:i], val[i+1:]...)
				}
			}
		case nil:
			delete(m, k)
			continue
		}
	}
	return m
}

func MapToStruct(m map[string]any, v any) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func StructToJsonString(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func AnyToJsonString(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
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

func getJsonName(entity any, fieldName string) string {
	t := reflect.TypeOf(entity)

	// If it's a pointer, get the element
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Handle the case where we are validating the wrapper struct
	field, found := t.FieldByName(fieldName)
	if !found {
		return strings.ToLower(fieldName) // Fallback
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return strings.ToLower(fieldName)
	}

	// Return only the name part of the tag (before any commas)
	return strings.Split(jsonTag, ",")[0]
}
