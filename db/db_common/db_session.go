package db_common

import (
	"database/sql"
	"time"
)

type DBSessionTimeline struct {
	Created time.Time `json:"created"`

	QueuedForInitialize time.Time `json:"queued_for_init"`

	InitializeStart  time.Time `json:"init_start"`
	InitializeFinish time.Time `json:"init_finish"`

	IntrospectionTableStart  time.Time `json:"introspection_table_start"`
	IntrospectionTableFinish time.Time `json:"introspection_table_finish"`

	PreparedStatementStart  time.Time `json:"prepared_statement_start"`
	PreparedStatementFinish time.Time `json:"prepared_statement_finish"`

	LastUsed time.Time `json:"last_used"`
}

// DBSession wraps over the raw database/sql.Conn and also allows for retaining useful instrumentation
type DBSession struct {
	BackendPid  int64              `json: "backend_pid"`
	Timeline    *DBSessionTimeline `json:"timeline"`
	UsedCount   int                `json:"used"`
	SearchPath  []string           `json:"-"`
	Initialized bool               `json:"-"`

	// this gets rewritten, since the database/sql gives back a new instance everytime
	Raw *sql.Conn `json:"-"`
}

func NewDBSession(backendPid int64) *DBSession {
	return &DBSession{
		Timeline:   &DBSessionTimeline{Created: time.Now()},
		BackendPid: backendPid,
	}
}

func (s *DBSession) UpdateUsage() {
	s.UsedCount++
	s.Timeline.LastUsed = time.Now()
}

func (s *DBSession) GetRaw() *sql.Conn {
	return s.Raw
}

func (s *DBSession) Close() error {
	if s.Raw != nil {
		err := s.Raw.Close()
		s.Raw = nil
		return err
	}
	return nil
}
