package db_common

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

// DatabaseSession wraps over the raw database/sql.Conn and also allows for retaining useful instrumentation
type DatabaseSession struct {
	UsedCount   int       `json:"used"`
	LastUsed    time.Time `json:"last_used"`
	SearchPath  []string  `json:"-"`
	Initialized bool      `json:"-"`

	// this gets rewritten, since the database/sql gives back a new instance everytime
	Connection *pgxpool.Conn `json:"-"`

	// the id of the last scan metadata retrieved
	ScanMetadataMaxId int64 `json:"-"`
}

func NewDBSession() *DatabaseSession {
	return &DatabaseSession{}
}

// UpdateUsage updates the UsedCount of the DatabaseSession and also the lastUsed time
func (s *DatabaseSession) UpdateUsage() {
	s.UsedCount++
	s.LastUsed = time.Now()
}

func (s *DatabaseSession) Close(waitForCleanup bool) {
	if s.Connection != nil {
		if waitForCleanup {
			// TODO KAI what to do here???
			//s.Connection.Raw(func(driverConn interface{}) error {
			//	conn := driverConn.(*stdlib.Conn)
			//	select {
			//	case <-time.After(5 * time.Second):
			//		return context.DeadlineExceeded
			//	case <-conn.Conn().PgConn().CleanupDone():
			//		return nil
			//	}
			//})
		}
	}
	s.Connection.Release()
	s.Connection = nil

}
