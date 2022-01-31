package parse

import (
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/hashicorp/hcl/v2"
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
func (p *decodeResult) handleDecodeDiags(bodyContent *hcl.BodyContent, resource modconfig.HclResource, diags hcl.Diagnostics, runCtx *RunContext) {
	var allDependencies []*modconfig.ResourceDependency
	for _, diag := range diags {
		if dependency := isDependencyError(diag); dependency != nil {
			allDependencies = append(allDependencies, dependency)

			// so it was a dependency error - determine whether this is a RUN TIMEdependency
			// - if so, do not raise a dependency error but instead store in the resources run time dependencies
			if dependency.IsRunTimeDependency() {
				if err := dependency.SetAsRuntimeDependency(bodyContent); err != nil {
					resource.AddRuntimeDependencies(dependency)
				}

			} else {
				// was this error caused by a missing dependency?
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
