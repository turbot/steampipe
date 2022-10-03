package utils

import (
	"golang.org/x/exp/maps"
	"sort"
)

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
		if _, ok := old[k]; ok {
			res[k] = v
		}
	}

	return res
}
func SortedStringKeys[V any](m map[string]V) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func StringMapsEqual(l, r map[string]string) bool {
	// treat nil as empty
	if l == nil {
		l = map[string]string{}
	}
	if r == nil {
		r = map[string]string{}
	}

	if len(l) != len(r) {
		return false
	}

	for k, lVal := range l {
		rVal, ok := r[k]
		if !ok || rVal != lVal {
			return false
		}
	}
	return true
}

func MapValues[V any](m map[string]V) []V {
	res := make([]V, len(m))
	idx := 0
	for _, v := range m {
		res[idx] = v
		idx++
	}
	return res
}
