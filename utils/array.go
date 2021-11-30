package utils

import "strings"

func BatchStringArray(items []string, batchSize int) []string {
	// batch up into 1000 queries at a time
	var batched []string
	remaining := len(items)
	idx := 0
	for remaining > 0 {
		if remaining < batchSize {
			batchSize = remaining
		}
		b := strings.Join(items[idx:idx+batchSize], "\n")
		batched = append(batched, b)
		idx += batchSize
		remaining -= batchSize
	}
	return batched
}
