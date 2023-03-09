package utils

import (
	"sort"

	"golang.org/x/exp/maps"
)

// MergeMaps merges 'new' onto 'old'.
// Values existing in old already have precedence
// Any value existing in new but not old is added to old
func MergeMaps[M ~map[K]V, K comparable, V any](old, new M) M {
	if old == nil {
		return new
	}
	if new == nil {
		return old
	}
	res := maps.Clone(old)
	for k, v := range new {
		if _, ok := old[k]; !ok {
			res[k] = v
		}
	}

	return res
}

func SortedMapKeys[V any](m map[string]V) []string {
	var keys = maps.Keys(m)
	sort.Strings(keys)
	return keys
}
