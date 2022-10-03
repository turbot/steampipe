package utils

import (
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

// TODO KAI uypdate to use maps package
func SortedMapKeys[V any](m map[string]V) []string {
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

// MapValues returns an array containing the values of map v
func MapValues[V any](m map[string]V) []V {
	res := make([]V, len(m))
	idx := 0
	for _, v := range m {
		res[idx] = v
		idx++
	}
	return res
}

// SortedMapValues returns an array containing the values of map v
// The values are returned in the order of the sorted keys
func SortedMapValues[V any](m map[string]V) []V {
	keys := SortedMapKeys(m)
	res := make([]V, len(m))
	for i, k := range keys {
		res[i] = m[k]
	}
	return res
}
