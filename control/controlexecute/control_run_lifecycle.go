package controlexecute

import "time"

// lifecycle events during a control run
type ControlRunLifecycleEvent string

const (
	ControlRunLifecycleEventConstructed           ControlRunLifecycleEvent = "constructed"
	ControlRunLifecycleEventExecuteStart          ControlRunLifecycleEvent = "execute_start"
	ControlRunLifecycleEventQueuedForSession      ControlRunLifecycleEvent = "queued_for_session"
	ControlRunLifecycleEventAcquiredSession       ControlRunLifecycleEvent = "acquired_session"
	ControlRunLifecycleEventQueryResolutionStart  ControlRunLifecycleEvent = "query_resolution_start"
	ControlRunLifecycleEventQueryResolutionFinish ControlRunLifecycleEvent = "query_resolution_finish"
	ControlRunLifecycleEventSetSearchPathStart    ControlRunLifecycleEvent = "set_search_path_start"
	ControlRunLifecycleEventSetSearchPathFinish   ControlRunLifecycleEvent = "set_search_path_finish"
	ControlRunLifecycleEventQueryStart            ControlRunLifecycleEvent = "query_start"
	ControlRunLifecycleEventQueryFinish           ControlRunLifecycleEvent = "query_finish"
	ControlRunLifecycleEventGatherResultStart     ControlRunLifecycleEvent = "gather_start"
	ControlRunLifecycleEventGatherResultFinish    ControlRunLifecycleEvent = "gather_finish"
	ControlRunLifecycleEventExecuteFinish         ControlRunLifecycleEvent = "execute_end"
)

// records time.Time for lifecycle events
type ControlRunLifecycle map[ControlRunLifecycleEvent]time.Time

func newControlRunLifecycle() ControlRunLifecycle {
	return map[ControlRunLifecycleEvent]time.Time{ControlRunLifecycleEventConstructed: time.Now()}
}

// GetDuration returns the duration between two events - if both exist
func (r ControlRunLifecycle) GetDuration(from, to ControlRunLifecycleEvent) time.Duration {
	return r[from].Sub(r[to])
}

func (r ControlRunLifecycle) Add(event ControlRunLifecycleEvent) {
	r[event] = time.Now()
}
