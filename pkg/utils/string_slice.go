package utils

import (
	"strings"
	"unicode"
)

// UnquoteStringArray removes quote marks from elements of string array
func UnquoteStringArray(stringArray []string) []string {
	res := make([]string, len(stringArray))
	for i, s := range stringArray {
		res[i] = strings.Replace(s, `"`, ``, -1)
	}
	return res
}

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
