package modconfig

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/hclhelpers"
)

type ResourceDependency struct {
	Range      hcl.Range
	Traversals []hcl.Traversal
}

func (d *ResourceDependency) String() string {
	traversalStrings := make([]string, len(d.Traversals))
	for i, t := range d.Traversals {
		traversalStrings[i] = hclhelpers.TraversalAsString(t)
	}
	return strings.Join(traversalStrings, ",")
}

func (d *ResourceDependency) IsRuntimeDependency() bool {
	// runtime dependency wil only have a single traversal
	if len(d.Traversals) > 1 {
		return false
	}
	// parse the traversal as a property path
	propertyPath, err := ParseResourcePropertyPath(hclhelpers.TraversalAsString(d.Traversals[0]))
	if err != nil {
		return false
	}

	return isRunTimeDependencyProperty(propertyPath)

}

func isRunTimeDependencyProperty(propertyPath *ParsedPropertyPath) bool {
	// supported runtime dependencies
	// map is keyed by resource type and contains a list of properties
	runTimeDependencyPropertyPaths := map[string][]string{
		"input": {"value"},
		"param": {"value"},
	}
	// is this property a supported runtime dependency property
	if supportedProperties, ok := runTimeDependencyPropertyPaths[propertyPath.ItemType]; ok {
		return helpers.StringSliceContains(supportedProperties, propertyPath.PropertyPathString())
	}
	return false
}

// getPropertiesFromContent finds any attributes in the given content which depend on this dependency
func (d *ResourceDependency) getPropertiesFromContent(content *hcl.BodyContent) []string {
	var res []string
	for _, a := range content.Attributes {
		vars := a.Expr.Variables()
		if len(d.Traversals) != len(vars) {
			break
		}
		// build map of paths
		var traversalMap = make(map[string]bool, len(vars))
		for _, t := range vars {
			traversalMap[hclhelpers.TraversalAsString(t)] = true
		}
		for _, t := range d.Traversals {
			if !traversalMap[hclhelpers.TraversalAsString(t)] {
				return res
			}
		}

		// ok so traversals match!
		res = append(res, a.Name)
	}
	return res
}
