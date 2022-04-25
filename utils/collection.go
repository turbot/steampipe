package utils

func Partition[V any](elements []V, predicate func(V) bool) ([]V, []V) {
	leftPartition := []V{}
	rightPartition := []V{}

	for _, v := range elements {
		if predicate(v) {
			leftPartition = append(leftPartition, v)
		} else {
			rightPartition = append(rightPartition, v)
		}
	}

	return leftPartition, rightPartition
}
