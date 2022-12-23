package modconfig

import (
	"fmt"
)

type RuntimeDependency struct {
	PropertyPath       *ParsedPropertyPath
	TargetPropertyName *string
	// TACTICAL the name of the parent property - either "args" or "param.<name>"
	ParentPropertyName  string
	TargetPropertyIndex *int

	// TACTICAL - if set, wrap the dependency value in an array
	// this provides support for args which convert a runtime dependency to an array, like:
	// arns = [input.arn]
	IsArray bool

	// resource which provides has the dependency
	Provider HclResource
}

func (d *RuntimeDependency) SourceResourceName() string {
	return d.PropertyPath.ToResourceName()
}

func (d *RuntimeDependency) String() string {
	if d.TargetPropertyIndex != nil {
		return fmt.Sprintf("%s.%d->%s", d.ParentPropertyName, *d.TargetPropertyIndex, d.PropertyPath.String())
	}

	return fmt.Sprintf("%s.%s->%s", d.ParentPropertyName, *d.TargetPropertyName, d.PropertyPath.String())
}

func (d *RuntimeDependency) ValidateSource(dashboard *Dashboard, workspace ResourceMapsProvider) error {
	// TODO  [node_reuse] re-add parse time validation https://github.com/turbot/steampipe/issues/2925
	//resourceName := d.PropertyPath.ToResourceName()
	//var found bool
	////var sourceResource HclResource
	//switch d.PropertyPath.ItemType {
	//// if this is a 'with' resolve from the parent resource
	//case BlockTypeParam:
	//	_, found = d.ParentResource.ResolveWithFromTree(resourceName)
	//case BlockTypeWith:
	//	_, found = d.ParentResource.ResolveWithFromTree(resourceName)
	//// if this dependency has a 'self' prefix, resolve from the current dashboard container
	//case BlockTypeInput:
	//	_, found = dashboard.GetInput(resourceName)
	//
	//	//default:
	//	//	// otherwise, resolve from the global inputs
	//	//	_, found = workspace.GetResourceMaps().GlobalDashboardInputs[resourceName]
	//}
	//if !found {
	//	return fmt.Errorf("could not resolve runtime dependency resource %s", d.PropertyPath)
	//}

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

	if d.SourceResourceName() != other.SourceResourceName() {
		return false
	}

	return true
}
