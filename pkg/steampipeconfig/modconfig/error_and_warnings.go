package modconfig

import "github.com/turbot/steampipe/pkg/error_helpers"

type ErrorAndWarnings struct {
	Error    error
	Warnings []string
}

func NewErrorsAndWarning(err error, warnings ...string) *ErrorAndWarnings {
	return &ErrorAndWarnings{
		Error: err, Warnings: warnings,
	}
}

func (r *ErrorAndWarnings) AddWarning(warning string) {
	r.Warnings = append(r.Warnings, warning)
}

func (r *ErrorAndWarnings) ShowWarnings() {
	for _, w := range r.Warnings {
		error_helpers.ShowWarning(w)
	}
}

func (r *ErrorAndWarnings) GetError() error {
	if r == nil {
		return nil
	}
	return r.Error
}
