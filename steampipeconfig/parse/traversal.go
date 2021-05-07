package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// TraversalAsString :: convert a traversal to a path string
func TraversalAsString(traversal hcl.Traversal) string {
	s := traversal.SimpleSplit()
	name := s.Abs.RootName()
	for _, r := range s.Rel {
		name += fmt.Sprintf(".%s", r.(hcl.TraverseAttr).Name)
	}
	return name
}
