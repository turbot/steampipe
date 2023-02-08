package utils

import "unicode"

// ContainsUpper returns true if the string contains any uppercase characters
func ContainsUpper(s string) bool {
	hasUpper := false
	for _, r := range s {
		if unicode.IsUpper(r) {
			hasUpper = true
			break
		}
	}
	return hasUpper
}
