package parse

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/steampipeconfig/hclhelpers"
)

type dependency struct {
	Range      hcl.Range
	Traversals []hcl.Traversal
}

func (d dependency) String() string {
	traversalStrings := make([]string, len(d.Traversals))
	for i, t := range d.Traversals {
		traversalStrings[i] = hclhelpers.TraversalAsString(t)
	}
	return fmt.Sprintf(`%s` /*d.Range.String(), */, strings.Join(traversalStrings, ","))
}

// struct to hold the result of a decoding operation
type decodeResult struct {
	Diags   hcl.Diagnostics
	Depends []*dependency
}

// Merge :: merge this decode result with another
func (p *decodeResult) Merge(other *decodeResult) *decodeResult {
	p.Diags = append(p.Diags, other.Diags...)
	p.Depends = append(p.Depends, other.Depends...)
	return p
}

// Success :: was the parsing successful - true if there are no errors and no dependencies
func (p *decodeResult) Success() bool {
	return !p.Diags.HasErrors() && len(p.Depends) == 0
}

// if the diags contains dependency errors, add dependencies to the result
// otherwise add diags to the result
func (p *decodeResult) handleDecodeDiags(diags hcl.Diagnostics) {
	for _, diag := range diags {
		if dependency := isDependencyError(diag); dependency != nil {
			// was this error caused by a missing dependency?
			p.Depends = append(p.Depends, dependency)
		}
	}
	// only register errors if there are NOT any missing variables
	if diags.HasErrors() && len(p.Depends) == 0 {
		p.addDiags(diags)
	}
}

func (p *decodeResult) addDiags(diags hcl.Diagnostics) {
	p.Diags = append(p.Diags, diags...)
}
