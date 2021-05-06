package parse

import "github.com/hashicorp/hcl/v2"

type dependency struct {
	Range      hcl.Range
	Traversals []hcl.Traversal
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
