package utils

import (
	"strings"
)

// UnquoteStringArray removes quote marks from elements of string array
func UnquoteStringArray(stringArray []string) []string {
	res := make([]string, len(stringArray))
	for i, s := range stringArray {
		res[i] = strings.ReplaceAll(s, `"`, ``)
	}
	return res
}

// RemoveElementFromSlice takes a slice of strings and an index to remove,
// and returns a new slice with the specified element removed.
func RemoveElementFromSlice(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}
