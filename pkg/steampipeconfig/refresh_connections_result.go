package steampipeconfig

import (
	"github.com/turbot/steampipe/pkg/error_helpers"
)

// RefreshConnectionResult is a structure used to contain the result of either a RefreshConnections or a NewLocalClient operation
type RefreshConnectionResult struct {
	UpdatedConnections bool
	Warnings           []string
	Error              error
	Updates            *ConnectionUpdates
}

func (r *RefreshConnectionResult) AddWarning(warning string) {
	r.Warnings = append(r.Warnings, warning)
}

func (r *RefreshConnectionResult) ShowWarnings() {
	for _, w := range r.Warnings {
		error_helpers.ShowWarning(w)
	}
}

func (r *RefreshConnectionResult) Merge(other *RefreshConnectionResult) {
	if other == nil {
		return
	}
	if other.UpdatedConnections {
		r.UpdatedConnections = other.UpdatedConnections
	}
	if other.Error != nil {
		r.Error = other.Error
	}
	r.Warnings = append(r.Warnings, other.Warnings...)
}
