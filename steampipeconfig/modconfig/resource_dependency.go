package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/steampipeconfig/hclhelpers"
)

type RuntimeDependency struct {
	Traversal        hcl.Traversal
	SourceResource   HclResource
	TargetProperties []string
}

func (d *RuntimeDependency) String() string {
	return hclhelpers.TraversalAsString(d.Traversal)
}

func (d *RuntimeDependency) ResolveResource(c *ReportContainer, workspace ResourceMapsProvider) error {
	// NOTE: for now this code assumes the runtime dependency resource is a ReportInput
	// as this is (currently) the only type
	// when the is expanded, this code will beed to change

	var resource HclResource
	var found bool
	dependencyPath, err := ParseResourcePropertyPath(hclhelpers.TraversalAsString(d.Traversal))
	if err != nil {
		return err
	}
	resourceName := dependencyPath.ToResourceName()

	// if this dependency has a 'self' 'prefix, resolve from the current report container
	if dependencyPath.Scope == runtimeDependencyReportScope {
		resource, found = c.GetInput(resourceName)
	} else {
		// otherwise resolve from the workspace
		resource, found = workspace.GetResourceMaps().ReportInputs[resourceName]
	}
	if !found {
		return fmt.Errorf("could not resolve runtime depdency resource %s", dependencyPath)
	}

	d.SourceResource = resource
	return nil
}
