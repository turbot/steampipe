package db_client

import (
	"time"
)

// SessionStats is a struct uses to store initialisation status for database sessions
type SessionStats struct {
	Created     time.Time
	LastUsed    time.Time
	Initialized time.Time
	UsedCount   int
	SearchPath  []string `json:"-"`
	BackendPid  int64
}

func NewSessionStat() *SessionStats {
	t := time.Now()
	return &SessionStats{
		Created:     t,
		LastUsed:    t,
		Initialized: t,
		UsedCount:   0,
	}
}

func (s *SessionStats) UpdateUsage() {
	s.LastUsed = time.Now()
	s.UsedCount++
}
