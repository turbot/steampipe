package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// validate that the provider does not contains both edges/nodes and a query/sql
// enrich the loaded nodes and edges with the fully parsed resources from the resourceMapProvider
func validateNodeAndEdgeProvider(resource modconfig.NodeAndEdgeProvider) hcl.Diagnostics {
	existingEdges := resource.GetEdges()
	existingNodes := resource.GetNodes()

	// validate that the resource does not declare both edges/nodes and sql/query
	providerDefinesQuery := resource.GetSQL() != nil || resource.GetQuery() != nil
	providerContainsEdgesOrNodes := (len(existingEdges) + len(existingNodes)) > 0
	if providerDefinesQuery && providerContainsEdgesOrNodes {
		return hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s contains edges/nodes AND has a query", resource.Name()),
			Subject:  resource.GetDeclRange(),
		}}
	}

	return nil
}

func validateQueryProvider(resource modconfig.QueryProvider) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if resource.GetSQL() != nil && resource.GetQuery() != nil {
		// either Query or SQL property may be set -  if Query property already set, error
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s has both 'SQL' and 'query' property set - only 1 of these may be set", resource.Name()),
			Subject:  resource.GetDeclRange(),
		})
	}

	// param block cannot be set if a query property is set - it is only valid if inline SQL ids defined
	if len(resource.GetParams()) > 0 && resource.GetQuery() != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s has 'query' property set so cannot define param blocks", resource.Name()),
			Subject:  resource.GetDeclRange(),
		})
	}
	return diags
}
