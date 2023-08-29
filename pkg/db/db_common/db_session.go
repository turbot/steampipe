package db_common

import (
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

// Releasable is the bare minimum set of functions that DatabaseSession needs from it's underlying
// connection.
type Releasable interface {
	Release()
	Conn() *pgx.Conn
}

type releasableConn struct {
	conn *pgx.Conn
}

func (r *releasableConn) Release()        {}
func (r *releasableConn) Conn() *pgx.Conn { return r.conn }

func ReleasableFromConn(conn *pgx.Conn) Releasable {
	return &releasableConn{conn: conn}
}

// DatabaseSession wraps over the raw database connection
// the purpose is to be able
//   - to store the current search path of the connection without having to make a database round-trip
//   - To store the last scan_metadata id used on this connection
type DatabaseSession struct {
	BackendPid uint32   `json:"backend_pid"`
	SearchPath []string `json:"-"`

	// this gets rewritten, since the database/sql gives back a new instance everytime
	Connection Releasable `json:"-"`

	// the id of the last scan metadata retrieved
	ScanMetadataMaxId int64 `json:"-"`
}

func NewDBSession(backendPid uint32) *DatabaseSession {
	return &DatabaseSession{
		BackendPid: backendPid,
	}
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
