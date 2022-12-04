package modconfig

import (
	"fmt"
)

type RuntimeDependency struct {
	PropertyPath *ParsedPropertyPath
	// the resolved resource which we depend on
	// get rid of this and resolve at runtime????
	//SourceResourceName string
	ArgName  *string
	ArgIndex *int
	// the resource which has the runtime dependency
	ParentResource QueryProvider
	// TACTICAL - if set, wrap the dependency value in an array
	// this provides support for args which convert a runtime dependency to an array, like:
	// arns = [input.arn]
	IsArray bool
}

func (d *RuntimeDependency) SourceResourceName() string {
	return d.PropertyPath.ToResourceName()
}
func (d *RuntimeDependency) String() string {
	if d.ArgIndex != nil {
		return fmt.Sprintf("arg.%d->%s", *d.ArgIndex, d.PropertyPath.String())
	}

	return fmt.Sprintf("arg.%s->%s", *d.ArgName, d.PropertyPath.String())
}

func (d *RuntimeDependency) ValidateSource(dashboard *Dashboard, workspace ResourceMapsProvider) error {
	//resourceName := d.PropertyPath.ToResourceName()
	var found bool
	// TODO KAI validate source resource in resource tree
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
	//	//	// TODO CAN WE REMOVE THIS AS GLOBAL INPUTS SHOULD NOT BE REFERENCED DIRECTLY
	//	//	// otherwise, resolve from the global inputs
	//	//	_, found = workspace.GetResourceMaps().GlobalDashboardInputs[resourceName]
	//}
	if !found {
		return fmt.Errorf("could not resolve runtime dependency resource %s", d.PropertyPath)
	}

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

	// TODO
	//// SourceResource
	//if d.SourceResource != other.SourceResource {
	//	return false
	//}

	return true
}

func (d *RuntimeDependency) SetParentResource(resource QueryProvider) {
	d.ParentResource = resource
}
