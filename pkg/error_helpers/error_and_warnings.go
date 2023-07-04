package error_helpers

// type ErrorAndWarnings struct {
// 	Error    error
// 	Warnings []string
// }

// func DiagsToErrorsAndWarnings(errPrefix string, diags hcl.Diagnostics) *ErrorAndWarnings {
// 	return NewErrorsAndWarning(
// 		plugin.DiagsToError(errPrefix, diags),
// 		plugin.DiagsToWarnings(diags)...,
// 	)
// }

// func NewErrorsAndWarning(err error, warnings ...string) *ErrorAndWarnings {
// 	return &ErrorAndWarnings{
// 		Error: err, Warnings: warnings,
// 	}
// }

// func (r *ErrorAndWarnings) WrapErrorWithMessage(msg string) *ErrorAndWarnings {
// 	if r.Error != nil {
// 		r.Error = sperr.WrapWithMessage(r.Error, msg)
// 	}
// 	return r
// }

// func (r *ErrorAndWarnings) AddWarning(warnings ...string) {
// 	r.Warnings = append(r.Warnings, warnings...)
// }

// func (r *ErrorAndWarnings) ShowWarnings() {
// 	for _, w := range r.Warnings {
// 		error_helpers.ShowWarning(w)
// 	}
// }

// func (r *ErrorAndWarnings) GetError() error {
// 	if r == nil {
// 	}
// 	return r.Error
// }

// func (r *ErrorAndWarnings) Merge(other *ErrorAndWarnings) {
// 		return
// 	}
// 	if r.Error == nil {
// 		r.Error = other.Error
// 	}
// 	if len(other.Warnings) > 0 {
// 		r.AddWarning(other.Warnings...)
// 	}
// }
