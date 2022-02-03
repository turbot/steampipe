package modconfig

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/steampipeconfig/hclhelpers"
)

type RuntimeDependency struct {
	PropertyPath     *ParsedPropertyPath
	SourceResource   HclResource
	TargetProperties []string
}

func (d *RuntimeDependency) String() string {
	return fmt.Sprintf("%s->%s", strings.Join(d.TargetProperties, ","), d.PropertyPath.String())
}

func (d *RuntimeDependency) ResolveSource(resource HclResource, report *ReportContainer, workspace ResourceMapsProvider) error {
	// NOTE: for now this code assumes the runtime dependency resource is a ReportInput
	// as this is (currently) the only type
	// when the is expanded, this code will beed to change

	var found bool
	dependencyPath, err := ParseResourcePropertyPath(hclhelpers.TraversalAsString(nil))
	if err != nil {
		return err
	}
	resourceName := dependencyPath.ToResourceName()

	// if this dependency has a 'root' prefix, resolve from the current report container
	if dependencyPath.Scope == runtimeDependencyRootScope {
		// TODO assume input only not param for now
		_, found = report.GetInput(resourceName)
		d.SourceResource = report
	} else if dependencyPath.Scope == runtimeDependencyParentScope {
		panic("parent not supported")
	} else {
		// otherwise, resolve from the workspace
		d.SourceResource, found = workspace.GetResourceMaps().ReportInputs[resourceName]
	}
	if !found {
		return fmt.Errorf("could not resolve runtime depdency resource %s", dependencyPath)
	}

	return nil
}
