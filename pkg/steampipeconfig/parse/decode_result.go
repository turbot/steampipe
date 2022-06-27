package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// struct to hold the result of a decoding operation
type decodeResult struct {
	Diags   hcl.Diagnostics
	Depends map[string]*modconfig.ResourceDependency
}

func newDecodeResult() *decodeResult {
	return &decodeResult{Depends: make(map[string]*modconfig.ResourceDependency)}
}

// Merge merges this decode result with another
func (p *decodeResult) Merge(other *decodeResult) *decodeResult {
	p.Diags = append(p.Diags, other.Diags...)
	for k, v := range other.Depends {
		p.Depends[k] = v
	}

	return p
}

// Success returns if the was parsing successful - true if there are no errors and no dependencies
func (p *decodeResult) Success() bool {
	return !p.Diags.HasErrors() && len(p.Depends) == 0
}

// if the diags contains dependency errors, add dependencies to the result
// otherwise add diags to the result
func (p *decodeResult) handleDecodeDiags(diags hcl.Diagnostics) {
	for _, diag := range diags {
		if dependency := diagsToDependency(diag); dependency != nil {
			p.Depends[dependency.String()] = dependency
		}
	}
	// only register errors if there are NOT any missing variables
	if diags.HasErrors() && len(p.Depends) == 0 {
		p.addDiags(diags)
	}
}

// determine whether the diag is a dependency error, and if so, return a dependency object
func diagsToDependency(diag *hcl.Diagnostic) *modconfig.ResourceDependency {
	if helpers.StringSliceContains(missingVariableErrors, diag.Summary) {
		return &modconfig.ResourceDependency{Range: diag.Expression.Range(), Traversals: diag.Expression.Variables()}
	}
	return nil
}

func (p *decodeResult) addDiags(diags hcl.Diagnostics) {
	p.Diags = append(p.Diags, diags...)
}
