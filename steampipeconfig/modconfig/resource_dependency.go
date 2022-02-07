package modconfig

import (
	"fmt"
	"strings"
)

type RuntimeDependency struct {
	PropertyPath     *ParsedPropertyPath
	SourceResource   HclResource
	TargetProperties []string
	Value            *string
}

func (d *RuntimeDependency) String() string {
	return fmt.Sprintf("%s->%s", strings.Join(d.TargetProperties, ","), d.PropertyPath.String())
}

func (d *RuntimeDependency) ResolveSource(resource HclResource, report *ReportContainer, workspace ResourceMapsProvider) error {
	// TODO THINK ABOUT REPORT PREFIX

	resourceName := d.PropertyPath.ToResourceName()
	var found bool
	// if this dependency has a 'root' prefix, resolve from the current report container
	if d.PropertyPath.Scope == runtimeDependencyReportScope {
		d.SourceResource, found = report.GetInput(resourceName)

	} else {
		// otherwise, resolve from the workspace
		d.SourceResource, found = workspace.GetResourceMaps().ReportInputs[resourceName]
	}
	if !found {
		return fmt.Errorf("could not resolve runtime dependency resource %s", d.PropertyPath)
	}

	return nil
}
