package modconfig

import (
	"fmt"
)

type RuntimeDependency struct {
	PropertyPath *ParsedPropertyPath
	// the resolved resource which we depend on
	SourceResource HclResource
	ArgName        *string
	ArgIndex       *int
	// the resource which has the runtime dependency
	ParentResource QueryProvider
	// TACTICAL - if set, wrap the dependency value in an array
	// this provides support for args which convert a runtime dependency to an array, like:
	// arns = [input.arn]
	IsArray bool
}

func (d *RuntimeDependency) String() string {
	if d.ArgIndex != nil {
		return fmt.Sprintf("arg.%d->%s", *d.ArgIndex, d.PropertyPath.String())
	}

	return fmt.Sprintf("arg.%s->%s", *d.ArgName, d.PropertyPath.String())
}

func (d *RuntimeDependency) ResolveSource(dashboard *Dashboard, workspace ResourceMapsProvider) error {
	resourceName := d.PropertyPath.ToResourceName()
	var found bool
	var sourceResource HclResource
	switch {
	// if this is a 'with' resolve from the parent resource
	case d.PropertyPath.ItemType == BlockTypeWith:
		sourceResource, found = d.ParentResource.GetWith(resourceName)
	// if this dependency has a 'self' prefix, resolve from the current dashboard container
	case d.PropertyPath.Scope == runtimeDependencyDashboardScope:
		sourceResource, found = dashboard.GetInput(resourceName)

	default:
		// otherwise, resolve from the global inputs
		sourceResource, found = workspace.GetResourceMaps().GlobalDashboardInputs[resourceName]
	}
	if !found {
		return fmt.Errorf("could not resolve runtime dependency resource %s", d.PropertyPath)
	}

	d.SourceResource = sourceResource
	return nil
}

func (d *RuntimeDependency) Equals(other *RuntimeDependency) bool {
	// TargetPropertyPath
	if d.PropertyPath.PropertyPath == nil {
		if other.PropertyPath.PropertyPath != nil {
			return false
		}
	} else {
		// we have TargetPropertyPath
		if other.PropertyPath.PropertyPath == nil {
			return false
		}

		if len(d.PropertyPath.PropertyPath) != len(other.PropertyPath.PropertyPath) {
			return false
		}
		for i, c := range d.PropertyPath.PropertyPath {
			if other.PropertyPath.PropertyPath[i] != c {
				return false
			}
		}
	}

	// SourceResource
	if d.SourceResource.Name() != other.SourceResource.Name() {
		return false
	}

	return true
}

func (d *RuntimeDependency) SetParentResource(resource QueryProvider) {
	d.ParentResource = resource
}
