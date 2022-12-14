package dashboardexecute

import (
	"context"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"log"
)

type DashboardParentImpl struct {
	DashboardTreeRunImpl
	children      []dashboardtypes.DashboardTreeRun
	childComplete chan dashboardtypes.DashboardTreeRun
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
			return false
		}
	}

	return true
}

func (r *DashboardParentImpl) ChildCompleteChan() chan dashboardtypes.DashboardTreeRun {
	return r.childComplete
}

func (r *DashboardParentImpl) createChildCompleteChan() {
	// create buffered child complete chan
	if childCount := len(r.children); childCount > 0 {
		r.childComplete = make(chan dashboardtypes.DashboardTreeRun, childCount)
	}
}

// if this leaf run has children (including with runs) execute them asynchronously
func (r *DashboardParentImpl) executeChildrenAsync(ctx context.Context) {
	for _, c := range r.children {
		go c.Execute(ctx)
	}
}

func (r *DashboardParentImpl) waitForChildren() chan error {
	log.Printf("[TRACE] %s waitForChildren", r.Name)
	var doneChan = make(chan error)
	if len(r.children) == 0 {
		log.Printf("[TRACE] %s waitForChildren - no children so we're done", r.Name)
		// if there are no children, return a closed channel so we do not wait
		close(doneChan)
	} else {
		go func() {
			// wait for children to complete
			var errors []error

			for !(r.ChildrenComplete()) {
				completeChild := <-r.childComplete
				log.Printf("[TRACE] %s got child complete for %s", r.Name, completeChild.GetName())
				if completeChild.GetRunStatus() == dashboardtypes.DashboardRunError {
					errors = append(errors, completeChild.GetError())
					log.Printf("[TRACE] %s child %s has error %v", r.Name, completeChild.GetName(), completeChild.GetError())
				}
				// fall through to recheck ChildrenComplete
			}

			log.Printf("[TRACE]  %s ALL children and withs complete, errors: %v", r.Name, errors)
			// so all children have completed - check for errors
			// TODO [node_reuse] format better error
			doneChan <- error_helpers.CombineErrors(errors...)
		}()
	}
	return doneChan
}
