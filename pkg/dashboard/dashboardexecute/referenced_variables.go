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
				addReferencedVars(resource.GetReferences())
				return true, nil
			},
		)
	case *CheckRun:
		benchmark, ok := r.DashboardNode.(*modconfig.Benchmark)
		if !ok {
			// not expected
			break
		}
		benchmark.WalkResources(
			func(resource modconfig.ModTreeItem) (bool, error) {
				addReferencedVars(resource.(modconfig.HclResource).GetReferences())
				return true, nil
			},
		)
	}

	return referencedVariables
}
