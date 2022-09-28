package utils

import "strings"

// TODO: investigate turbot/go-kit/helpers
func StringSliceDistinct(slice []string) []string {
	var res []string
	occurenceMap := make(map[string]struct{})
	for _, item := range slice {
		occurenceMap[item] = struct{}{}
	}
	for item := range occurenceMap {
		res = append(res, item)
	}
	return res
}

// UnquoteStringArray removes quote marks from elements of string array
func UnquoteStringArray(stringArray []string) []string {
	res := make([]string, len(stringArray))
	for i, s := range stringArray {
		res[i] = strings.Replace(s, `"`, ``, -1)
	}
	return res
}
