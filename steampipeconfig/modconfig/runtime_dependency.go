package modconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/steampipeconfig/hclhelpers"
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

// IsRunTimeDependency determines whether this is a runtime dependency
// adependency is run time if:
// - there is a single traversal
// - the property referenced is one of the defined runtime dependency properties
// - the dependency resource exists in the mod
func (d *ResourceDependency) IsRunTimeDependency() bool {
	if len(d.Traversals) > 1 {
		return false
	}
	parsedPropertyPath, err := ParseResourcePropertyPath(hclhelpers.TraversalAsString(d.Traversals[0]))
	if err != nil {
		return false
	}

	// supported runtime dependencies
	// map is keyed by resource type and contains a list of properties
	runTimeDependencyPropertyPaths := map[string][]string{"input": {"result"}}

	// is this property a supported runtime dependency property
	if supportedProperties, ok := runTimeDependencyPropertyPaths[parsedPropertyPath.ItemType]; ok {
		if helpers.StringSliceContains(supportedProperties, parsedPropertyPath.PropertyPathString()) {
			return true
		}
	}

	return false
}

func (d *ResourceDependency) ToRuntimeDependency(bodyContent *hcl.BodyContent) (*RuntimeDependency, error) {
	if bodyContent == nil {
		return nil, fmt.Errorf("nil body passed to ToRuntimeDependency")
	}
	res := &RuntimeDependency{
		TargetProperties: d.getPropertiesFromContent(bodyContent),
		Traversal:        d.Traversals[0],
	}
	if len(res.TargetProperties) == 0 {
		return nil, fmt.Errorf("failed to resolve any properties using dependency %s", d)
	}
	return res, nil
}

// getPropertiesFromContent finds any attributes in the given content which depend on this dependency
func (d *ResourceDependency) getPropertiesFromContent(content *hcl.BodyContent) []string {
	var res []string
	for _, a := range content.Attributes {
		if scopeTraversal, ok := a.Expr.(*hclsyntax.ScopeTraversalExpr); ok {
			if len(d.Traversals) == 1 &&
				hclhelpers.TraversalsEqual(d.Traversals[0], scopeTraversal.Traversal) {
				res = append(res, a.Name)
			}
		}
	}
	return res
}
