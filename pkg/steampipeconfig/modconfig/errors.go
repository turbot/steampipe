package modconfig

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/utils"
)

type MissingVariableError struct {
	MissingVariables []*Variable
}

func (m MissingVariableError) Error() string {
	strs := make([]string, len(m.MissingVariables))
	for i, v := range m.MissingVariables {
		strs[i] = v.Name()
	}
	return fmt.Sprintf("missing %d variable %s: %s", len(strs), utils.Pluralize("value", len(strs)), strings.Join(strs, ","))
}

type VariableValidationFailedError struct {
}

func (m VariableValidationFailedError) Error() string {
	return "variable validation failed"
}
