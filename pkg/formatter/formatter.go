package formatter

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// CapitalizeFirstLetter capitalizes the first letter of a given string. If the input string is empty, it returns the empty string. Otherwise, it converts the first character to uppercase and concatenates it with the rest of the string unchanged.
func CapitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
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

func MapToJSONString(m map[string]any) string {
	b, _ := json.Marshal(m)
	return string(b)
}

type structValidator struct {
	validate *validator.Validate
}

func Validator() *structValidator {
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

func (v *structValidator) GetErrorMessages(err error) map[string]string {
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

func GetURLHeaders(fileURL string) (http.Header, error) {
	// Implementation for checking MIME type based on file URL
	log.Println("[INFO][formatter/formatter.go][GetURLHeaders] Checking URL headers for:", fileURL)
	client := http.Client{}
	resp, err := client.Head(fileURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Header, nil
}

func DownloadFile(fileURL string) ([]byte, http.Header, error) {
	client := http.Client{}
	resp, err := client.Get(fileURL)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	fileData := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(fileData)
	if err != nil {
		return nil, nil, err
	}
	return fileData, resp.Header, nil
}

func IsGCSSignedURLExpired(signedURL string) (bool, error) {
	u, err := url.Parse(signedURL)
	if err != nil {
		return false, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()

	// Try V4 first
	if date := q.Get("X-Goog-Date"); date != "" {
		if expiresIn := q.Get("X-Goog-Expires"); expiresIn != "" {
			startTime, err := time.Parse("20060102T150405Z", date)
			if err != nil {
				return false, err
			}
			seconds, err := strconv.ParseInt(expiresIn, 10, 64)
			if err != nil {
				return false, err
			}
			return time.Now().After(startTime.Add(time.Duration(seconds) * time.Second)), nil
		}
	}

	// Fall back to V2
	if expires := q.Get("Expires"); expires != "" {
		unixTime, err := strconv.ParseInt(expires, 10, 64)
		if err != nil {
			return false, err
		}
		return time.Now().After(time.Unix(unixTime, 0)), nil
	}

	return false, fmt.Errorf("unrecognized or missing expiration parameters")
}

func GetFileNameFromURL(fileURL string) string {
	u, err := url.Parse(fileURL)
	if err != nil {
		return ""
	}
	urlParts := strings.Split(u.Path, "/")
	if len(urlParts) > 0 {
		tempName := urlParts[len(urlParts)-1]
		tempName = strings.Split(tempName, "?")[0] // remove query params
		// check if the last part of the URL path has a valid filename format (e.g., has an extension)
		if strings.Contains(tempName, ".") {
			return tempName
		}
	}
	return ""
}
