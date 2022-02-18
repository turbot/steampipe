package modconfig

import (
	"fmt"
	"strings"
	"sync"
)

type RuntimeDependency struct {
	PropertyPath       *ParsedPropertyPath
	SourceResource     RuntimeDependencySource
	TargetPropertyPath []string
	// function to set the target
	SetTargetFunc func(string)

	// the resolved value
	value     *string
	valueLock sync.Mutex
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
		// otherwise, resolve from the workspace
		sourceResource, found = workspace.GetResourceMaps().DashboardInputs[resourceName]
	}
	if !found {
		return fmt.Errorf("could not resolve runtime dependency resource %s", d.PropertyPath)
	}

	// cast source to RuntimeDependencySource
	d.SourceResource = sourceResource.(RuntimeDependencySource)
	return nil
}

func (d *RuntimeDependency) IsResolved() bool {
	d.valueLock.Lock()
	defer d.valueLock.Unlock()
	return d.value != nil
}

func (d *RuntimeDependency) Resolve() bool {
	d.valueLock.Lock()
	defer d.valueLock.Unlock()

	// did we succeed?
	d.value = d.SourceResource.GetValue()

	if d.value != nil {
		d.SetTargetFunc(*d.value)
		return true
	}

	return false
}
