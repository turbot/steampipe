package modconfig

import (
	"fmt"
	"strings"
)

type RuntimeDependency struct {
	PropertyPath       *ParsedPropertyPath
	SourceResource     HclResource
	TargetPropertyPath []string
	// function to set the target
	SetTargetFunc func(string)
}

func (d *RuntimeDependency) String() string {
	return fmt.Sprintf("%s->%s", strings.Join(d.TargetPropertyPath, "."), d.PropertyPath.String())
}

func (d *RuntimeDependency) ResolveSource(dashboard *Dashboard, workspace ResourceMapsProvider) error {
	// TODO THINK ABOUT REPORT PREFIX

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
