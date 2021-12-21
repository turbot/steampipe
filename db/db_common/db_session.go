package db_common

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/stdlib"
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

func (s *DatabaseSession) Close(waitForCleanup bool) error {
	var err error
	if s.Connection != nil {
		if waitForCleanup {
			s.Connection.Raw(func(driverConn interface{}) error {
				conn := driverConn.(*stdlib.Conn)
				select {
				case <-time.After(5 * time.Second):
					return context.DeadlineExceeded
				case <-conn.Conn().PgConn().CleanupDone():
					return nil
				}
			})
		}

		err = s.Connection.Close()
		s.Connection = nil
		fmt.Println(err)
		return err
	}
	return nil
}
