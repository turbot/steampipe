package dashboardexecute

import (
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

type DashboardTreeRunImpl struct {
	DashboardName    string                            `json:"dashboard"`
	Description      string                            `json:"description,omitempty"`
	Display          string                            `cty:"display" hcl:"display" json:"display,omitempty"`
	Documentation    string                            `json:"documentation,omitempty"`
	ErrorString      string                            `json:"error,omitempty"`
	Name             string                            `json:"name"`
	NodeType         string                            `json:"panel_type"`
	SourceDefinition string                            `json:"source_definition"`
	Status           dashboardtypes.DashboardRunStatus `json:"status"`
	Tags             map[string]string                 `json:"tags,omitempty"`
	Title            string                            `json:"title,omitempty"`
	Type             string                            `json:"display_type,omitempty"`
	Width            int                               `json:"width,omitempty"`

	err           error
	parent        dashboardtypes.DashboardParent
	executionTree *DashboardExecutionTree
	resource      modconfig.HclResource
}

func NewDashboardTreeRunImpl(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, executionTree *DashboardExecutionTree) DashboardTreeRunImpl {
	// NOTE: for now we MUST declare children inline - therefore we cannot share children between runs in the tree
	// (if we supported the children property then we could reuse resources)
	// so FOR NOW it is safe to use the container name directly as the run name
	res := DashboardTreeRunImpl{
		Name:             resource.Name(),
		Title:            resource.GetTitle(),
		NodeType:         resource.BlockType(),
		Width:            resource.GetWidth(),
		Display:          resource.GetDisplay(),
		Description:      resource.GetDescription(),
		Documentation:    resource.GetDocumentation(),
		Type:             resource.GetType(),
		Tags:             resource.GetTags(),
		SourceDefinition: resource.GetMetadata().SourceDefinition,

		// set to complete, optimistically
		// if any children have SQL we will set this to DashboardRunReady instead
		Status:        dashboardtypes.DashboardRunComplete,
		parent:        parent,
		executionTree: executionTree,
		resource:      resource,
	}

	// TACTICAL if this run was created to create a snapshot output for a control run,
	// there will be no execution tree
	if executionTree != nil {
		res.DashboardName = executionTree.dashboardName
	} else {
		// there is no execution tree - use the resource name as the dashboard name
		res.DashboardName = resource.Name()
	}
	return res
}

// GetName implements DashboardTreeRun
func (r *DashboardTreeRunImpl) GetName() string {
	return r.Name
}

// GetRunStatus implements DashboardTreeRun
func (r *DashboardTreeRunImpl) GetRunStatus() dashboardtypes.DashboardRunStatus {
	return r.Status
}

// GetError implements DashboardTreeRun
func (r *DashboardTreeRunImpl) GetError() error {
	return r.err
}

// RunComplete implements DashboardTreeRun
func (r *DashboardTreeRunImpl) RunComplete() bool {
	return r.Status == dashboardtypes.DashboardRunComplete || r.Status == dashboardtypes.DashboardRunError
}

// GetInputsDependingOn implements DashboardTreeRun
// defaults to nothing
func (r *DashboardTreeRunImpl) GetInputsDependingOn(_ string) []string { return nil }

// GetParent implements DashboardTreeRun
func (r *DashboardTreeRunImpl) GetParent() dashboardtypes.DashboardParent {
	return r.parent
}

// GetTitle implements DashboardTreeRun
func (r *DashboardTreeRunImpl) GetTitle() string {
	return r.Title
}

// GetNodeType implements DashboardTreeRun
func (r *DashboardTreeRunImpl) GetNodeType() string {
	return r.NodeType
}
