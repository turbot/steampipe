package utils

import "time"

type LifecycleEvent struct {
	Event string
	Time  time.Time
}

// LifecycleTimer records the time for lifecycle events
type LifecycleTimer struct {
	events []*LifecycleEvent
}

func NewLifecycleTimer() *LifecycleTimer {
	return &LifecycleTimer{}
}

// GetDuration returns the duration between two events - if both exist
func (r LifecycleTimer) GetDuration() time.Duration {

	return r.events[0].Time.Sub(r.events[len(r.events)-1].Time)
}

func (r *LifecycleTimer) Add(event string) {
	r.events = append(r.events, &LifecycleEvent{event, time.Now()})
}
