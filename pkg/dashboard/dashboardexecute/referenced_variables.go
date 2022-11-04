package dashboardexecute

import (
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/workspace"
	"strings"
)

// GetReferencedVariables builds map of variables values containing only those mod variables which are referenced
func GetReferencedVariables(root dashboardtypes.DashboardNodeRun, w *workspace.Workspace) map[string]string {
	var referencedVariables = make(map[string]string)

	addReferencedVars := func(refs []*modconfig.ResourceReference) {
		for _, ref := range refs {
			parts := strings.Split(ref.To, ".")
			if len(parts) == 2 && parts[0] == "var" {
				varName := parts[1]
				referencedVariables[varName] = w.VariableValues[varName]
			}
		}
	}

	switch r := root.(type) {
	case *DashboardRun:
		r.dashboardNode.WalkResources(
			func(resource modconfig.HclResource) (bool, error) {
				if resourceWithMetadata, ok := resource.(modconfig.ResourceWithMetadata); ok {
					addReferencedVars(resourceWithMetadata.GetReferences())
				}
				return true, nil
			},
		)
	case *CheckRun:
		switch n := r.DashboardNode.(type) {
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
