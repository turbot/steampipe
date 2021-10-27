package db_client

import "time"

type SessionStats struct {
	Created     time.Time
	LastUsed    time.Time
	Initialized time.Time
	UsedCount   int
	Waits       []time.Duration `json:"-"`
	SearchPath  []string        `json:"-"`
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
