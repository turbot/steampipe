package utils

import "strings"

// UnquoteStringArray removes quote marks from elements of string array
func UnquoteStringArray(stringArray []string) []string {
	res := make([]string, len(stringArray))
	for i, s := range stringArray {
		res[i] = strings.Replace(s, `"`, ``, -1)
	}
	return res
}
