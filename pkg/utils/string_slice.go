package utils

import (
	"strings"
)

// UnquoteStringArray removes quote marks from elements of string array
func UnquoteStringArray(stringArray []string) []string {
	res := make([]string, len(stringArray))
	for i, s := range stringArray {
		res[i] = strings.Replace(s, `"`, ``, -1)
	}
	return res
}

// StringSlicesEqual returns whether the 2 string slices are identical
func StringSlicesEqual(l, r []string) bool {
	return strings.Join(l, ",") == strings.Join(r, ",")
}
