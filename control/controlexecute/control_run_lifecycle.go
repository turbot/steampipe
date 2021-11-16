package controlexecute

import "time"

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

type ControlRunLifecycle map[ControlRunLifecycleEvent]time.Time

func newControlRunLifecycle() ControlRunLifecycle {
	return map[ControlRunLifecycleEvent]time.Time{ControlRunLifecycleEventConstructed: time.Now()}
}

// GetDuration returns the duration between two events - if both exist
func (r ControlRunLifecycle) GetDuration(from, to ControlRunLifecycleEvent) time.Duration {
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
