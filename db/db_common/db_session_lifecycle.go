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
func (r DBSessionLifecycle) Add(event DBSessionLifecycleEvent) {
	r[event] = time.Now()
}
