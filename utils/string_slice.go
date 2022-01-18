package utils

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
