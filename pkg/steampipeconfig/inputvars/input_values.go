package inputvars

import (
	"github.com/turbot/terraform-components/terraform"

	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// SetVariableValues determines whether the given variable is a public variable and if so sets its value
func SetVariableValues(vv terraform.InputValues, m *modconfig.ModVariableMap) {
	for name, inputValue := range vv {
		variable, ok := m.PublicVariables[name]
		// if this variable does not exist in public variables, skip
		if !ok {
			// we should have already caught this
			continue
		}
		variable.SetInputValue(
			inputValue.Value,
			inputValue.SourceTypeString(),
			inputValue.SourceRange)
	}
}
