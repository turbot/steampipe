package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

func getChildNames(childNames []NamedItem, parentName string, block *hcl.Block) ([]string, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	// validate each child name appears only once
	nameMap := make(map[string]bool)
	childNameStrings := make([]string, len(childNames))

	for i, n := range childNames {
		if nameMap[n.Name] {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("'%s' has duplicate child name '%s'", parentName, n.Name),
				Subject:  &block.DefRange})
			continue
		}
		childNameStrings[i] = n.Name
		nameMap[n.Name] = true
	}

	return childNameStrings, diags
}
