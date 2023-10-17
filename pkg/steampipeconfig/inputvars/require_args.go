package inputvars

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"
	"github.com/turbot/terraform-components/terraform"
	"github.com/turbot/terraform-components/tfdiags"
)

const ValueFromModFile terraform.ValueSourceType = 'M'

func CollectVariableValuesFromModRequire(m *modconfig.Mod, lock *versionmap.WorkspaceLock) (terraform.InputValues, error) {
	res := make(terraform.InputValues)
	if m.Require != nil {
		for _, depModConstraint := range m.Require.Mods {
			if args := depModConstraint.Args; args != nil {
				// find the loaded dep mod which satisfies this constraint
				resolvedConstraint := lock.GetMod(depModConstraint.Name, m)
				if resolvedConstraint == nil {
					return nil, fmt.Errorf("dependency mod %s is not loaded", depModConstraint.Name)
				}
				for varName, varVal := range args {
					varFullName := fmt.Sprintf("%s.var.%s", resolvedConstraint.Alias, varName)

					sourceRange := tfdiags.SourceRange{
						Filename: m.Require.DeclRange.Filename,
						Start: tfdiags.SourcePos{
							Line:   m.Require.DeclRange.Start.Line,
							Column: m.Require.DeclRange.Start.Column,
							Byte:   m.Require.DeclRange.Start.Byte,
						},
						End: tfdiags.SourcePos{
							Line:   m.Require.DeclRange.End.Line,
							Column: m.Require.DeclRange.End.Column,
							Byte:   m.Require.DeclRange.End.Byte,
						},
					}

					res[varFullName] = &terraform.InputValue{
						Value:       varVal,
						SourceType:  ValueFromModFile,
						SourceRange: sourceRange,
					}
				}
			}
		}
	}
	return res, nil
}
