package dashboardexecute

import (
	"context"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"log"
	"sync"
)

type DashboardParentImpl struct {
	DashboardTreeRunImpl
	children          []dashboardtypes.DashboardTreeRun
	childCompleteChan chan dashboardtypes.DashboardTreeRun
	// are we blocked by a child run
	blockedByChild  bool
	childStatusLock *sync.Mutex
}

func newDashboardParentImpl(resource modconfig.DashboardLeafNode, parent dashboardtypes.DashboardParent, run dashboardtypes.DashboardTreeRun, executionTree *DashboardExecutionTree) DashboardParentImpl {
	return DashboardParentImpl{
		DashboardTreeRunImpl: NewDashboardTreeRunImpl(resource, parent, run, executionTree),
		childStatusLock:      new(sync.Mutex),
	}
}

func (r *DashboardParentImpl) initialiseChildren(ctx context.Context) error {
	var errors []error
	for _, child := range r.children {
		child.Initialise(ctx)

		if err := child.GetError(); err != nil {
			errors = append(errors, err)
		}
	}

	return error_helpers.CombineErrors(errors...)

}

// GetChildren implements DashboardTreeRun
func (r *DashboardParentImpl) GetChildren() []dashboardtypes.DashboardTreeRun {
	return r.children
}

// ChildrenComplete implements DashboardTreeRun
func (r *DashboardParentImpl) ChildrenComplete() bool {
	for _, child := range r.children {
		if !child.RunComplete() {
			log.Printf("[TRACE] %s ChildrenComplete child %s NOT complete state %s", r.Name, child.GetName(), child.GetRunStatus())
			return false
		}
	}

	return true
}

func (r *DashboardParentImpl) ChildCompleteChan() chan dashboardtypes.DashboardTreeRun {
	return r.childCompleteChan
}
func (r *DashboardParentImpl) createChildCompleteChan() {
	// create buffered child complete chan
	if childCount := len(r.children); childCount > 0 {
		r.childCompleteChan = make(chan dashboardtypes.DashboardTreeRun, childCount)
	}
}

// if this leaf run has children (including with runs) execute them asynchronously
func (r *DashboardParentImpl) executeChildrenAsync(ctx context.Context) {
	for _, c := range r.children {
		go c.Execute(ctx)
	}
}

// if this leaf run has with runs execute them asynchronously
func (r *DashboardParentImpl) executeWithsAsync(ctx context.Context) {
	for _, c := range r.children {
		if c.GetNodeType() == modconfig.BlockTypeWith {
			go c.Execute(ctx)
		}
	}
}

func (r *DashboardParentImpl) waitForChildrenAsync(ctx context.Context) chan error {
	log.Printf("[TRACE] %s waitForChildrenAsync", r.Name)
	var doneChan = make(chan error)
	if len(r.children) == 0 {
		log.Printf("[TRACE] %s waitForChildrenAsync - no children so we're done", r.Name)
		// if there are no children, return a closed channel so we do not wait
		close(doneChan)
		return doneChan
	}

	go func() {
		// wait for children to complete
		var errors []error
		for !(r.ChildrenComplete()) {
			completeChild := <-r.childCompleteChan
			log.Printf("[TRACE] %s waitForChildrenAsync got child complete for %s", r.Name, completeChild.GetName())
			if completeChild.GetRunStatus().IsError() {
				errors = append(errors, completeChild.GetError())
				log.Printf("[TRACE] %s child %s has error %v", r.Name, completeChild.GetName(), completeChild.GetError())
			}
		}

		log.Printf("[TRACE] %s ALL children and withs complete, errors: %v", r.Name, errors)

		// so all children have completed - check for errors
		// TODO [node_reuse] format better error https://github.com/turbot/steampipe/issues/2920
		err := error_helpers.CombineErrors(errors...)

		// if context is cancelled, just return context cancellation error
		if ctx.Err() != nil {
			err = ctx.Err()
		}

		doneChan <- err
	}()

	return doneChan
}

func (r *DashboardParentImpl) ChildStatusChanged(ctx context.Context) {
	// this function may be called asyncronously by children
	r.childStatusLock.Lock()
	defer r.childStatusLock.Unlock()

	// if we are currently blocked by a child or we are currently in running state,
	// call setRunning() to determine whether any of our children are now blocked
	if r.blockedByChild || r.GetRunStatus() == dashboardtypes.RunRunning {
		log.Printf("[TRACE] %s ChildStatusChanged - calling setRunning to see if we are still running, status %s blockedByChild %v", r.Name, r.GetRunStatus(), r.blockedByChild)

		// try setting our status to running again
		r.setRunning(ctx)
	}
}

// override DashboardTreeRunImpl) setStatus(
func (r *DashboardParentImpl) setRunning(ctx context.Context) {
	// if the run is already complete (for example, canceled), do nothing
	if r.GetRunStatus().IsFinished() {
		log.Printf("[TRACE] %s setRunning - run already terminated - current state %s - NOT setting running", r.Name, r.GetRunStatus())
		return
	}

	status := dashboardtypes.RunRunning
	// if we are trying to set status to running, check if any of our children are blocked,
	// and if so set our status to blocked

	// if any children are blocked, we are blocked
	for _, c := range r.children {
		if c.GetRunStatus() == dashboardtypes.RunBlocked {
			status = dashboardtypes.RunBlocked
			r.blockedByChild = true
			break
		}
		// to get here, no children can be blocked - clear blockedByChild
		r.blockedByChild = false
	}

	// set status if it has changed
	if status != r.GetRunStatus() {
		log.Printf("[TRACE] %s setRunning - setting state %s, blockedByChild %v", r.Name, status, r.blockedByChild)
		r.DashboardTreeRunImpl.setStatus(ctx, status)
	} else {
		log.Printf("[TRACE] %s setRunning - state unchanged %s, blockedByChild %v", r.Name, status, r.blockedByChild)
	}
}
