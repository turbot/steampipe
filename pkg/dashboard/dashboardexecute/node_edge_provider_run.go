package dashboardexecute

import (
	"encoding/json"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"log"
)

type NodeEdgeProviderRun struct {
	LeafRun
	Properties map[string]any `json:"-"`
}

func NewNodeEdgeProviderRun(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardNodeParent, executionTree *DashboardExecutionTree) (*NodeEdgeProviderRun, error) {
	leafRun, err := NewLeafRun(resource, parent, executionTree)
	if err != nil {
		return nil, err
	}
	res := &NodeEdgeProviderRun{
		LeafRun: *leafRun,
	}

	// HACK
	j, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(j, &res.Properties)

	// now populate node and edge names
	for _, c := range leafRun.GetChildren() {
		childResource := c.(*LeafRun).DashboardNode
		var childKey string
		log.Printf("[WARN] NewNodeEdgeProviderRun resource %s %p", childResource.Name(), childResource)

		switch childResource.(type) {
		case *modconfig.DashboardNode:
			childKey = "nodes"
		case *modconfig.DashboardEdge:
			childKey = "edges"
		}
		// add this child to the appropriate array
		target, _ := res.Properties[childKey].([]string)
		if target == nil {
			target = []string{}
		}
		res.Properties[childKey] = append(target, c.GetName())
	}
	return res, nil
}
