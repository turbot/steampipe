package error_helpers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/tfdiags"
	"github.com/turbot/go-kit/helpers"
)

// DiagsToError converts tfdiags diags into an error
func DiagsToError(prefix string, diags tfdiags.Diagnostics) error {
	// convert the first diag into an error
	if !diags.HasErrors() {
		return nil
	}
	errorStrings := []string{fmt.Sprintf("%s", prefix)}
	// store list of messages (without the range) and use for deduping (we may get the same message for multiple ranges)
	errorMessages := []string{}
	for _, diag := range diags {
		if diag.Severity() == tfdiags.Error {
			errorString := fmt.Sprintf("%s", diag.Description().Summary)
			if diag.Description().Detail != "" {
				errorString += fmt.Sprintf(": %s", diag.Description().Detail)
			}

			if !helpers.StringSliceContains(errorMessages, errorString) {
				errorMessages = append(errorMessages, errorString)
				// now add in the subject and add to the output array
				if diag.Source().Subject != nil && len(diag.Source().Subject.Filename) > 0 {
					errorString += fmt.Sprintf("\n(%s)", diag.Source().Subject.StartString())
				}
				errorStrings = append(errorStrings, errorString)

			}
		}
	}
	if len(errorStrings) > 0 {
		errorString := strings.Join(errorStrings, "\n")
		if len(errorStrings) > 1 {
			errorString += "\n"
		}
		return errors.New(errorString)
	}
	return diags.Err()
}
