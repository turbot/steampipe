package dashboardexecute

import (
	"context"
	"fmt"
	"log"

	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// LeafRun is a struct representing the execution of a leaf dashboard node
type LeafRun struct {
	Name string `json:"name"`

	Title               string `json:"title,omitempty"`
	Width               int    `json:"width,omitempty"`
	executeSQL          string
	RawSQL              string                      `json:"sql,omitempty"`
	Args                []string                    `json:"args,omitempty"`
	Data                *LeafData                   `json:"data,omitempty"`
	ErrorString         string                      `json:"error,omitempty"`
	DashboardNode       modconfig.DashboardLeafNode `json:"properties"`
	NodeType            string                      `json:"node_type"`
	DashboardName       string                      `json:"dashboard"`
	SourceDefinition    string                      `json:"source_definition"`
	error               error
	parent              dashboardinterfaces.DashboardNodeParent
	runStatus           dashboardinterfaces.DashboardRunStatus
	executionTree       *DashboardExecutionTree
	runtimeDependencies map[string]*ResolvedRuntimeDependency
}

func NewLeafRun(resource modconfig.DashboardLeafNode, parent dashboardinterfaces.DashboardNodeParent, executionTree *DashboardExecutionTree) (*LeafRun, error) {
	// NOTE: for now we MUST declare container/dashboard children inline - therefore we cannot share children between runs in the tree
	// (if we supported the children property then we could reuse resources)
	// so FOR NOW it is safe to use the node name directly as the run name
	name := resource.Name()

	r := &LeafRun{
		Name:                name,
		Title:               resource.GetTitle(),
		Width:               resource.GetWidth(),
		DashboardNode:       resource,
		DashboardName:       executionTree.dashboardName,
		SourceDefinition:    resource.GetMetadata().SourceDefinition,
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
	// if we have a query provider which requires execution, set status to ready
	if provider, ok := resource.(modconfig.QueryProvider); ok && provider.RequiresExecution(provider) {
		// if the provider has sql or a query, set status to ready
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

		// if the node has no runtime dependencies, resolve the sql
		if len(r.runtimeDependencies) == 0 {
			if err := r.resolveSQL(); err != nil {
				return nil, err
			}
		}

	}

	// add r into execution tree
	executionTree.runs[r.Name] = r
	return r, nil
}

// Execute implements DashboardRunNode
func (r *LeafRun) Execute(ctx context.Context) {
	// if there is nothing to do, return
	if r.runStatus == dashboardinterfaces.DashboardRunComplete {
		return
	}

	log.Printf("[TRACE] LeafRun '%s' Execute()", r.DashboardNode.Name())

	// to get here, we must be a query provider

	// if there are any unresolved runtime dependencies, wait for them
	if len(r.runtimeDependencies) > 0 {
		if err := r.waitForRuntimeDependencies(ctx); err != nil {
			r.SetError(err)
			return
		}

		// ok now we have runtime dependencies, we can resolve the query
		if err := r.resolveSQL(); err != nil {
			r.SetError(err)
			return
		}
	}

	log.Printf("[TRACE] LeafRun '%s' SQL resolved, executing", r.DashboardNode.Name())

	queryResult, err := r.executionTree.client.ExecuteSync(ctx, r.executeSQL)
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' query failed: %s", r.DashboardNode.Name(), err.Error())
		// set the error status on the counter - this will raise counter error event
		r.SetError(err)
		return

	}
	log.Printf("[TRACE] LeafRun '%s' complete", r.DashboardNode.Name())

	r.Data = NewLeafData(queryResult)
	// set complete status on counter - this will raise counter complete event
	r.SetComplete()
}

// GetName implements DashboardNodeRun
func (r *LeafRun) GetName() string {
	return r.Name
}

// GetRunStatus implements DashboardNodeRun
func (r *LeafRun) GetRunStatus() dashboardinterfaces.DashboardRunStatus {
	return r.runStatus
}

// SetError implements DashboardNodeRun
func (r *LeafRun) SetError(err error) {
	r.error = err
	// error type does not serialise to JSON so copy into a string
	r.ErrorString = err.Error()

	r.runStatus = dashboardinterfaces.DashboardRunError
	// raise counter error event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.LeafNodeError{
		LeafNode:    r,
		Session:     r.executionTree.sessionId,
		ExecutionId: r.executionTree.id,
	})
	r.parent.ChildCompleteChan() <- r
}

// GetError implements DashboardNodeRun
func (r *LeafRun) GetError() error {
	return r.error
}

// SetComplete implements DashboardNodeRun
func (r *LeafRun) SetComplete() {
	r.runStatus = dashboardinterfaces.DashboardRunComplete
	// raise counter complete event
	r.executionTree.workspace.PublishDashboardEvent(&dashboardevents.LeafNodeComplete{
		LeafNode:    r,
		Session:     r.executionTree.sessionId,
		ExecutionId: r.executionTree.id,
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

// GetInputsDependingOn implements DashboardNodeRun
//return nothing for LeafRun
func (r *LeafRun) GetInputsDependingOn(changedInputName string) []string { return nil }

func (r *LeafRun) waitForRuntimeDependencies(ctx context.Context) error {
	log.Printf("[TRACE] LeafRun '%s' waitForRuntimeDependencies", r.DashboardNode.Name())
	for _, resolvedDependency := range r.runtimeDependencies {
		// check with the top level dashboard whether the dependency is available
		if !resolvedDependency.Resolve() {
			log.Printf("[TRACE] waitForRuntimeDependency %s", resolvedDependency.dependency.String())
			if err := r.executionTree.waitForRuntimeDependency(ctx, resolvedDependency.dependency); err != nil {
				return err
			}
		}

		log.Printf("[TRACE] dependency %s should be available", resolvedDependency.dependency.String())
		// now again resolve the dependency value - this sets the arg to have the runtime dependency value
		if !resolvedDependency.Resolve() {
			log.Printf("[TRACE] dependency %s not resolved after waitForRuntimeDependency returned", resolvedDependency.dependency.String())
			// should now be resolved`
			return fmt.Errorf("dependency %s not resolved after waitForRuntimeDependency returned", resolvedDependency.dependency.String())
		}
	}

	if len(r.runtimeDependencies) > 0 {
		log.Printf("[TRACE] LeafRun '%s' all runtime dependencies ready", r.DashboardNode.Name())
	}
	return nil
}

// resolve the sql for this leaf run into the source sql (i.e. NOT the prepared statement name) and resolved args
func (r *LeafRun) resolveSQL() error {
	log.Printf("[TRACE] LeafRun '%s' resolveSQL", r.DashboardNode.Name())
	queryProvider := r.DashboardNode.(modconfig.QueryProvider)
	if !queryProvider.RequiresExecution(queryProvider) {
		log.Printf("[TRACE] LeafRun '%s'does NOT require execution - returning", r.DashboardNode.Name())
		return nil
	}
	err := queryProvider.VerifyQuery(queryProvider)
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' VerifyQuery failed: %s", r.DashboardNode.Name(), err.Error())
		return err
	}

	// convert runtime dependencies into arg map
	runtimeArgs, err := r.buildRuntimeDependencyArgs()
	if err != nil {
		log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs failed: %s", r.DashboardNode.Name(), err.Error())
		return err
	}

	log.Printf("[TRACE] LeafRun '%s' built runtime args: %v", r.DashboardNode.Name(), runtimeArgs)

	resolvedQuery, err := r.executionTree.workspace.ResolveQueryFromQueryProvider(queryProvider, runtimeArgs)
	if err != nil {
		return err
	}
	r.RawSQL = resolvedQuery.RawSQL
	r.executeSQL = resolvedQuery.ExecuteSQL
	r.Args = resolvedQuery.Args
	return nil
}

func (r *LeafRun) buildRuntimeDependencyArgs() (*modconfig.QueryArgs, error) {
	res := modconfig.NewQueryArgs()

	log.Printf("[TRACE] LeafRun '%s' buildRuntimeDependencyArgs - %d runtime dependencies", r.DashboardNode.Name(), len(r.runtimeDependencies))

	// if the runtime dependencies use position args, get the max index and ensure the args array is large enough
	maxArgIndex := -1
	for _, dep := range r.runtimeDependencies {
		if dep.dependency.ArgIndex != nil && *dep.dependency.ArgIndex > maxArgIndex {
			maxArgIndex = *dep.dependency.ArgIndex
		}
	}
	if maxArgIndex != -1 {
		res.ArgList = make([]*string, maxArgIndex+1)
	}

	// build map of default params
	for _, dep := range r.runtimeDependencies {
		// format the arg value as a postgres string (this will also work for numbers)
		formattedVal := pgEscapeParamString(fmt.Sprintf("%v", dep.value))
		if dep.dependency.ArgName != nil {
			res.ArgMap[*dep.dependency.ArgName] = formattedVal
		} else {
			if dep.dependency.ArgIndex == nil {
				return nil, fmt.Errorf("invalid runtime dependency - both ArgName and ArgIndex are nil ")
			}

			// now add at correct index
			res.ArgList[*dep.dependency.ArgIndex] = &formattedVal
		}
	}
	return res, nil
}

// format a string for use as a postgres string param
func pgEscapeParamString(val string) string {
	return fmt.Sprintf("'%s'", val)
}
