package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
)

// validate the resource
func validateResource(resource modconfig.HclResource) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if qp, ok := resource.(modconfig.NodeAndEdgeProvider); ok {
		moreDiags := validateNodeAndEdgeProvider(qp)
		diags = append(diags, moreDiags...)
	} else if qp, ok := resource.(modconfig.QueryProvider); ok {
		moreDiags := validateQueryProvider(qp)
		diags = append(diags, moreDiags...)
	}

	if wp, ok := resource.(modconfig.WithProvider); ok {
		moreDiags := validateRuntimeDependencyProvider(wp)
		diags = append(diags, moreDiags...)
	}
	return diags
}

func validateRuntimeDependencyProvider(wp modconfig.WithProvider) hcl.Diagnostics {
	resource := wp.(modconfig.HclResource)
	var diags hcl.Diagnostics
	if len(wp.GetWiths()) > 0 && !resource.IsTopLevel() {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Only top level resources can have `with` blocks",
			Detail:   fmt.Sprintf("%s contains 'with' blocks but is not a top level resource.", resource.Name()),
			Subject:  resource.GetDeclRange(),
		})
	}
	return diags
}

// validate that the provider does not contains both edges/nodes and a query/sql
// enrich the loaded nodes and edges with the fully parsed resources from the resourceMapProvider
func validateNodeAndEdgeProvider(resource modconfig.NodeAndEdgeProvider) hcl.Diagnostics {
	// TODO [node_reuse] add NodeAndEdgeProviderImpl and move validate there
	// https://github.com/turbot/steampipe/issues/2918

	var diags hcl.Diagnostics
	containsEdgesOrNodes := len(resource.GetEdges())+len(resource.GetNodes()) > 0
	definesQuery := resource.GetSQL() != nil || resource.GetQuery() != nil

	// cannot declare both edges/nodes AND sql/query
	if definesQuery && containsEdgesOrNodes {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s contains edges/nodes AND has a query", resource.Name()),
			Subject:  resource.GetDeclRange(),
		})
	}

	// if resource is NOT top level must have either edges/nodes OR sql/query
	if !resource.IsTopLevel() && !definesQuery && !containsEdgesOrNodes {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s does not define a query or SQL, and has no edges/nodes", resource.Name()),
			Subject:  resource.GetDeclRange(),
		})
	}

	diags = append(diags, validateSqlAndQueryNotBothSet(resource)...)

	diags = append(diags, validateParamAndQueryNotBothSet(resource)...)

	return diags
}

func validateQueryProvider(resource modconfig.QueryProvider) hcl.Diagnostics {
	var diags hcl.Diagnostics

	diags = append(diags, resource.ValidateQuery()...)

	diags = append(diags, validateSqlAndQueryNotBothSet(resource)...)

	diags = append(diags, validateParamAndQueryNotBothSet(resource)...)

	return diags
}

func validateParamAndQueryNotBothSet(resource modconfig.QueryProvider) hcl.Diagnostics {
	var diags hcl.Diagnostics

	// param block cannot be set if a query property is set - it is only valid if inline SQL ids defined
	if len(resource.GetParams()) > 0 {
		if resource.GetQuery() != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  fmt.Sprintf("Deprecated usage: %s has 'query' property set so should not define 'param' blocks", resource.Name()),
				Subject:  resource.GetDeclRange(),
			})
		}
		if !resource.IsTopLevel() && !resource.ParamsInheritedFromBase() {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  "Deprecated usage: Only top level resources can have 'param' blocks",
				Detail:   fmt.Sprintf("%s contains 'param' blocks but is not a top level resource.", resource.Name()),
				Subject:  resource.GetDeclRange(),
			})
		}
	}
	return diags
}

func validateSqlAndQueryNotBothSet(resource modconfig.QueryProvider) hcl.Diagnostics {
	var diags hcl.Diagnostics
	// are both sql and query set?
	if resource.GetSQL() != nil && resource.GetQuery() != nil {
		// either Query or SQL property may be set -  if Query property already set, error
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s has both 'SQL' and 'query' property set - only 1 of these may be set", resource.Name()),
			Subject:  resource.GetDeclRange(),
		})
	}
	return diags
}
