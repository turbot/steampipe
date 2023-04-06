package dashboardexecute

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/workspace"
	"strings"
)

// GetReferencedVariables builds map of variables values containing only those mod variables which are referenced
// NOTE: we refer to variables in dependency mods in the format which is valid for an SPVARS filer, i.e.
// <mod>.<var-name>
// the VariableValues map will contain these variables with the name format <mod>.var.<var-name>,
// so we must convert the name
func GetReferencedVariables(root dashboardtypes.DashboardTreeRun, w *workspace.Workspace) map[string]string {
	var referencedVariables = make(map[string]string)

	// TODO KAI UPDATE TO HANDLE DEPENDENCYPATH
	addReferencedVars := func(refs []*modconfig.ResourceReference) {
		for _, ref := range refs {
			parts := strings.Split(ref.To, ".")
			if len(parts) == 2 && parts[0] == "var" {
				varName := parts[1]
				varValueName := varName
				// NOTE: if the ref is NOT for the workspace mod, then use the fully qualifed name
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
		r.dashboard.WalkResources(
			func(resource modconfig.HclResource) (bool, error) {
				if resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata); ok {
					addReferencedVars(resourceWithMetadata.GetReferences())
				}
				return true, nil
			},
		)
	case *CheckRun:
		switch n := r.resource.(type) {
		case *modconfig.Benchmark:
			n.WalkResources(
				func(resource modconfig.ModTreeItem) (bool, error) {
					if resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata); ok {
						addReferencedVars(resourceWithMetadata.GetReferences())
					}
					return true, nil
				},
			)
		case *modconfig.Control:
			addReferencedVars(n.GetReferences())
		}
	}

	return referencedVariables
}
