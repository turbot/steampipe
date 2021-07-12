package db

// RefreshConnectionResult is a structure used to contain the result of either a RefreshConnections or a NewClient operation
type RefreshConnectionResult struct {
	UpdatedConnections bool
	Warning            string
	Error              error
}
