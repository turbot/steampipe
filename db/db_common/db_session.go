package db_common

import (
	"database/sql"
)

// DBSession wraps over the raw database/sql.Conn and also allows for retaining useful instrumentation
type DBSession struct {
	BackendPid  int64              `json:"backend_pid"`
	Timeline    DBSessionLifecycle `json:"lifecycle"`
	UsedCount   int                `json:"used"`
	SearchPath  []string           `json:"-"`
	Initialized bool               `json:"-"`

	// this gets rewritten, since the database/sql gives back a new instance everytime
	Raw *sql.Conn `json:"-"`
}

func NewDBSession(backendPid int64) *DBSession {
	return &DBSession{
		Timeline:   DBSessionLifecycle{},
		BackendPid: backendPid,
	}
}

func (s *DBSession) UpdateUsage() {
	s.UsedCount++
	s.Timeline.LastUsed()
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
