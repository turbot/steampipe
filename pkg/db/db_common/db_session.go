package db_common

import (
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseSession wraps over the raw database/sql.Conn and also allows for retaining useful instrumentation
type DatabaseSession struct {
	BackendPid  uint32    `json:"backend_pid"`
	UsedCount   int       `json:"used"`
	LastUsed    time.Time `json:"last_used"`
	SearchPath  []string  `json:"-"`
	Initialized bool      `json:"-"`

	// this gets rewritten, since the database/sql gives back a new instance everytime
	Connection *pgxpool.Conn `json:"-"`

	// the id of the last scan metadata retrieved
	ScanMetadataMaxId int64 `json:"-"`
}

func NewDBSession(backendPid uint32) *DatabaseSession {
	return &DatabaseSession{
		BackendPid: backendPid,
	}
}

// UpdateUsage updates the UsedCount of the DatabaseSession and also the lastUsed time
func (s *DatabaseSession) UpdateUsage() {
	s.UsedCount++
	s.LastUsed = time.Now()
}

func (s *DatabaseSession) Close(waitForCleanup bool) {
	if s.Connection != nil {
		if waitForCleanup {
			log.Printf("[TRACE] DatabaseSession.Close wait for connection cleanup")
			select {
			case <-time.After(5 * time.Second):
				log.Printf("[TRACE] DatabaseSession.Close timed out waiting for connection cleanup")
			case <-s.Connection.Conn().PgConn().CleanupDone():
				log.Printf("[TRACE] DatabaseSession.Close connection cleanup complete")
			}
		}
		s.Connection.Release()
	}
	s.Connection = nil

}
