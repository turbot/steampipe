package steampipeconfig

import (
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/v2/error_helpers"
	"github.com/turbot/pipe-fittings/v2/utils"
)

// RefreshConnectionResult is a structure used to contain the result of either a RefreshConnections or a NewLocalClient operation
type RefreshConnectionResult struct {
	error_helpers.ErrorAndWarnings
	UpdatedConnections bool
	FailedConnections  map[string]string
}

func NewErrorRefreshConnectionResult(err error) *RefreshConnectionResult {
	return &RefreshConnectionResult{ErrorAndWarnings: error_helpers.NewErrorsAndWarning(err)}
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
	for c, err := range other.FailedConnections {
		if _, ok := r.FailedConnections[c]; !ok {
			r.AddFailedConnection(c, err)
		}
	}
}

func (r *RefreshConnectionResult) String() string {
	var op strings.Builder
	if len(r.Warnings) > 0 {
		op.WriteString(fmt.Sprintf("%s:\n\t%s\n", utils.Pluralize("Warning", len(r.Warnings)), strings.Join(r.Warnings, "\n\t")))
	}
	if r.Error != nil {
		op.WriteString(fmt.Sprintf("%s\n", r.Error.Error()))
	}
	op.WriteString(fmt.Sprintf("UpdatedConnections: %v\n", r.UpdatedConnections))
	return op.String()
}

func (r *RefreshConnectionResult) AddFailedConnection(c string, failure string) {
	if r.FailedConnections == nil {
		r.FailedConnections = make(map[string]string)
	}

	r.FailedConnections[c] = failure
}
