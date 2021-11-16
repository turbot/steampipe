package db_common

import "time"

type DBSessionLifecycleEvent string

const (
	DBSessionLifecycleEventCreated                  DBSessionLifecycleEvent = "created"
	DBSessionLifecycleEventQueuedForInitialize      DBSessionLifecycleEvent = "queued_for_init"
	DBSessionLifecycleEventInitializeStart          DBSessionLifecycleEvent = "init_start"
	DBSessionLifecycleEventInitializeFinish         DBSessionLifecycleEvent = "init_finish"
	DBSessionLifecycleEventIntrospectionTableStart  DBSessionLifecycleEvent = "introspection_table_start"
	DBSessionLifecycleEventIntrospectionTableFinish DBSessionLifecycleEvent = "introspection_table_finish"
	DBSessionLifecycleEventPreparedStatementStart   DBSessionLifecycleEvent = "prepared_statement_start"
	DBSessionLifecycleEventPreparedStatementFinish  DBSessionLifecycleEvent = "prepared_statement_finish"
	DBSessionLifecycleEventLastUsed                 DBSessionLifecycleEvent = "last_used"
)

type DBSessionLifecycle map[DBSessionLifecycleEvent]time.Time

// GetDuration returns the duration between two events - if both exist
func (r DBSessionLifecycle) GetDuration(from, to DBSessionLifecycleEvent) time.Duration {
	return r[from].Sub(r[to])
}

func (r DBSessionLifecycle) Created() {
	r[DBSessionLifecycleEventCreated] = time.Now()
}
func (r DBSessionLifecycle) QueuedForInitialize() {
	r[DBSessionLifecycleEventQueuedForInitialize] = time.Now()
}
func (r DBSessionLifecycle) InitializeStart() {
	r[DBSessionLifecycleEventInitializeStart] = time.Now()
}
func (r DBSessionLifecycle) InitializeFinish() {
	r[DBSessionLifecycleEventInitializeFinish] = time.Now()
}
func (r DBSessionLifecycle) IntrospectionTableStart() {
	r[DBSessionLifecycleEventIntrospectionTableStart] = time.Now()
}
func (r DBSessionLifecycle) IntrospectionTableFinish() {
	r[DBSessionLifecycleEventIntrospectionTableFinish] = time.Now()
}
func (r DBSessionLifecycle) PreparedStatementStart() {
	r[DBSessionLifecycleEventPreparedStatementStart] = time.Now()
}
func (r DBSessionLifecycle) PreparedStatementFinish() {
	r[DBSessionLifecycleEventPreparedStatementFinish] = time.Now()
}
func (r DBSessionLifecycle) LastUsed() {
	r[DBSessionLifecycleEventLastUsed] = time.Now()
}
