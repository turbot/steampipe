package controlexecute

import "time"

type LifecycleEvent string

const (
	ControlRunLifecycleEventConstructed           LifecycleEvent = "constructed"
	ControlRunLifecycleEventExecuteStart          LifecycleEvent = "execute_start"
	ControlRunLifecycleEventQueuedForSession      LifecycleEvent = "queued_for_session"
	ControlRunLifecycleEventAcquiredSession       LifecycleEvent = "acquired_session"
	ControlRunLifecycleEventQueryResolutionStart  LifecycleEvent = "query_resolution_start"
	ControlRunLifecycleEventQueryResolutionFinish LifecycleEvent = "query_resolution_finish"
	ControlRunLifecycleEventSetSearchPathStart    LifecycleEvent = "set_search_path_start"
	ControlRunLifecycleEventSetSearchPathFinish   LifecycleEvent = "set_search_path_finish"
	ControlRunLifecycleEventQueryStart            LifecycleEvent = "query_start"
	ControlRunLifecycleEventQueryFinish           LifecycleEvent = "query_finish"
	ControlRunLifecycleEventGatherResultStart     LifecycleEvent = "gather_start"
	ControlRunLifecycleEventGatherResultFinish    LifecycleEvent = "gather_finish"
	ControlRunLifecycleEventExecuteFinish         LifecycleEvent = "execute_end"
)

type ControlRunLifecycle map[LifecycleEvent]time.Time

func newControlRunLifecycle() ControlRunLifecycle {
	return map[LifecycleEvent]time.Time{ControlRunLifecycleEventConstructed: time.Now()}
}

// GetDuration returns the duration between two events - if both exist
func (r ControlRunLifecycle) GetDuration(from, to LifecycleEvent) time.Duration {
	return r[from].Sub(r[to])
}

func (r ControlRunLifecycle) Constructed() {
	r[ControlRunLifecycleEventConstructed] = time.Now()
}
func (r ControlRunLifecycle) ExecuteStart() {
	r[ControlRunLifecycleEventExecuteStart] = time.Now()
}
func (r ControlRunLifecycle) QueuedForSession() {
	r[ControlRunLifecycleEventQueuedForSession] = time.Now()
}
func (r ControlRunLifecycle) AcquiredSession() {
	r[ControlRunLifecycleEventAcquiredSession] = time.Now()
}
func (r ControlRunLifecycle) QueryResolutionStart() {
	r[ControlRunLifecycleEventQueryResolutionStart] = time.Now()
}
func (r ControlRunLifecycle) QueryResolutionFinish() {
	r[ControlRunLifecycleEventQueryResolutionFinish] = time.Now()
}
func (r ControlRunLifecycle) SetSearchPathStart() {
	r[ControlRunLifecycleEventSetSearchPathStart] = time.Now()
}
func (r ControlRunLifecycle) SetSearchPathFinish() {
	r[ControlRunLifecycleEventSetSearchPathFinish] = time.Now()
}
func (r ControlRunLifecycle) QueryStart() {
	r[ControlRunLifecycleEventQueryStart] = time.Now()
}
func (r ControlRunLifecycle) QueryFinish() {
	r[ControlRunLifecycleEventQueryFinish] = time.Now()
}
func (r ControlRunLifecycle) GatherResultStart() {
	r[ControlRunLifecycleEventGatherResultStart] = time.Now()
}
func (r ControlRunLifecycle) GatherResultFinish() {
	r[ControlRunLifecycleEventGatherResultFinish] = time.Now()
}
func (r ControlRunLifecycle) ExecuteFinish() {
	r[ControlRunLifecycleEventExecuteFinish] = time.Now()
}
