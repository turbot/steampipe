package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// struct to hold the result of a decoding operation
type decodeResult struct {
	Diags   hcl.Diagnostics
	Depends []*modconfig.ResourceDependency
}

// Merge merges this decode result with another
func (p *decodeResult) Merge(other *decodeResult) *decodeResult {
	p.Diags = append(p.Diags, other.Diags...)
	p.Depends = append(p.Depends, other.Depends...)
	return p
}

// Success :: was the parsing successful - true if there are no errors and no dependencies
func (p *decodeResult) Success() bool {
	return !p.Diags.HasErrors() && len(p.Depends) == 0
}

// if the diags containsdependency errors, add dependencies to the result
// otherwise add diags to the result
func (p *decodeResult) handleDecodeDiags(bodyContent *hcl.BodyContent, resource modconfig.HclResource, diags hcl.Diagnostics) {
	var allDependencies []*modconfig.ResourceDependency
	for _, diag := range diags {
		if dependency := isDependencyError(diag); dependency != nil {
			allDependencies = append(allDependencies, dependency)

			// so it was a dependency error - determine whether this is a RUN TIME dependency
			// - if so, do not raise a dependency error but instead store in the resources run time dependencies
			if runtimeDependency := dependency.ToRuntimeDependency(bodyContent); runtimeDependency != nil {
				// resource must be convertible to a ReportLeafNode
				// - these are the only resources to support runtime dependencies
				leafNode, ok := resource.(modconfig.ReportLeafNode)
				if !ok {
					p.addDiags(hcl.Diagnostics{&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  fmt.Sprintf("invalid resource type %s declares a runtime depdnency - only ReportLeafNodes may use them", resource.Name()),
						Subject:  resource.GetDeclRange(),
					}})
				}
				leafNode.AddRuntimeDependencies(runtimeDependency)
			} else {
				// this is not a runtime dependency - register a normal dependency
				p.Depends = append(p.Depends, dependency)
			}
		}
	}
	// only register errors if there are NOT any missing variables
	if diags.HasErrors() && len(allDependencies) == 0 {
		p.addDiags(diags)
	}
}

// determine whether the diag is a dependency error, and if so, return a dependency object
func isDependencyError(diag *hcl.Diagnostic) *modconfig.ResourceDependency {
	if helpers.StringSliceContains(missingVariableErrors, diag.Summary) {
		return &modconfig.ResourceDependency{Range: diag.Expression.Range(), Traversals: diag.Expression.Variables()}
	}
	return nil
}

func (p *decodeResult) addDiags(diags hcl.Diagnostics) {
	p.Diags = append(p.Diags, diags...)
}
