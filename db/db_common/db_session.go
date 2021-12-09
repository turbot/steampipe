package db_common

import (
	"database/sql"
	"time"

	"github.com/turbot/steampipe/utils"
)

// DatabaseSession wraps over the raw database/sql.Conn and also allows for retaining useful instrumentation
type DatabaseSession struct {
	BackendPid  uint32                `json:"backend_pid"`
	LifeCycle   *utils.LifecycleTimer `json:"lifecycle"`
	UsedCount   int                   `json:"used"`
	LastUsed    time.Time             `json:"last_used"`
	SearchPath  []string              `json:"-"`
	Initialized bool                  `json:"-"`

	// this gets rewritten, since the database/sql gives back a new instance everytime
	Connection *sql.Conn `json:"-"`
}

func NewDBSession(backendPid uint32) *DatabaseSession {
	return &DatabaseSession{
		LifeCycle:  utils.NewLifecycleTimer(),
		BackendPid: backendPid,
	}
}

// UpdateUsage updates the UsedCount of the DatabaseSession and also the lastUsed time
func (s *DatabaseSession) UpdateUsage() {
	s.UsedCount++
	s.LastUsed = time.Now()
}

func (s *DatabaseSession) Close() error {
	if s.Connection != nil {
		err := s.Connection.Close()
		s.Connection = nil
		return err
	}
	return nil
}
