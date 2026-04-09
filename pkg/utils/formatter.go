package utils

import (
	"unicode"
)

// CapitalizeFirstLetter capitalizes the first letter of a given string. If the input string is empty, it returns the empty string. Otherwise, it converts the first character to uppercase and concatenates it with the rest of the string unchanged.
func CapitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
}
