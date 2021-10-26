package db_client

import "time"

type SessionStats struct {
	Created     time.Time
	LastUsed    time.Time
	Initialized time.Time
	Waits       []time.Duration
	UsedCount   int
	SearchPath  []string
}

func NewSessionStat() SessionStats {
	t := time.Now()
	return SessionStats{
		Created:     t,
		LastUsed:    t,
		Initialized: t,
		UsedCount:   0,
	}
}
