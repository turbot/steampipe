package db_common

import "fmt"

// RefreshConnectionResult is a structure used to contain the result of either a RefreshConnections or a NewLocalClient operation
type RefreshConnectionResult struct {
	UpdatedConnections bool
	Warnings           []string
	Error              error
}

func (r *RefreshConnectionResult) ShowWarnings() {
	for _, w := range r.Warnings {
		fmt.Println(w)
	}
}
