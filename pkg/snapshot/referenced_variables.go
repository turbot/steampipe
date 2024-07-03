package snapshot

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/workspace"
)

// GetReferencedVariables builds map of variables values containing only those mod variables which are referenced
// NOTE: we refer to variables in dependency mods in the format which is valid for an SPVARS filer, i.e.
// <mod>.<var-name>
// the VariableValues map will contain these variables with the name format <mod>.var.<var-name>,
// so we must convert the name
func GetReferencedVariables(root dashboardtypes.DashboardTreeRun, w *workspace.Workspace) map[string]string {
	var referencedVariables = make(map[string]string)

	addReferencedVars := func(refs []*modconfig.ResourceReference) {
		for _, ref := range refs {
			parts := strings.Split(ref.To, ".")
			if len(parts) == 2 && parts[0] == "var" {
				varName := parts[1]
				varValueName := varName
				// NOTE: if the ref is NOT for the workspace mod, then use the qualified variable name
				// (e.g. aws_insights.var.v1)
				if refMod := ref.GetMetadata().ModName; refMod != w.Mod.ShortName {
					varValueName = fmt.Sprintf("%s.var.%s", refMod, varName)
					varName = fmt.Sprintf("%s.%s", refMod, varName)
				}
				referencedVariables[varName] = w.VariableValues[varValueName]
			}
		}
	}

	switch r := root.(type) {
	case *DashboardRun:
		//nolint:errcheck // we don't care about errors here, since the callback does not return an error
		r.dashboard.WalkResources(
			func(resource modconfig.HclResource) (bool, error) {
				if resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata); ok {
					addReferencedVars(resourceWithMetadata.GetReferences())
				}
				return true, nil
			},
		)
	}

	return referencedVariables
}
