package dashboardexecute

import (
	"context"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"log"
)

type DashboardTreeRunImpl struct {
	DashboardName    string                   `json:"dashboard"`
	Description      string                   `json:"description,omitempty"`
	Display          string                   `cty:"display" hcl:"display" json:"display,omitempty"`
	Documentation    string                   `json:"documentation,omitempty"`
	ErrorString      string                   `json:"error,omitempty"`
	Name             string                   `json:"name"`
	NodeType         string                   `json:"panel_type"`
	SourceDefinition string                   `json:"source_definition"`
	Status           dashboardtypes.RunStatus `json:"status"`
	Tags             map[string]string        `json:"tags,omitempty"`
	Title            string                   `json:"title,omitempty"`
	Type             string                   `json:"display_type,omitempty"`
	Width            int                      `json:"width,omitempty"`

	err           error
	parent        dashboardtypes.DashboardParent
	executionTree *DashboardExecutionTree
	resource      modconfig.DashboardLeafNode
	// store the top level run which embeds this struct
	// we need this for setStatus which serialises the run for the message payload
	run dashboardtypes.DashboardTreeRun
}

func NewDashboardTreeRunImpl(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, run dashboardtypes.DashboardTreeRun, executionTree *DashboardExecutionTree) DashboardTreeRunImpl {
	// NOTE: we MUST declare children inline - therefore we cannot share children between runs in the tree
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
		Status:        dashboardtypes.RunComplete,
		parent:        parent,
		executionTree: executionTree,
		resource:      resource,
		run:           run,
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
func (r *DashboardTreeRunImpl) GetRunStatus() dashboardtypes.RunStatus {
	return r.Status
}

// GetError implements DashboardTreeRun
func (r *DashboardTreeRunImpl) GetError() error {
	return r.err
}

// RunComplete implements DashboardTreeRun
func (r *DashboardTreeRunImpl) RunComplete() bool {
	return r.Status.IsFinished()
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

// Initialise implements DashboardTreeRun
func (r *DashboardTreeRunImpl) Initialise(context.Context) {
	panic("must be implemented by child struct")
}

// Execute implements DashboardTreeRun
func (r *DashboardTreeRunImpl) Execute(ctx context.Context) {
	panic("must be implemented by child struct")
}

// AsTreeNode implements DashboardTreeRun
func (r *DashboardTreeRunImpl) AsTreeNode() *dashboardtypes.SnapshotTreeNode {
	panic("must be implemented by child struct")
}

// GetResource implements DashboardTreeRun
func (r *DashboardTreeRunImpl) GetResource() modconfig.DashboardLeafNode {
	return r.resource
}

// SetError implements DashboardTreeRun
func (r *DashboardTreeRunImpl) SetError(ctx context.Context, err error) {
	log.Printf("[TRACE] %s SetError err %v", r.Name, err)
	r.err = err
	// error type does not serialise to JSON so copy into a string
	r.ErrorString = err.Error()
	// set status (this sends update event)
	if error_helpers.IsContextCancelledError(err) {
		r.setStatus(ctx, dashboardtypes.RunCanceled)
	} else {
		r.setStatus(ctx, dashboardtypes.RunError)
	}
	// tell parent we are done
	r.notifyParentOfCompletion()
}

// SetComplete implements DashboardTreeRun
func (r *DashboardTreeRunImpl) SetComplete(ctx context.Context) {
	// set status (this sends update event)
	r.setStatus(ctx, dashboardtypes.RunComplete)
	// tell parent we are done
	r.notifyParentOfCompletion()
}

func (r *DashboardTreeRunImpl) setStatus(ctx context.Context, status dashboardtypes.RunStatus) {
	r.Status = status
	// notify our parent that our status has changed
	r.parent.ChildStatusChanged(ctx)

	// raise LeafNodeUpdated event
	// TODO [node_reuse] do this a different way https://github.com/turbot/steampipe/issues/2919
	// TACTICAL: pass the full run struct - 'r.run', rather than ourselves - so we serialize all properties
	e, _ := dashboardevents.NewLeafNodeUpdate(r.run, r.executionTree.sessionId, r.executionTree.id)
	r.executionTree.workspace.PublishDashboardEvent(ctx, e)

}

func (r *DashboardTreeRunImpl) notifyParentOfCompletion() {
	r.parent.ChildCompleteChan() <- r
}
