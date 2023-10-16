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
			SourceTypeString(inputValue),
			inputValue.SourceRange)
	}
}

func SourceTypeString(v *terraform.InputValue) string {
	switch v.SourceType {
	case terraform.ValueFromConfig:
		return "config"
	case terraform.ValueFromAutoFile:
		return "auto file"
	case terraform.ValueFromNamedFile:
		return "name file"
	case terraform.ValueFromCLIArg:
		return "CLI arg"
	case terraform.ValueFromEnvVar:
		return "env var"
	case terraform.ValueFromInput:
		return "user input"
	default:
		return "unknown"
	}
}
