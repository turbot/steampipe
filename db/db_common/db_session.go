package db_common

import (
	"database/sql"
)

// DatabaseSession wraps over the raw database/sql.Conn and also allows for retaining useful instrumentation
type DatabaseSession struct {
	BackendPid  int64              `json:"backend_pid"`
	Timeline    DBSessionLifecycle `json:"lifecycle"`
	UsedCount   int                `json:"used"`
	SearchPath  []string           `json:"-"`
	Initialized bool               `json:"-"`

	// this gets rewritten, since the database/sql gives back a new instance everytime
	Connection *sql.Conn `json:"-"`
}

func NewDBSession(backendPid int64) *DatabaseSession {
	return &DatabaseSession{
		Timeline:   DBSessionLifecycle{},
		BackendPid: backendPid,
	}
}

// UpdateUsage updates the UsedCount of the DatabaseSession and also the lastUsed time
func (s *DatabaseSession) UpdateUsage() {
	s.UsedCount++
	s.Timeline.Add(DBSessionLifecycleEventLastUsed)
}

func (s *DatabaseSession) Close() error {
	if s.Connection != nil {
		err := s.Connection.Close()
		s.Connection = nil
		return err
	}
	return nil
}
