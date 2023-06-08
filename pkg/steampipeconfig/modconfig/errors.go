package modconfig

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/utils"
)

type MissingVariable struct {
	Variable *Variable
	Path     string
}

type MissingVariableError struct {
	MissingVariables           []*MissingVariable
	MissingTransitiveVariables []*MissingVariable
}

func (m *MissingVariableError) Error() string {
	allMissing := append(m.MissingVariables, m.MissingTransitiveVariables...)
	missingCount := len(allMissing)
	missingPaths := make([]string, missingCount)
	for i, v := range allMissing {
		missingPaths[i] = v.Path
	}

	return fmt.Sprintf("missing %d variable %s:\n\t%s",
		missingCount,
		utils.Pluralize("value", missingCount),
		strings.Join(missingPaths, "\n\t"),
	)
}

func (m *MissingVariableError) Add(missingVars []*MissingVariable, isTopLevel bool) {
	if isTopLevel {
		m.MissingVariables = append(m.MissingVariables, missingVars...)
	} else {
		m.MissingTransitiveVariables = append(m.MissingTransitiveVariables, missingVars...)
	}
}

type VariableValidationFailedError struct {
}

func (m VariableValidationFailedError) Error() string {
	return "variable validation failed"
}
