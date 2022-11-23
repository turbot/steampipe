package utils

// Partition splits the array of elements into two groups:
// the left partition contains elements that the predicate returns `true` for.
// the right partition contains elements that the predicate returns `false` for.
//
// The predicate is invoked with each element
func Partition[V any](elements []V, predicate func(V) bool) ([]V, []V) {
	leftPartition := make([]V, 0, len(elements))
	rightPartition := make([]V, 0, len(elements))
	for _, v := range elements {
		if predicate(v) {
			leftPartition = append(leftPartition, v)
		} else {
			rightPartition = append(rightPartition, v)
		}
	}
	return leftPartition, rightPartition
}

// Filter returns a new slice only containing the elements for which invoking the predicate
// returned true
//
// The predicate is invoked with each element
func Filter[V any](elements []V, predicate func(V) bool) []V {
	filtered := make([]V, 0, len(elements))
	for _, v := range elements {
		if predicate(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// Map returns a new slice where every value is mapped to a new value through the mapper
//
// The mapper is invoked with each element
func Map[V any, M any](elements []V, mapper func(V) M) []M {
	mapped := make([]M, 0, len(elements))
	for _, v := range elements {
		mapped = append(mapped, mapper(v))
	}
	return mapped
}
