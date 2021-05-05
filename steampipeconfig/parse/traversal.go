package parse

import (
	"fmt"

	"github.com/turbot/go-kit/helpers"

	"github.com/hashicorp/hcl/v2"
)

// IsQualifiedTraversal :: a 'qualified traversal' is of form
// <mod>.<query|action|policy>.<name>.xxx.xxx
func IsQualifiedTraversal(traversal hcl.Traversal) bool {
	if len(traversal) < 3 {
		return false
	}
	s := traversal.SimpleSplit()
	if isReferenceable(s.Abs.RootName()) {
		return false
	}
	return isReferenceable(s.Rel[0].(hcl.TraverseAttr).Name)
}

// TraversalAsString :: convert a traversal to a path string
func TraversalAsString(traversal hcl.Traversal) string {
	s := traversal.SimpleSplit()
	name := s.Abs.RootName()
	for _, r := range s.Rel {
		name += fmt.Sprintf(".%s", r.(hcl.TraverseAttr).Name)
	}
	return name
}

func isReferenceable(name string) bool {
	// TODO USE block types
	return helpers.StringSliceContains([]string{"mod", "control", "query", "benchmark"}, name)
}
