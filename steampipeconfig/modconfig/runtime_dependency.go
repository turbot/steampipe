package modconfig

import (
	"fmt"
)

type RuntimeDependency struct {
	PropertyPath   *ParsedPropertyPath
	SourceResource HclResource
	ArgName        *string
	ArgIndex       *int
	IsDefault      bool
}

func (d *RuntimeDependency) String() string {
	if d.ArgIndex != nil {
		return fmt.Sprintf("arg.%d->%s", d.ArgIndex, d.PropertyPath.String())
	}
	if d.IsDefault {
		return fmt.Sprintf("param.%s.default->%s", *d.ArgName, d.PropertyPath.String())
	}

	return fmt.Sprintf("arg.%s->%s", *d.ArgName, d.PropertyPath.String())
}

func (d *RuntimeDependency) ResolveSource(dashboard *Dashboard, workspace ResourceMapsProvider) error {
	resourceName := d.PropertyPath.ToResourceName()
	var found bool
	var sourceResource HclResource
	// if this dependency has a 'self' prefix, resolve from the current dashboard container
	if d.PropertyPath.Scope == runtimeDependencyDashboardScope {
		sourceResource, found = dashboard.GetInput(resourceName)
	} else {
		// otherwise, resolve from the global inputs
		sourceResource, found = workspace.GetResourceMaps().GlobalDashboardInputs[resourceName]
	}
	if !found {
		return fmt.Errorf("could not resolve runtime dependency resource %s", d.PropertyPath)
	}

	d.SourceResource = sourceResource
	return nil
}
