package hclhelpers

import (
	"github.com/turbot/steampipe/pkg/type_conversion"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

// TraversalAsString converts a traversal to a path string
// (if an absolute traversal is passed - convert to relative)
func TraversalAsString(traversal hcl.Traversal) string {
	var parts = make([]string, len(traversal))
	offset := 0

	if !traversal.IsRelative() {
		s := traversal.SimpleSplit()
		parts[0] = s.Abs.RootName()
		offset++
		traversal = s.Rel
	}
	for i, r := range traversal {
		switch t := r.(type) {
		case hcl.TraverseAttr:
			parts[i+offset] = t.Name
		case hcl.TraverseIndex:
			idx, err := type_conversion.CtyToString(t.Key)
			if err != nil {
				// we do not expect this to fail
				continue
			}
			parts[i+offset] = idx
		}
	}
	return strings.Join(parts, ".")
}

func TraversalsEqual(t1, t2 hcl.Traversal) bool {
	return TraversalAsString(t1) == TraversalAsString(t2)
}

// ResourceNameFromTraversal converts a traversal to the name of the referenced resource
// We must take into account possible mod-name as first traversal element
func ResourceNameFromTraversal(resourceType string, traversal hcl.Traversal) (string, bool) {
	traversalString := TraversalAsString(traversal)
	split := strings.Split(traversalString, ".")

	// the resource reference will be of the form
	// var.<var_name>
	// or
	// <resource_type>.<resource_name>.<property>
	// or
	// <mod_name>.<resource_type>.<resource_name>.<property>

	if split[0] == "var" {
		return strings.Join(split, "."), true
	}
	if len(split) >= 2 && split[0] == resourceType {
		return strings.Join(split[:2], "."), true
	}
	if len(split) >= 3 && split[1] == resourceType {
		return strings.Join(split[:3], "."), true
	}
	return "", false
}
