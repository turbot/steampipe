package dashboardexecute

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// LeafRun is a struct representing the execution of a leaf dashboard node
type LeafRun struct {
	Name string `json:"name"`

	Title               string                      `json:"title,omitempty"`
	Width               int                         `json:"width,omitempty"`
	SQL                 string                      `json:"sql,omitempty"`
	Data                *LeafData                   `json:"data,omitempty"`
	Error               error                       `json:"error,omitempty"`
	DashboardNode       modconfig.DashboardLeafNode `json:"properties"`
	NodeType            string                      `json:"node_type"`
	DashboardName       string                      `json:"dashboard"`
	Path                []string                    `json:"-"`
	parent              dashboardinterfaces.DashboardNodeParent
	runStatus           dashboardinterfaces.DashboardRunStatus
	executionTree       *DashboardExecutionTree
	runtimeDependencies map[string]*ResolvedRuntimeDependency
}

func NewLeafRun(resource modconfig.DashboardLeafNode, parent dashboardinterfaces.DashboardNodeParent, executionTree *DashboardExecutionTree) (*LeafRun, error) {
	// ensure the tree node name is unique
	name := executionTree.GetUniqueName(resource.Name())

	r := &LeafRun{
		Name:                name,
		Title:               resource.GetTitle(),
		Width:               resource.GetWidth(),
		Path:                resource.GetPaths()[0],
		DashboardNode:       resource,
		DashboardName:       executionTree.dashboardName,
		executionTree:       executionTree,
		parent:              parent,
		runtimeDependencies: make(map[string]*ResolvedRuntimeDependency),
		// set to complete, optimistically
		// if any children have SQL we will set this to DashboardRunReady instead
		runStatus: dashboardinterfaces.DashboardRunComplete,
	}

	parsedName, err := modconfig.ParseResourceName(resource.Name())
	if err != nil {
		return nil, err
	}
	r.NodeType = parsedName.ItemType
	// if we have a query provider, set status to ready
	if _, ok := resource.(modconfig.QueryProvider); ok {
		r.runStatus = dashboardinterfaces.DashboardRunReady
	}

	// if this node has runtime dependencies, create runtime depdency instances which we use to resolve the values
	// only QueryProvider resources support runtime dependencies
	queryProvider, ok := r.DashboardNode.(modconfig.QueryProvider)
	if ok {
		runtimeDependencies := queryProvider.GetRuntimeDependencies()
		for name, dep := range runtimeDependencies {
			r.runtimeDependencies[name] = NewResolvedRuntimeDependency(dep, executionTree)
		}

	}

	// add r into execution tree
	executionTree.runs[r.Name] = r
	return r, nil
}

// Execute implements DashboardRunNode
func (r *LeafRun) Execute(ctx context.Context) error {
	// if there is nothing to do, return
	if r.runStatus == dashboardinterfaces.DashboardRunComplete {
		return nil
	}

	// if there are any unresolved runtime dependencies, wait for them
	if err := r.waitForRuntimeDependencies(ctx); err != nil {
		return err
	}

	// ok now we have runtime depdencies, we can resolve the query
	queryProvider := r.DashboardNode.(modconfig.QueryProvider)
	sql, err := r.executionTree.workspace.ResolveQueryFromQueryProvider(queryProvider, nil)
	if err != nil {
		return err
	}
	r.SQL = sql

	queryResult, err := r.executionTree.client.ExecuteSync(ctx, r.SQL)
	if err != nil {
		// set the error status on the counter - this will raise counter error event
		r.SetError(err)
		return err

	}
	r.Data = NewLeafData(queryResult)
	// set complete status on counter - this will raise counter complete event
	r.SetComplete()
	return nil
}

// GetName implements DashboardNodeRun
func (r *LeafRun) GetName() string {
	return r.Name
}

// GetPath implements DashboardNodeRun
func (r *LeafRun) GetPath() modconfig.NodePath {
	return r.Path
}

// GetRunStatus implements DashboardNodeRun
func (r *LeafRun) GetRunStatus() dashboardinterfaces.DashboardRunStatus {
	return r.runStatus
}

// SetError implements DashboardNodeRun
func (r *LeafRun) SetError(err error) {
	r.Error = err
	r.runStatus = dashboardinterfaces.DashboardRunError
	// raise counter error event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.LeafNodeError{
		Node:    r,
		Session: r.executionTree.sessionId,
	})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r

}

// SetComplete implements DashboardNodeRun
func (r *LeafRun) SetComplete() {
	r.runStatus = dashboardinterfaces.DashboardRunComplete
	// raise counter complete event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.LeafNodeComplete{
		Node:    r,
		Session: r.executionTree.sessionId,
	})
	// tell parent we are done
	r.parent.ChildCompleteChan() <- r
}

// RunComplete implements DashboardNodeRun
func (r *LeafRun) RunComplete() bool {
	return r.runStatus == dashboardinterfaces.DashboardRunComplete || r.runStatus == dashboardinterfaces.DashboardRunError
}

// ChildrenComplete implements DashboardNodeRun
func (r *LeafRun) ChildrenComplete() bool {
	return true
}

func (r *LeafRun) waitForRuntimeDependencies(ctx context.Context) error {
	for _, resolvedDependency := range r.runtimeDependencies {
		// check with the top level dashboard whether the dependency is available
		if !resolvedDependency.Resolve() {
			if err := r.executionTree.waitForRuntimeDependency(ctx, resolvedDependency.dependency); err != nil {
				return err
			}
		}
		// now again resolve the dependency value - this sets the arg to have the runtime dependency value
		if !resolvedDependency.Resolve() {
			// should now be resolved`
			return fmt.Errorf("dependency not resolved after waitForRuntimeDependency returned")
		}
	}

	return nil
}
