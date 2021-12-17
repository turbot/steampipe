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

// GetDuration returns the duration between first and the last event
func (r LifecycleTimer) GetDuration() time.Duration {
	lastEvent := r.events[len(r.events)-1]
	firstEvent := r.events[0]
	return lastEvent.Time.Sub(firstEvent.Time)
}

func (r *LifecycleTimer) Add(event string) {
	r.events = append(r.events, &LifecycleEvent{event, time.Now()})
}
