package modconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/utils"
)

// validate that the provider does not contains both edges/nodes and a query/sql
// enrich the loaded nodes and edges with the fully parsed resources from the resourceMapProvider
func initialiseEdgesAndNodes(p EdgeAndNodeProvider, resourceMapProvider ModResourcesProvider) hcl.Diagnostics {
	existingEdges := p.GetEdges()
	existingNodes := p.GetNodes()

	// validate that the resource does not declare both edges/nodes and sql/query
	providerDefinesQuery := p.GetSQL() != nil || p.GetQuery() != nil
	providerContainsEdgesOrNodes := (len(existingEdges) + len(existingNodes)) > 0
	if providerDefinesQuery && providerContainsEdgesOrNodes {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s contains edges/nodes AND has a query", p.Name()),
			Subject:  p.GetDeclRange(),
		}}

	}

	// when we reference resources (i.e. nodes/edges),
	// not all properties are retrieved as they are no cty serialisable
	// repopulate all nodes/edges from resourceMapProvider
	resourceMaps := resourceMapProvider.GetResourceMaps()

	parentRuntimeDependencies := utils.MapValues[*RuntimeDependency](p.GetRuntimeDependencies())
	var diags hcl.Diagnostics
	edges := make(DashboardEdgeList, len(existingEdges))
	for i, e := range existingEdges {
		fullEdge, ok := resourceMaps.DashboardEdges[e.Name()]
		if !ok {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("%s contains edge %s but this has not been loaded", p.Name(), e.Name()),
				Subject:  p.GetDeclRange(),
			}}

		}
		edges[i] = fullEdge
	}

	nodes := make(DashboardNodeList, len(existingNodes))
	for i, e := range existingNodes {
		fullNode, ok := resourceMaps.DashboardNodes[e.Name()]
		if !ok {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("%s contains node %s but this has not been loaded", p.Name(), e.Name()),
				Subject:  p.GetDeclRange(),
			}}
		}
		nodes[i] = fullNode
	}

	// write back the enriched nodes and edges
	p.SetNodes(nodes)
	p.SetEdges(edges)
	return nil
}
