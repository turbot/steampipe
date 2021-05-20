package hclhelpers

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

// TraversalAsString converts a traversal to a path string
func TraversalAsString(traversal hcl.Traversal) string {
	s := traversal.SimpleSplit()
	name := s.Abs.RootName()
	for _, r := range s.Rel {
		name += fmt.Sprintf(".%s", r.(hcl.TraverseAttr).Name)
	}
	return name
}

// ResourceNameFromTraversal converts a traversal to the name of the referenced resource
// We must take into account possible mod-name as first traversal element
func ResourceNameFromTraversal(resource string, traversal hcl.Traversal) (string, bool) {
	traversalString := TraversalAsString(traversal)
	split := strings.Split(traversalString, ".")

	// the resource reference  will be of the form
	// <resource_type>.<resource_name>.<property>
	// or
	// <mod_name>.<resource_type>.<resource_name>.<property>

	if split[0] == resource && len(split) >= 2 {
		return strings.Join(split[:2], "."), true
	}
	if split[1] == resource && len(split) >= 3 {
		return strings.Join(split[:3], "."), true
	}
	return "", false
}
