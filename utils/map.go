package utils

// MergeStringMaps merges 'new' onto old. Any vakue existing in new but not old is added to old
// NOTE this mutates old
func MergeStringMaps(old, new map[string]string) map[string]string {
	if old == nil {
		return new
	}
	if new == nil {
		return old
	}
	for k, v := range new {
		if _, ok := old[k]; ok {
			old[k] = v
		}
	}

	return old
}
