package utils

import (
	"encoding/json"
	"fmt"
	"unicode"
)

// CapitalizeFirstLetter capitalizes the first letter of a given string. If the input string is empty, it returns the empty string. Otherwise, it converts the first character to uppercase and concatenates it with the rest of the string unchanged.
func CapitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
}

func AnyToBytes(data any) ([]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("data is nil")
	}
	if _, ok := data.(string); ok {
		return []byte(data.(string)), nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data to bytes: %v", err)
	}
	return bytes, nil
}

func AnySliceToStringSlice(data any) ([]string, error) {
	if data == nil {
		return nil, fmt.Errorf("data is nil")
	}
	if slice, ok := data.([]string); ok {
		return slice, nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data to bytes: %v", err)
	}
	var stringSlice []string
	err = json.Unmarshal(bytes, &stringSlice)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal bytes to []string: %v", err)
	}
	return stringSlice, nil
}
